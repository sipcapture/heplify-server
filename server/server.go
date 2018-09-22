package input

import (
	"crypto/tls"
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
	addr   string
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

func NewHEPInput() *HEPInput {
	h := &HEPInput{
		addr:   config.Setting.HEPAddr,
		buffer: &sync.Pool{New: func() interface{} { return make([]byte, 8192) }},
		quit:   make(chan struct{}),
		wg:     &sync.WaitGroup{},
	}
	return h
}

func (h *HEPInput) Run() {
	for n := 0; n < runtime.NumCPU()*4; n++ {
		go h.hepWorker()
	}

	if config.Setting.DBAddr != "" {
		go func() {
			d := database.New(config.Setting.DBDriver)
			d.Chan = dbCh

			if err := d.Run(); err != nil {
				logp.Err("%v", err)
			}
		}()
	}

	if config.Setting.MQAddr != "" && config.Setting.MQDriver != "" {
		go func() {
			q := queue.New(config.Setting.MQDriver)
			q.Topic = config.Setting.MQTopic
			q.Chan = mqCh

			if err := q.Run(); err != nil {
				logp.Err("%v", err)
			}
		}()
	}

	if config.Setting.PromAddr != "" {
		go func() {
			m := metric.New("prometheus")
			m.Chan = pmCh

			if err := m.Run(); err != nil {
				logp.Err("%v", err)
			}
		}()
	}

	if config.Setting.ESAddr != "" {
		go func() {
			e := elastic.New("elasticsearch")
			e.Chan = esCh

			if err := e.Run(); err != nil {
				logp.Err("%v", err)
			}
		}()
	}

	h.wg.Add(1)
	go h.serveUDP()
	go h.serveTLS()
	go h.logStats()
	logp.Info("start %s with %#v\n", config.Version, config.Setting)
}

func (h *HEPInput) serveUDP() {
	ua, err := net.ResolveUDPAddr("udp", h.addr)
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
		} else if n > 8192 {
			logp.Warn("received too big packet with %d bytes", n)
			atomic.AddUint64(&h.stats.ErrCount, 1)
			continue
		}
		inCh <- buf[:n]
		atomic.AddUint64(&h.stats.PktCount, 1)
	}
}

func (h *HEPInput) serveTLS() {
	ca := NewCertificateAuthority()
	listener, err := tls.Listen("tcp", h.addr, &tls.Config{
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
		} else if n > 8192 {
			logp.Warn("received too big packet with %d bytes", n)
			atomic.AddUint64(&h.stats.ErrCount, 1)
			continue
		}
		inCh <- buf[:n]
		atomic.AddUint64(&h.stats.PktCount, 1)
	}
}

func (h *HEPInput) closeConn() {
	close(h.quit)
	h.wg.Wait()
}

func (h *HEPInput) End() {
	logp.Info("stopping heplify-server...")
	h.closeConn()
	time.Sleep(1 * time.Second)
	logp.Info("heplify-server has been stopped")
	close(inCh)
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
		h.buffer.Put(msg[:8192])

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

		if config.Setting.DBAddr != "" {
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

		if config.Setting.PromAddr != "" {
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

		if config.Setting.MQAddr != "" {
			select {
			case mqCh <- msg:
			default:
				mqCnt++
				if mqCnt%1024 == 0 {
					mqCnt = 0
					logp.Warn("overflowing queue channel by 1024 packets")
				}
			}
		}

		if config.Setting.ESAddr != "" {
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
			logp.Info("Packet stats since last 5 minutes. Total: %d, HEP: %d, Duplicate: %d, Error: %d",
				atomic.LoadUint64(&h.stats.PktCount),
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
