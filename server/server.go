package input

import (
	"os"
	"os/signal"
	"runtime"
	"sync"
	"sync/atomic"
	"syscall"
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
	inputCh    chan []byte
	dbCh       chan *decoder.HEP
	promCh     chan *decoder.HEP
	esCh       chan *decoder.HEP
	lokiCh     chan *decoder.HEP
	wg         *sync.WaitGroup
	buffer     *sync.Pool
	exitUDP    chan bool
	exitTCP    chan bool
	exitTLS    chan bool
	exitWS     chan bool
	exitWorker chan bool
	quit       chan bool
	stopped    uint32
	stats      HEPStats
	useDB      bool
	usePM      bool
	useES      bool
	useLK      bool
}

type HEPStats struct {
	DupCount uint64
	ErrCount uint64
	HEPCount uint64
	PktCount uint64
}

const maxPktLen = 65507

func NewHEPInput() *HEPInput {
	h := &HEPInput{
		inputCh:    make(chan []byte, 40000),
		buffer:     &sync.Pool{New: func() interface{} { return make([]byte, maxPktLen) }},
		wg:         &sync.WaitGroup{},
		quit:       make(chan bool),
		exitUDP:    make(chan bool),
		exitTCP:    make(chan bool),
		exitTLS:    make(chan bool),
		exitWS:     make(chan bool),
		exitWorker: make(chan bool),
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
		go h.worker()
	}

	s := config.Setting
	s.DBPass = "<private>"
	logp.Info("start %s with %#v\n", config.Version, s)
	go h.logStats()
	go h.reloadWorker()

	if len(config.Setting.HEPAddr) > 2 {
		go h.serveUDP(config.Setting.HEPAddr)
	}
	if len(config.Setting.HEPWSAddr) > 2 {
		go h.serveWS(config.Setting.HEPWSAddr)
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

	if len(config.Setting.HEPAddr) > 2 {
		<-h.exitUDP
	}
	if len(config.Setting.HEPTCPAddr) > 2 {
		<-h.exitTCP
	}
	if len(config.Setting.HEPWSAddr) > 2 {
		<-h.exitWS
	}
	if len(config.Setting.HEPTLSAddr) > 2 {
		<-h.exitTLS
	}

	h.exitWorker <- true
	<-h.exitWorker

	h.quit <- true
	<-h.quit
	close(h.inputCh)
}

func (h *HEPInput) worker() {
	defer h.wg.Done()

	var ok bool
	var err error
	var script decoder.ScriptEngine
	lastWarn := time.Now()
	msg := h.buffer.Get().([]byte)
	useScript := config.Setting.ScriptEnable

	if useScript {
		script, err = decoder.NewScriptEngine()
		if err != nil {
			logp.Err("%v, please fix and run killall -HUP heplify-server", err)
			useScript = false
		} else {
			defer script.Close()
		}
	}

	for {
		h.buffer.Put(msg[:maxPktLen])
		select {
		case <-h.exitWorker:
			h.exitWorker <- true
			return
		case msg, ok = <-h.inputCh:
			if !ok {
				return
			}
			hepPkt, err := decoder.DecodeHEP(msg)
			if err != nil {
				atomic.AddUint64(&h.stats.ErrCount, 1)
				continue
			} else if hepPkt.ProtoType == 0 {
				atomic.AddUint64(&h.stats.DupCount, 1)
				continue
			}
			atomic.AddUint64(&h.stats.HEPCount, 1)

			if useScript {
				for _, v := range config.Setting.ScriptHEPFilter {
					if hepPkt.ProtoType == uint32(v) {
						if err = script.Run(hepPkt); err != nil {
							logp.Err("%v", err)
						}
						break
					}
				}
				if hepPkt == nil || hepPkt.ProtoType == 1 && hepPkt.SIP == nil {
					logp.Warn("nil struct after script processing")
					continue
				}
			}

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
				logp.Debug("send data: ", "%+v\n\n", hepPkt)
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
				for _, v := range config.Setting.LokiHEPFilter {
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

func (h *HEPInput) reloadWorker() {
	s := make(chan os.Signal, 1)
	defer close(s)
	signal.Notify(s, syscall.SIGHUP)

	for {
		select {
		case <-s:
			logp.Info("reload all worker")
			h.wg.Add(1)

			h.exitWorker <- true
			<-h.exitWorker

			for n := 0; n < runtime.NumCPU(); n++ {
				h.wg.Add(1)
				go h.worker()
			}
			h.wg.Done()
		case <-h.quit:
			h.quit <- true
			return
		}
	}
}
