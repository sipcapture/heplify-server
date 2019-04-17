package input

import (
	"runtime"
	"sync"
	"sync/atomic"
	"time"

	"github.com/negbie/logp"
	"github.com/sipcapture/heplify-server/config"
	"github.com/sipcapture/heplify-server/database"
	"github.com/sipcapture/heplify-server/decoder"
	"github.com/sipcapture/heplify-server/metric"
	"github.com/sipcapture/heplify-server/queue"
	"github.com/sipcapture/heplify-server/remotelog"
	"github.com/sipcapture/heplify-server/rotator"
)

type HEPInput struct {
	dbCh      chan *decoder.HEP
	pmCh      chan *decoder.HEP
	esCh      chan *decoder.HEP
	lkCh      chan *decoder.HEP
	inCh      chan []byte
	mqCh      chan []byte
	wg        *sync.WaitGroup
	buffer    *sync.Pool
	exitedTCP chan bool
	exitedTLS chan bool
	quitUDP   chan bool
	quitTCP   chan bool
	quitTLS   chan bool
	quit      chan bool
	stats     HEPStats
	useDB     bool
	useMQ     bool
	usePM     bool
	useES     bool
	useLK     bool
}

type HEPStats struct {
	DupCount uint64
	ErrCount uint64
	HEPCount uint64
	PktCount uint64
}

const maxPktLen = 8192

func NewHEPInput() *HEPInput {
	h := &HEPInput{
		inCh:      make(chan []byte, 40000),
		buffer:    &sync.Pool{New: func() interface{} { return make([]byte, maxPktLen) }},
		wg:        &sync.WaitGroup{},
		quit:      make(chan bool),
		quitUDP:   make(chan bool),
		quitTCP:   make(chan bool),
		quitTLS:   make(chan bool),
		exitedTCP: make(chan bool),
		exitedTLS: make(chan bool),
	}
	if len(config.Setting.DBAddr) > 2 {
		h.useDB = true
		h.dbCh = make(chan *decoder.HEP, config.Setting.DBBuffer)
	}
	if len(config.Setting.MQAddr) > 2 && len(config.Setting.MQDriver) > 2 {
		h.useMQ = true
		h.mqCh = make(chan []byte, 40000)
	}
	if len(config.Setting.PromAddr) > 2 {
		h.usePM = true
		h.pmCh = make(chan *decoder.HEP, 40000)
	}
	if len(config.Setting.ESAddr) > 2 {
		h.useES = true
		h.esCh = make(chan *decoder.HEP, 40000)
	}
	if len(config.Setting.LokiURL) > 2 {
		h.useLK = true
		h.lkCh = make(chan *decoder.HEP, config.Setting.LokiBuffer)
	}

	return h
}

func (h *HEPInput) Run() {
	for n := 0; n < runtime.NumCPU(); n++ {
		h.wg.Add(1)
		go h.hepWorker()
	}

	logp.Info("start %s with %#v\n", config.Version, config.Setting)
	go h.logStats()

	if config.Setting.HEPAddr != "" {
		go h.serveUDP(config.Setting.HEPAddr)
	}
	if config.Setting.HEPTCPAddr != "" {
		go h.serveTCP(config.Setting.HEPTCPAddr)
	}
	if config.Setting.HEPTLSAddr != "" {
		go h.serveTLS(config.Setting.HEPTLSAddr)
	}

	if h.usePM {
		m := metric.New("prometheus")
		m.Chan = h.pmCh

		if err := m.Run(); err != nil {
			logp.Err("%v", err)
		}
		defer m.End()
	}

	if h.useMQ {
		q := queue.New(config.Setting.MQDriver)
		q.Topic = config.Setting.MQTopic
		q.Chan = h.mqCh

		if err := q.Run(); err != nil {
			logp.Err("%v", err)
		}
		defer q.End()
	}

	if h.useES {
		r := remotelog.New("elasticsearch")
		r.Chan = h.esCh

		if err := r.Run(); err != nil {
			logp.Err("%v", err)
		}
		defer r.End()
	}

	if h.useLK {
		l := remotelog.New("loki")
		l.Chan = h.lkCh

		if err := l.Run(); err != nil {
			logp.Err("%v", err)
		}
		defer l.End()
	}

	if config.Setting.DBRotate {
		r := rotator.Setup(h.quit)
		r.Rotate()
		defer r.End()
	}

	if h.useDB {
		d := database.New(config.Setting.DBDriver)
		d.Chan = h.dbCh

		if err := d.Run(); err != nil {
			logp.Err("%v", err)
		}
		defer d.End()
	}

	h.wg.Wait()
}

func (h *HEPInput) End() {
	logp.Info("stopping heplify-server...")

	if config.Setting.HEPAddr != "" {
		h.quitUDP <- true
		<-h.quitUDP
	}
	if config.Setting.HEPTCPAddr != "" {
		close(h.quitTCP)
		<-h.exitedTCP
	}
	if config.Setting.HEPTLSAddr != "" {
		close(h.quitTLS)
		<-h.exitedTLS
	}

	h.quit <- true
	<-h.quit

	logp.Info("heplify-server has been stopped")
}

func (h *HEPInput) hepWorker() {
	lastWarn := time.Now()
	msg := h.buffer.Get().([]byte)

	for {
		h.buffer.Put(msg[:maxPktLen])
		select {
		case <-h.quit:
			h.quit <- true
			h.wg.Done()
			return
		case msg = <-h.inCh:
			hepPkt, err := decoder.DecodeHEP(msg)
			if err != nil {
				atomic.AddUint64(&h.stats.ErrCount, 1)
				continue
			} else if hepPkt.ProtoType == 0 {
				atomic.AddUint64(&h.stats.DupCount, 1)
				continue
			}
			atomic.AddUint64(&h.stats.HEPCount, 1)

			if h.useDB {
				select {
				case h.dbCh <- hepPkt:
				default:
					if time.Since(lastWarn) > 5e8 {
						logp.Warn("overflowing db channel, please adjust DBWorker or DBBuffer setting")
					}
					lastWarn = time.Now()
				}
			}

			if h.usePM {
				select {
				case h.pmCh <- hepPkt:
				default:
					if time.Since(lastWarn) > 5e8 {
						logp.Warn("overflowing metric channel")
					}
					lastWarn = time.Now()
				}
			}

			if h.useMQ {
				select {
				case h.mqCh <- append([]byte{}, msg...):
				default:
					if time.Since(lastWarn) > 5e8 {
						logp.Warn("overflowing queue channel")
					}
					lastWarn = time.Now()
				}
			}

			if h.useES {
				select {
				case h.esCh <- hepPkt:
				default:
					if time.Since(lastWarn) > 5e8 {
						logp.Warn("overflowing elasticsearch channel")
					}
					lastWarn = time.Now()
				}
			}

			if h.useLK {
				select {
				case h.lkCh <- hepPkt:
				default:
					if time.Since(lastWarn) > 5e8 {
						logp.Warn("overflowing loki channel")
					}
					lastWarn = time.Now()
				}
			}
		}
	}
}

func (h *HEPInput) logStats() {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			logp.Info("stats since last 5 minutes. PPS: %d, HEP: %d, Filtered: %d, Error: %d",
				atomic.LoadUint64(&h.stats.PktCount)/300,
				atomic.LoadUint64(&h.stats.HEPCount),
				atomic.LoadUint64(&h.stats.DupCount),
				atomic.LoadUint64(&h.stats.ErrCount),
			)
			atomic.StoreUint64(&h.stats.PktCount, 0)
			atomic.StoreUint64(&h.stats.HEPCount, 0)
			atomic.StoreUint64(&h.stats.DupCount, 0)
			atomic.StoreUint64(&h.stats.ErrCount, 0)

		case <-h.quit:
			h.quit <- true
			return
		}
	}
}
