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
	PktCount uint64
	HEPCount uint64
	DupCount uint64
	ErrCount uint64
}

var (
	inCh     = make(chan []byte, 10000)
	dbCh     = make(chan *decoder.HEPPacket, 10000)
	metricCh = make(chan *decoder.HEPPacket, 10000)

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
		pool:     make(chan chan struct{}, runtime.NumCPU()*1e2),
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

	logp.Info("heplify-server is listening at %s with %d workers", h.addr, h.workers)

	go func() {
		d := database.New("mysql")
		d.ErrCount = &h.stats.ErrCount
		d.Chan = dbCh

		if err := d.Run(); err != nil {
			logp.Err("%v", err)
		}
	}()

	go func() {
		m := metric.New("prometheus")
		m.Chan = metricCh

		if err := m.Run(); err != nil {
			logp.Err("%v", err)
		}
	}()

	for !h.stop {
		buf := hepBuffer.Get().([]byte)
		conn.SetReadDeadline(time.Now().Add(1e9))
		n, _, err := conn.ReadFrom(buf)
		if err != nil {
			continue
		}
		atomic.AddUint64(&h.stats.PktCount, 1)
		inCh <- buf[:n]
	}
}

func (h *HEPInput) End() {
	h.stop = true
	time.Sleep(1 * time.Second)
	logp.Info("heplify-server has been stopped")
	close(inCh)
}

func (h *HEPInput) hepWorker(shut chan struct{}) {
	var (
		hepPkt *decoder.HEPPacket
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
		if err != nil {
			continue
		}

		atomic.AddUint64(&h.stats.HEPCount, 1)

		select {
		case dbCh <- hepPkt:
		default:
			logp.Warn("overflowing db channel")
		}

		select {
		case metricCh <- hepPkt:
		default:
			logp.Warn("overflowing metric channel")
		}
	}
}
