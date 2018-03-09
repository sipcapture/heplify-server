package input

import (
	"bytes"
	"net"
	"runtime"
	"sync"
	"sync/atomic"
	"time"

	"github.com/coocood/freecache"
	"github.com/negbie/heplify-server"
	"github.com/negbie/heplify-server/config"
	"github.com/negbie/heplify-server/database"
	"github.com/negbie/heplify-server/logp"
	"github.com/negbie/heplify-server/metric"
	"github.com/negbie/heplify-server/queue"
)

type HEPInput struct {
	addr     string
	pool     chan chan struct{}
	stats    HEPStats
	stop     bool
	workers  int
	dupCache *freecache.Cache
}

type HEPStats struct {
	DupCount uint64
	ErrCount uint64
	HEPCount uint64
	PktCount uint64
}

var (
	inCh = make(chan []byte, 1000)
	dbCh = make(chan *decoder.HEP, 2000)
	mqCh = make(chan []byte, 1000)
	mCh  = make(chan *decoder.HEP, 2000)

	hepBuffer = &sync.Pool{
		New: func() interface{} {
			return make([]byte, 8192)
		},
	}
)

func NewHEP() *HEPInput {
	return &HEPInput{
		addr:     config.Setting.HEPAddr,
		workers:  config.Setting.HEPWorkers,
		pool:     make(chan chan struct{}, runtime.NumCPU()*1e4),
		dupCache: freecache.NewCache(20 * 1024 * 1024),
	}
}

func (h *HEPInput) Run() {
	udpAddr, err := net.ResolveUDPAddr("udp", h.addr)
	if err != nil {
		logp.Critical("%v", err)
	}

	conn, err := net.ListenUDP("udp", udpAddr)
	if err != nil {
		logp.Critical("%v", err)
	}

	for n := 0; n < h.workers; n++ {
		go func() {
			shut := make(chan struct{})
			h.pool <- shut
			h.hepWorker(shut)
		}()
	}

	logp.Info("hep input address: %s, workders: %d\n", h.addr, h.workers)

	if config.Setting.DBAddr != "" {
		go func() {
			d := database.New(config.Setting.DBDriver)
			d.ErrCount = &h.stats.ErrCount
			d.Chan = dbCh

			if err := d.Run(); err != nil {
				logp.Err("%v", err)
			}
		}()
	}

	if config.Setting.MQAddr != "" && config.Setting.MQName != "" {
		go func() {
			q := queue.New(config.Setting.MQName)
			q.ErrCount = &h.stats.ErrCount
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

	for !h.stop {
		buf := hepBuffer.Get().([]byte)
		conn.SetReadDeadline(time.Now().Add(1e9))
		n, _, err := conn.ReadFrom(buf)
		if err != nil {
			continue
		} else if n > 8192 {
			logp.Warn("received to big packet with %d bytes", n)
			continue
		}
		atomic.AddUint64(&h.stats.PktCount, 1)
		inCh <- buf[:n]
	}
}

func (h *HEPInput) End() {
	h.stop = true
	time.Sleep(2 * time.Second)
	logp.Info("heplify-server has been stopped")
	close(inCh)
}

func (h *HEPInput) hepWorker(shut chan struct{}) {
	var (
		hepPkt *decoder.HEP
		msg    = hepBuffer.Get().([]byte)
		buf    = new(bytes.Buffer)
		err    error
		ok     bool
	)

GO:
	for {

		hepBuffer.Put(msg[:8192])
		buf.Reset()

		select {
		case <-shut:
			break GO
		case msg, ok = <-inCh:
			if !ok {
				break GO
			}
		}

		_, err = h.dupCache.Get(msg)
		if err == nil {
			atomic.AddUint64(&h.stats.DupCount, 1)
			continue
		}
		err = h.dupCache.Set(msg, nil, 4)
		if err != nil {
			logp.Warn("%v", err)
		}

		hepPkt, err = decoder.DecodeHEP(msg)
		if err != nil || hepPkt == nil {
			continue
		}

		atomic.AddUint64(&h.stats.HEPCount, 1)

		if config.Setting.PromAddr != "" {
			select {
			case mCh <- hepPkt:
			default:
				logp.Warn("overflowing metric channel")
			}
		}

		if config.Setting.MQAddr != "" {
			select {
			case mqCh <- msg:
			default:
				logp.Warn("overflowing queue channel")
			}
		}

		if config.Setting.DBAddr != "" {
			dbCh <- hepPkt
		}
	}
}
