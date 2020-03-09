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
	"github.com/sipcapture/heplify-server/remotelog"
	"github.com/sipcapture/heplify-server/rotator"
)

type HEPInput struct {
	inputCh chan []byte
	dbCh    chan *decoder.HEP
	promCh  chan *decoder.HEP
	esCh    chan *decoder.HEP
	lokiCh  chan *decoder.HEP
	wg      *sync.WaitGroup
	buffer  *sync.Pool
	exitTCP chan bool
	exitTLS chan bool
	quit    chan bool
	stopped uint32
	stats   HEPStats
	lokiTF  []int
	useDB   bool
	usePM   bool
	useES   bool
	useLK   bool
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
		inputCh: make(chan []byte, 40000),
		buffer:  &sync.Pool{New: func() interface{} { return make([]byte, maxPktLen) }},
		wg:      &sync.WaitGroup{},
		quit:    make(chan bool),
		exitTCP: make(chan bool),
		exitTLS: make(chan bool),
		lokiTF:  config.Setting.LokiHEPFilter,
	}
	if len(config.Setting.DBAddr) > 2 {
		h.useDB = true
		h.dbCh = make(chan *decoder.HEP, config.Setting.DBBuffer)
	}
	if len(config.Setting.PromAddr) > 2 {
		h.usePM = true
		h.promCh = make(chan *decoder.HEP, 40000)
	}
	if len(config.Setting.ESAddr) > 2 {
		h.useES = true
		h.esCh = make(chan *decoder.HEP, 40000)
	}
	if len(config.Setting.LokiURL) > 2 {
		h.useLK = true
		h.lokiCh = make(chan *decoder.HEP, config.Setting.LokiBuffer)
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

	if len(config.Setting.HEPAddr) > 2 {
		go h.serveUDP(config.Setting.HEPAddr)
	}
	if len(config.Setting.HEPTCPAddr) > 2 {
		go h.serveTCP(config.Setting.HEPTCPAddr)
	}
	if len(config.Setting.HEPTLSAddr) > 2 {
		go h.serveTLS(config.Setting.HEPTLSAddr)
	}

	if h.usePM {
		m := metric.New("prometheus")
		m.Chan = h.promCh

		if err := m.Run(); err != nil {
			logp.Err("%v", err)
		}
		defer m.End()
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
		l.Chan = h.lokiCh

		if err := l.Run(); err != nil {
			logp.Err("%v", err)
		}
		defer l.End()
	}

	if h.useDB && config.Setting.DBRotate &&
		(config.Setting.DBDriver == "mysql" || config.Setting.DBDriver == "postgres") {
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
	atomic.StoreUint32(&h.stopped, 1)

	if len(config.Setting.HEPTCPAddr) > 2 {
		<-h.exitTCP
	}
	if len(config.Setting.HEPTLSAddr) > 2 {
		<-h.exitTLS
	}

	h.quit <- true
	<-h.quit
	close(h.inputCh)
}

func (h *HEPInput) hepWorker() {
	defer h.wg.Done()
	lastWarn := time.Now()
	for msg := range h.inputCh {
		h.buffer.Put(msg[:maxPktLen])
		hepPkt, err := decoder.DecodeHEP(msg)
		if err != nil {
			atomic.AddUint64(&h.stats.ErrCount, 1)
			continue
		} else if hepPkt.ProtoType == 0 {
			atomic.AddUint64(&h.stats.DupCount, 1)
			continue
		}
		atomic.AddUint64(&h.stats.HEPCount, 1)

		if h.usePM {
			select {
			case h.promCh <- hepPkt:
			default:
				if time.Since(lastWarn) > 1e9 {
					logp.Warn("overflowing metric channel")
				}
				lastWarn = time.Now()
			}
		}

		if h.useDB {
			select {
			case h.dbCh <- hepPkt:
			default:
				if time.Since(lastWarn) > 1e9 {
					logp.Warn("overflowing db channel, please adjust DBWorker or DBBuffer setting")
				}
				lastWarn = time.Now()
			}
		}

		if h.useES {
			select {
			case h.esCh <- hepPkt:
			default:
				if time.Since(lastWarn) > 1e9 {
					logp.Warn("overflowing elasticsearch channel")
				}
				lastWarn = time.Now()
			}
		}

		if h.useLK {
			for _, v := range h.lokiTF {
				if hepPkt.ProtoType == uint32(v) {
					select {
					case h.lokiCh <- hepPkt:
					default:
						if time.Since(lastWarn) > 1e9 {
							logp.Warn("overflowing loki channel")
						}
						lastWarn = time.Now()
					}
					break
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
