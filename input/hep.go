package input

import (
	"bytes"
	"net"
	"runtime"
	"strconv"
	"sync"
	"sync/atomic"
	"time"

	hep "github.com/negbie/heplify-server"
	"github.com/negbie/heplify-server/config"
	"github.com/negbie/heplify-server/logp"
	"github.com/negbie/heplify-server/output"
)

type HEPInput struct {
	addr    string
	port    int
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
	hepOutCh = make(chan []byte, 20000)

	hepBuffer = &sync.Pool{
		New: func() interface{} {
			return make([]byte, 65536)
		},
	}
)

func NewHEP() *HEPInput {
	return &HEPInput{
		port:    config.Cfg.HEPPort,
		workers: config.Cfg.HEPWorkers,
		pool:    make(chan chan struct{}, runtime.NumCPU()*1e2),
	}
}

func (h *HEPInput) Run() {
	hostPort := net.JoinHostPort(h.addr, strconv.Itoa(h.port))
	udpAddr, _ := net.ResolveUDPAddr("udp", hostPort)

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
		o := output.New(config.Cfg.OutName)
		o.ConfFile = config.Cfg.OutConfigFile
		o.ErrCount = &h.stats.ErrCount
		o.Topic = config.Cfg.HEPTopic

		if err := o.Run(); err != nil {
			logp.Err("%v", err)
		}
	}()

	for !h.stop {
		buf := hepBuffer.Get().([]byte)
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
		hepPkt *hep.Packet
		msg    = hepBuffer.Get().([]byte)
		buf    = new(bytes.Buffer)
		err    error
		ok     bool
		b      []byte
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

		hepPkt, err = hep.Decode(msg)
		if hepPkt == nil || err != nil {
			continue
		}

		atomic.AddUint64(&h.stats.HEPCount, 1)

		if hepPkt != nil {

			select {
			case hepOutCh <- b:
			default:
			}
		}
	}
}
