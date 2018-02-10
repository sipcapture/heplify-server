package input

import (
	"bytes"
	"net"
	"runtime"
	"sync"
	"sync/atomic"
	"time"

	"github.com/negbie/heplify-server"
	"github.com/negbie/heplify-server/config"
	"github.com/negbie/heplify-server/database"
	"github.com/negbie/heplify-server/logp"
)

type HEPInput struct {
	addr    string
	pool    chan chan struct{}
	stats   HEPStats
	stop    bool
	workers int
}

type HEPStats struct {
	PktCount uint64
	HEPCount uint64
	ErrCount uint64
	Workers  int32
}

var (
	hepInCh  = make(chan []byte, 20000)
	hepOutCh = make(chan *decoder.HEPPacket, 20000)

	hepBuffer = &sync.Pool{
		New: func() interface{} {
			return make([]byte, 65536)
		},
	}
)

func NewHEP() *HEPInput {
	return &HEPInput{
		addr:    config.Setting.HEPAddr,
		workers: config.Setting.HEPWorkers,
		pool:    make(chan chan struct{}, runtime.NumCPU()*1e2),
	}
}

func (h *HEPInput) Run() {
	udpAddr, _ := net.ResolveUDPAddr("udp", h.addr)

	conn, err := net.ListenUDP("udp", udpAddr)
	if err != nil {
		logp.Critical("%v", err)
	}

	atomic.AddInt32(&h.stats.Workers, int32(h.workers))
	for n := 0; n < h.workers; n++ {
		go func() {
			shut := make(chan struct{})
			h.pool <- shut
			h.hepWorker(shut)
		}()
	}

	logp.Info("heplify-server is running")

	go func() {
		d := database.New("mysql")
		d.ErrCount = &h.stats.ErrCount
		d.Topic = config.Setting.HEPTopic
		d.Chan = hepOutCh

		if err := d.Run(); err != nil {
			logp.Err("%v", err)
		}
	}()

	for !h.stop {
		buf := make([]byte, 65536)
		conn.SetReadDeadline(time.Now().Add(1e9))
		n, _, err := conn.ReadFromUDP(buf)
		if err != nil {
			continue
		}
		atomic.AddUint64(&h.stats.PktCount, 1)
		hepInCh <- buf[:n]
	}
}

func (h *HEPInput) End() {
	h.stop = true
	time.Sleep(2 * time.Second)
	close(hepInCh)
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

		hepBuffer.Put(msg)
		buf.Reset()

		select {
		case <-shut:
			break GO
		case msg, ok = <-hepInCh:
			if !ok {
				break GO
			}
		}

		hepPkt, err = decoder.DecodeHEP(msg)
		if hepPkt == nil || err != nil {
			continue
		}

		atomic.AddUint64(&h.stats.HEPCount, 1)

		if hepPkt != nil {
			select {
			case hepOutCh <- hepPkt:
			default:
			}
		}
	}
}
