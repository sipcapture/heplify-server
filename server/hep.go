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
	"github.com/negbie/heplify-server/metric"
	"github.com/negbie/heplify-server/queue"
	"github.com/negbie/logp"
)

type HEPInput struct {
	addr    string
	ch      chan struct{}
	wg      *sync.WaitGroup
	pool    chan chan struct{}
	stats   HEPStats
	isAlive bool
	workers int
}

type HEPStats struct {
	DupCount uint64
	ErrCount uint64
	HEPCount uint64
	PktCount uint64
}

var (
	inCh  = make(chan []byte, 20000)
	dbCh  = make(chan *decoder.HEP, 200000)
	mqCh  = make(chan []byte, 20000)
	pmCh  = make(chan *decoder.HEP, 20000)
	dbCnt int
	mqCnt int
	pmCnt int

	hepBuffer = &sync.Pool{
		New: func() interface{} {
			return make([]byte, 8192)
		},
	}
)

func NewHEP() *HEPInput {
	h := &HEPInput{
		addr:    config.Setting.HEPAddr,
		ch:      make(chan struct{}),
		wg:      &sync.WaitGroup{},
		workers: runtime.NumCPU() * 4,
		pool:    make(chan chan struct{}, runtime.NumCPU()*1e2),
		isAlive: true,
	}
	return h
}

func (h *HEPInput) Run() {
	ua, err := net.ResolveUDPAddr("udp", h.addr)
	if err != nil {
		logp.Critical("%v", err)
	}

	uc, err := net.ListenUDP("udp", ua)
	if err != nil {
		logp.Critical("%v", err)
	}
	defer uc.Close()

	for n := 0; n < h.workers; n++ {
		go func() {
			shut := make(chan struct{})
			h.pool <- shut
			h.hepWorker(shut)
		}()
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

	logp.Info("hep input address: %s, hep workers: %d\n", h.addr, h.workers)
	go h.logStats()
	go h.serveTLS()

	for h.isAlive {
		uc.SetReadDeadline(time.Now().Add(1e9))
		buf := hepBuffer.Get().([]byte)
		n, _, err := uc.ReadFrom(buf)
		if err != nil {
			continue
		} else if n > 8192 {
			logp.Warn("received to big packet with %d bytes", n)
			atomic.AddUint64(&h.stats.ErrCount, 1)
			continue
		}
		atomic.AddUint64(&h.stats.PktCount, 1)
		inCh <- buf[:n]
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
		case <-h.ch:
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
		case <-h.ch:
			return
		default:
		}

		c.SetReadDeadline(time.Now().Add(1e9))
		buf := hepBuffer.Get().([]byte)
		n, err := c.Read(buf)
		if err != nil {
			if opErr, ok := err.(*net.OpError); ok && opErr.Timeout() {
				continue
			} else {
				return
			}
		} else if n > 8192 {
			logp.Warn("received to big packet with %d bytes", n)
			atomic.AddUint64(&h.stats.ErrCount, 1)
			continue
		}
		atomic.AddUint64(&h.stats.PktCount, 1)
		inCh <- buf[:n]
	}
}

func (h *HEPInput) closeTLS() {
	close(h.ch)
	h.wg.Wait()
}

func (h *HEPInput) End() {
	logp.Info("stopping heplify-server...")
	h.isAlive = false
	h.closeTLS()
	time.Sleep(2 * time.Second)
	logp.Info("heplify-server has been stopped")
	close(inCh)

	/* 	for i := 0; i < config.Setting.HEPWorkers; i++ {
		wQuit := <-h.pool
		close(wQuit)
	} */

}

func (h *HEPInput) hepWorker(shut chan struct{}) {
	var (
		hepPkt *decoder.HEP
		msg    = hepBuffer.Get().([]byte)
		err    error
		ok     bool
	)

GO:
	for {
		hepBuffer.Put(msg[:8192])

		select {
		case <-shut:
			break GO
		case msg, ok = <-inCh:
			if !ok {
				break GO
			}
		}

		hepPkt, err = decoder.DecodeHEP(msg)
		if err != nil {
			atomic.AddUint64(&h.stats.ErrCount, 1)
			continue
		} else if hepPkt.ProtoType == 0 {
			atomic.AddUint64(&h.stats.DupCount, 1)
			continue
		}

		atomic.AddUint64(&h.stats.HEPCount, 1)

		if config.Setting.DBAddr != "" {
			select {
			case dbCh <- hepPkt:
			default:
				dbCnt++
				if dbCnt%2048 == 0 {
					dbCnt = 0
					logp.Warn("overflowing db channel by 2048 packets")
				}
			}
		}

		if config.Setting.PromAddr != "" {
			select {
			case pmCh <- hepPkt:
			default:
				pmCnt++
				if pmCnt%2048 == 0 {
					pmCnt = 0
					logp.Warn("overflowing metric channel by 2048 packets")
				}
			}
		}

		if config.Setting.MQAddr != "" {
			select {
			case mqCh <- msg:
			default:
				mqCnt++
				if mqCnt%2048 == 0 {
					mqCnt = 0
					logp.Warn("overflowing queue channel by 2048 packets")
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
