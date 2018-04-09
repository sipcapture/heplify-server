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
	"github.com/negbie/heplify-server/logp"
	"github.com/negbie/heplify-server/metric"
	"github.com/negbie/heplify-server/queue"
)

type HEPInput struct {
	addr    string
	ch      chan bool
	wg      *sync.WaitGroup
	pool    chan chan struct{}
	stats   HEPStats
	stop    bool
	workers int
}

type HEPStats struct {
	DupCount uint64
	ErrCount uint64
	HEPCount uint64
	PktCount uint64
}

var (
	inCh  = make(chan []byte, 10000)
	dbCh  = make(chan *decoder.HEP, 10000)
	mqCh  = make(chan []byte, 10000)
	mCh   = make(chan *decoder.HEP, 10000)
	dbCnt int
	mqCnt int
	mCnt  int

	hepBuffer = &sync.Pool{
		New: func() interface{} {
			return make([]byte, 8192)
		},
	}
)

func NewHEP() *HEPInput {
	h := &HEPInput{
		addr:    config.Setting.HEPAddr,
		ch:      make(chan bool),
		wg:      &sync.WaitGroup{},
		workers: config.Setting.HEPWorkers,
		pool:    make(chan chan struct{}, runtime.NumCPU()*1e4),
	}
	h.wg.Add(1)
	return h
}

func (h *HEPInput) Run() {
	var (
		err error
		ua  = &net.UDPAddr{}
		uc  = &net.UDPConn{}
		ta  = &net.TCPAddr{}
		tl  = &net.TCPListener{}
	)
	if config.Setting.Network == "udp" {
		ua, err = net.ResolveUDPAddr("udp", h.addr)
		if err != nil {
			logp.Critical("%v", err)
		}

		uc, err = net.ListenUDP("udp", ua)
		if err != nil {
			logp.Critical("%v", err)
		}
		defer uc.Close()
	} else if config.Setting.Network == "tcp" || config.Setting.Network == "tls" {
		ta, err = net.ResolveTCPAddr("tcp", h.addr)
		if err != nil {
			logp.Critical("%v", err)
		}

		tl, err = net.ListenTCP("tcp", ta)
		if err != nil {
			logp.Critical("%v", err)
		}
		defer tl.Close()
	} else {
		logp.Critical("Not supported network type %s", config.Setting.Network)
	}

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

	if config.Setting.MQAddr != "" && config.Setting.MQName != "" {
		go func() {
			q := queue.New(config.Setting.MQName)
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
			m.Chan = mCh

			if err := m.Run(); err != nil {
				logp.Err("%v", err)
			}
		}()
	}

	logp.Info("hep input address: %s:%s, workders: %d\n", config.Setting.Network, h.addr, h.workers)
	go h.logStats()

	if config.Setting.Network == "udp" {
		for !h.stop {
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
	} else if config.Setting.Network == "tcp" || config.Setting.Network == "tls" {
		h.serveTCP(tl)
	}
}

func (h *HEPInput) serveTCP(tcpListener *net.TCPListener) {
	var listener net.Listener
	if config.Setting.Network == "tls" {
		ca := NewCertificateAuthority()
		listener = tls.NewListener(tcpListener, &tls.Config{
			GetCertificate: ca.GetCertificate,
		})
	} else {
		listener = tcpListener
	}
	defer h.wg.Done()
	for {
		select {
		case <-h.ch:
			listener.Close()
			return
		default:
		}
		tcpListener.SetDeadline(time.Now().Add(1e9))
		conn, err := listener.Accept()
		if nil != err {
			if opErr, ok := err.(*net.OpError); ok && opErr.Timeout() {
				continue
			}
		}
		h.wg.Add(1)
		go h.handleTCP(conn)
	}
}

func (h *HEPInput) handleTCP(tc net.Conn) {
	defer tc.Close()
	defer h.wg.Done()
	for {
		select {
		case <-h.ch:
			return
		default:
		}
		tc.SetReadDeadline(time.Now().Add(1e9))
		buf := hepBuffer.Get().([]byte)
		n, err := tc.Read(buf)
		if nil != err {
			if opErr, ok := err.(*net.OpError); ok && opErr.Timeout() {
				continue
			} else {
				return
			}
		}
		atomic.AddUint64(&h.stats.PktCount, 1)
		inCh <- buf[:n]
	}
}

func (h *HEPInput) stopTCP() {
	close(h.ch)
	h.wg.Wait()
}

func (h *HEPInput) End() {
	logp.Info("stopping heplify-server...")
	h.stop = true
	if config.Setting.Network == "tcp" || config.Setting.Network == "tls" {
		h.stopTCP()
	}
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
				if dbCnt%128 == 0 {
					dbCnt = 0
					logp.Warn("overflowing db channel by 128 packets")
				}
			}
		}

		if config.Setting.PromAddr != "" {
			select {
			case mCh <- hepPkt:
			default:
				mCnt++
				if mCnt%128 == 0 {
					mCnt = 0
					logp.Warn("overflowing metric channel by 128 packets")
				}
			}
		}

		if config.Setting.MQAddr != "" {
			select {
			case mqCh <- msg:
			default:
				mqCnt++
				if mqCnt%128 == 0 {
					mqCnt = 0
					logp.Warn("overflowing queue channel by 128 packets")
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
			logp.Info("Packet stats since last minute Total: %d, HEP: %d, Duplicate: %d, Error: %d",
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
