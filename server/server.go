package input

import (
	"bufio"
	"crypto/tls"
	"encoding/binary"
	"io"
	"net"
	"runtime"
	"sync"
	"sync/atomic"
	"time"

	"github.com/negbie/heplify-server"
	"github.com/negbie/heplify-server/config"
	"github.com/negbie/heplify-server/database"
	"github.com/negbie/heplify-server/elastic"
	"github.com/negbie/heplify-server/metric"
	"github.com/negbie/heplify-server/queue"
	"github.com/negbie/logp"
)

type HEPInput struct {
	useDB  bool
	useMQ  bool
	usePM  bool
	useES  bool
	buffer *sync.Pool
	quit   chan struct{}
	stats  HEPStats
	wg     *sync.WaitGroup
}

type HEPStats struct {
	DupCount uint64
	ErrCount uint64
	HEPCount uint64
	PktCount uint64
}

var (
	inCh = make(chan []byte, 20000)
	dbCh = make(chan *decoder.HEP, 200000)
	mqCh = make(chan []byte, 20000)
	pmCh = make(chan *decoder.HEP, 20000)
	esCh = make(chan *decoder.HEP, 20000)
)

const maxPktLen = 8192

func NewHEPInput() *HEPInput {
	h := &HEPInput{
		buffer: &sync.Pool{New: func() interface{} { return make([]byte, maxPktLen) }},
		quit:   make(chan struct{}),
		wg:     &sync.WaitGroup{},
	}
	if len(config.Setting.DBAddr) > 2 {
		h.useDB = true
	}
	if len(config.Setting.MQAddr) > 2 && len(config.Setting.MQDriver) > 2 {
		h.useMQ = true
	}
	if len(config.Setting.PromAddr) > 2 {
		h.usePM = true
	}
	if len(config.Setting.ESAddr) > 2 {
		h.useES = true
	}
	return h
}

func (h *HEPInput) Run() {
	for n := 0; n < runtime.NumCPU()*4; n++ {
		go h.hepWorker()
	}

	if h.useDB {
		go func() {
			d := database.New(config.Setting.DBDriver)
			d.Chan = dbCh

			if err := d.Run(); err != nil {
				logp.Err("%v", err)
			}
		}()
	}

	if h.useMQ {
		go func() {
			q := queue.New(config.Setting.MQDriver)
			q.Topic = config.Setting.MQTopic
			q.Chan = mqCh

			if err := q.Run(); err != nil {
				logp.Err("%v", err)
			}
		}()
	}

	if h.usePM {
		go func() {
			m := metric.New("prometheus")
			m.Chan = pmCh

			if err := m.Run(); err != nil {
				logp.Err("%v", err)
			}
		}()
	}

	if h.useES {
		go func() {
			e := elastic.New("elasticsearch")
			e.Chan = esCh

			if err := e.Run(); err != nil {
				logp.Err("%v", err)
			}
		}()
	}

	go h.logStats()
	if config.Setting.HEPAddr != "" {
		h.wg.Add(1)
		go h.serveUDP()
	}
	if config.Setting.HEPTCPAddr != "" {
		go h.serveTCP()
	}
	if config.Setting.HEPTLSAddr != "" {
		go h.serveTLS()
	}
	logp.Info("start %s with %#v\n", config.Version, config.Setting)
}

func (h *HEPInput) serveUDP() {
	ua, err := net.ResolveUDPAddr("udp", config.Setting.HEPAddr)
	if err != nil {
		logp.Critical("%v", err)
	}

	uc, err := net.ListenUDP("udp", ua)
	if err != nil {
		logp.Critical("%v", err)
	}
	defer uc.Close()
	defer h.wg.Done()
	for {
		select {
		case <-h.quit:
			return
		default:
		}
		uc.SetReadDeadline(time.Now().Add(1e9))
		buf := h.buffer.Get().([]byte)
		n, err := uc.Read(buf)
		if err != nil {
			continue
		} else if n > maxPktLen {
			logp.Warn("received too big packet with %d bytes", n)
			atomic.AddUint64(&h.stats.ErrCount, 1)
			continue
		}
		inCh <- buf[:n]
		atomic.AddUint64(&h.stats.PktCount, 1)
	}
}

func (h *HEPInput) serveTCP() {
	listener, err := net.Listen("tcp", config.Setting.HEPTCPAddr)
	if err != nil {
		logp.Err("%v", err)
		return
	}

	for {
		select {
		case <-h.quit:
			listener.Close()
			return
		default:
		}

		conn, err := listener.Accept()
		logp.Info("new TCP connection %s -> %s", conn.RemoteAddr(), conn.LocalAddr())
		if err != nil {
			logp.Err("%v", err)
			continue
		}
		h.wg.Add(1)
		go h.handleTCP(conn)
	}
}

func (h *HEPInput) handleTCP(c net.Conn) {
	r := bufio.NewReader(c)
	defer c.Close()
	defer h.wg.Done()
	for {
		select {
		case <-h.quit:
			return
		default:
		}

		c.SetReadDeadline(time.Now().Add(1e9))
		buf := h.buffer.Get().([]byte)
		hb, err := r.Peek(6)
		if err != nil {
			if opErr, ok := err.(*net.OpError); ok && opErr.Timeout() {
				continue
			} else {
				return
			}
		} else {
			size := binary.BigEndian.Uint16(hb[4:6])
			if size > maxPktLen {
				logp.Warn("unexpected packet, did you send TLS into plain TCP input?")
				return
			}
			n, err := io.ReadFull(r, buf[:size])
			if err != nil || n > maxPktLen {
				logp.Warn("%v, unusal packet size with %d bytes", err, n)
				atomic.AddUint64(&h.stats.ErrCount, 1)
				continue
			}
			inCh <- buf[:n]
			atomic.AddUint64(&h.stats.PktCount, 1)
		}
	}
}

func (h *HEPInput) serveTLS() {
	ca := NewCertificateAuthority()
	listener, err := tls.Listen("tcp", config.Setting.HEPTLSAddr, &tls.Config{
		GetCertificate: ca.GetCertificate,
	})
	if err != nil {
		logp.Err("%v", err)
		return
	}

	for {
		select {
		case <-h.quit:
			listener.Close()
			return
		default:
		}

		conn, err := listener.Accept()
		if err != nil {
			logp.Err("%v", err)
			continue
		}
		h.wg.Add(1)
		go h.handleTLS(conn)
	}
}

func (h *HEPInput) handleTLS(c net.Conn) {
	defer c.Close()
	defer h.wg.Done()
	for {
		select {
		case <-h.quit:
			return
		default:
		}

		c.SetReadDeadline(time.Now().Add(1e9))
		buf := h.buffer.Get().([]byte)
		n, err := c.Read(buf)
		if err != nil {
			if opErr, ok := err.(*net.OpError); ok && opErr.Timeout() {
				continue
			} else {
				return
			}
		} else if n > maxPktLen {
			logp.Warn("received too big packet with %d bytes", n)
			atomic.AddUint64(&h.stats.ErrCount, 1)
			continue
		}
		inCh <- buf[:n]
		atomic.AddUint64(&h.stats.PktCount, 1)
	}
}

func (h *HEPInput) End() {
	logp.Info("stopping heplify-server...")
	close(h.quit)
	h.wg.Wait()
	close(inCh)
	logp.Info("heplify-server has been stopped")
}

func (h *HEPInput) hepWorker() {
	var (
		hepPkt *decoder.HEP
		msg    = h.buffer.Get().([]byte)
		err    error
		ok     bool
		dbCnt  int
		mqCnt  int
		pmCnt  int
		esCnt  int
	)

OUT:
	for {
		h.buffer.Put(msg[:maxPktLen])

		select {
		case msg, ok = <-inCh:
			if !ok {
				break OUT
			}
		}

		hepPkt, err = decoder.DecodeHEP(msg)
		if err != nil {
			atomic.AddUint64(&h.stats.ErrCount, 1)
			continue
		} else if hepPkt.ProtoType == 0 {
			atomic.AddUint64(&h.stats.DupCount, 1)
			continue
		} else if hepPkt.Payload == "DISCARD" {
			continue
		}

		atomic.AddUint64(&h.stats.HEPCount, 1)

		if h.useDB {
			select {
			case dbCh <- hepPkt:
			default:
				dbCnt++
				if dbCnt%1024 == 0 {
					dbCnt = 0
					logp.Warn("overflowing db channel by 1024 packets")
				}
			}
		}

		if h.usePM {
			select {
			case pmCh <- hepPkt:
			default:
				pmCnt++
				if pmCnt%1024 == 0 {
					pmCnt = 0
					logp.Warn("overflowing metric channel by 1024 packets")
				}
			}
		}

		if h.useMQ {
			select {
			case mqCh <- append([]byte{}, msg...):
			default:
				mqCnt++
				if mqCnt%1024 == 0 {
					mqCnt = 0
					logp.Warn("overflowing queue channel by 1024 packets")
				}
			}
		}

		if h.useES {
			select {
			case esCh <- hepPkt:
			default:
				esCnt++
				if esCnt%1024 == 0 {
					esCnt = 0
					logp.Warn("overflowing elastic channel by 1024 packets")
				}
			}
		}
	}
}

func (h *HEPInput) logStats() {
	ticker := time.NewTicker(5 * time.Minute)
	for {
		select {
		case <-ticker.C:
			logp.Info("stats since last 5 minutes. PPS: %d, HEP: %d, Duplicate: %d, Error: %d",
				atomic.LoadUint64(&h.stats.PktCount)/300,
				atomic.LoadUint64(&h.stats.HEPCount),
				atomic.LoadUint64(&h.stats.DupCount),
				atomic.LoadUint64(&h.stats.ErrCount),
			)
			atomic.StoreUint64(&h.stats.PktCount, 0)
			atomic.StoreUint64(&h.stats.HEPCount, 0)
			atomic.StoreUint64(&h.stats.DupCount, 0)
			atomic.StoreUint64(&h.stats.ErrCount, 0)
		}
	}
}
