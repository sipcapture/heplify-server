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
	hepInp := &HEPInput{
		inputCh: make(chan []byte, 40000),
		buffer:  &sync.Pool{New: func() interface{} { return make([]byte, maxPktLen) }},
		wg:      &sync.WaitGroup{},
		quit:    make(chan bool),
		exitTCP: make(chan bool),
		exitTLS: make(chan bool),
		lokiTF:  config.Setting.LokiHEPFilter,
	}
	if len(config.Setting.DBAddr) > 2 {
		hepInp.useDB = true
		hepInp.dbCh = make(chan *decoder.HEP, config.Setting.DBBuffer)
	}
	if len(config.Setting.PromAddr) > 2 {
		hepInp.usePM = true
		hepInp.promCh = make(chan *decoder.HEP, 40000)
	}
	if len(config.Setting.ESAddr) > 2 {
		hepInp.useES = true
		hepInp.esCh = make(chan *decoder.HEP, 40000)
	}
	if len(config.Setting.LokiURL) > 2 {
		hepInp.useLK = true
		hepInp.lokiCh = make(chan *decoder.HEP, config.Setting.LokiBuffer)
	}

	return hepInp
}

func (hepInp *HEPInput) Run() {
	for n := 0; n < runtime.NumCPU(); n++ {
		hepInp.wg.Add(1)
		go hepInp.hepWorker()
	}

	logp.Info("start %s with %#v\n", config.Version, config.Setting)
	go hepInp.logStats()

	if len(config.Setting.HEPAddr) > 2 {
		go hepInp.serveUDP(config.Setting.HEPAddr)
	}
	if len(config.Setting.HEPTCPAddr) > 2 {
		go hepInp.serveTCP(config.Setting.HEPTCPAddr)
	}
	if len(config.Setting.HEPTLSAddr) > 2 {
		go hepInp.serveTLS(config.Setting.HEPTLSAddr)
	}

	if hepInp.usePM {
		m := metric.New("prometheus")
		m.Chan = hepInp.promCh

		if err := m.Run(); err != nil {
			logp.Err("%v", err)
		}
		defer m.End()
	}

	if hepInp.useES {
		r := remotelog.New("elasticsearch")
		r.Chan = hepInp.esCh

		if err := r.Run(); err != nil {
			logp.Err("%v", err)
		}
		defer r.End()
	}

	if hepInp.useLK {
		l := remotelog.New("loki")
		l.Chan = hepInp.lokiCh

		if err := l.Run(); err != nil {
			logp.Err("%v", err)
		}
		defer l.End()
	}

	if hepInp.useDB && config.Setting.DBRotate &&
		(config.Setting.DBDriver == "mysql" || config.Setting.DBDriver == "postgres") {
		r := rotator.Setup(hepInp.quit)
		r.Rotate()
		defer r.End()
	}

	if hepInp.useDB {
		d := database.New(config.Setting.DBDriver)
		d.Chan = hepInp.dbCh

		if err := d.Run(); err != nil {
			logp.Err("%v", err)
		}
		defer d.End()
	}

	hepInp.wg.Wait()
}

func (hepInp *HEPInput) End() {
	atomic.StoreUint32(&hepInp.stopped, 1)

	if len(config.Setting.HEPTCPAddr) > 2 {
		<-hepInp.exitTCP
	}
	if len(config.Setting.HEPTLSAddr) > 2 {
		<-hepInp.exitTLS
	}

	hepInp.quit <- true
	<-hepInp.quit
	close(hepInp.inputCh)
}

func (hepInp *HEPInput) hepWorker() {
	lastWarn := time.Now()
	msg := hepInp.buffer.Get().([]byte)
	var ok bool
	defer hepInp.wg.Done()

	for {
		hepInp.buffer.Put(msg[:maxPktLen])
		select {
		case msg, ok = <-hepInp.inputCh:
			if !ok {
				return
			}
			hepPkt, err := decoder.DecodeHEP(msg)
			if err != nil {
				atomic.AddUint64(&hepInp.stats.ErrCount, 1)
				continue
			} else if hepPkt.AppProto == 0 {
				atomic.AddUint64(&hepInp.stats.DupCount, 1)
				continue
			}
			atomic.AddUint64(&hepInp.stats.HEPCount, 1)

			if hepInp.usePM {
				select {
				case hepInp.promCh <- hepPkt:
				default:
					if time.Since(lastWarn) > 1e9 {
						logp.Warn("overflowing metric channel")
					}
					lastWarn = time.Now()
				}
			}

			if hepInp.useDB {
				select {
				case hepInp.dbCh <- hepPkt:
				default:
					if time.Since(lastWarn) > 1e9 {
						logp.Warn("overflowing db channel, please adjust DBWorker or DBBuffer setting")
					}
					lastWarn = time.Now()
				}
			}

			if hepInp.useES {
				select {
				case hepInp.esCh <- hepPkt:
				default:
					if time.Since(lastWarn) > 1e9 {
						logp.Warn("overflowing elasticsearch channel")
					}
					lastWarn = time.Now()
				}
			}

			if hepInp.useLK {
				for _, v := range hepInp.lokiTF {
					if hepPkt.AppProto == uint32(v) {
						select {
						case hepInp.lokiCh <- hepPkt:
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

func (hepInp *HEPInput) logStats() {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			logp.Info("stats since last 5 minutes. PPS: %d, HEP: %d, Filtered: %d, Error: %d",
				atomic.LoadUint64(&hepInp.stats.PktCount)/300,
				atomic.LoadUint64(&hepInp.stats.HEPCount),
				atomic.LoadUint64(&hepInp.stats.DupCount),
				atomic.LoadUint64(&hepInp.stats.ErrCount),
			)
			atomic.StoreUint64(&hepInp.stats.PktCount, 0)
			atomic.StoreUint64(&hepInp.stats.HEPCount, 0)
			atomic.StoreUint64(&hepInp.stats.DupCount, 0)
			atomic.StoreUint64(&hepInp.stats.ErrCount, 0)

		case <-hepInp.quit:
			hepInp.quit <- true
			return
		}
	}
}
