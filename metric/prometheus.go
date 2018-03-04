package metric

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"unicode"

	"github.com/negbie/heplify-server"
	"github.com/negbie/heplify-server/config"
	"github.com/negbie/heplify-server/logp"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

type Prometheus struct {
	HunterIP          []string
	HunterName        []string
	GaugeMetrics      map[string]prometheus.Gauge
	GaugeVecMetrics   map[string]*prometheus.GaugeVec
	CounterVecMetrics map[string]*prometheus.CounterVec
}

func (p *Prometheus) setup() error {
	var err error

	promHunterIP := cutSpace(config.Setting.PromHunterIP)
	promHunterName := cutSpace(config.Setting.PromHunterName)

	p.HunterIP = strings.Split(promHunterIP, ",")
	p.HunterName = strings.Split(promHunterName, ",")

	dup := make(map[string]bool)

	uniqueNames := []string{}
	for _, tn := range p.HunterName {
		if _, v := dup[tn]; !v {
			dup[tn] = true
			uniqueNames = append(uniqueNames, tn)
		} else {
			return fmt.Errorf("please give every prometheus Hunter a unique name")
		}
		if len(tn) < 2 {
			return fmt.Errorf("please give every prometheus Hunter name at least 2 characters")
		}
	}

	uniqueIP := []string{}
	for _, ti := range p.HunterIP {
		if _, ok := dup[ti]; !ok {
			dup[ti] = true
			uniqueIP = append(uniqueIP, ti)
		} else {
			return fmt.Errorf("please give every prometheus Hunter a unique IP")
		}
		if len(ti) < 7 {
			return fmt.Errorf("please give every prometheus Hunter IP at least 7 characters")
		}
	}

	p.HunterIP = uniqueIP
	p.HunterName = uniqueNames

	if len(p.HunterIP) != len(p.HunterName) {
		return fmt.Errorf("please give every prometheus Hunter a IP address and a unique name")
	}

	p.GaugeMetrics = map[string]prometheus.Gauge{}
	p.GaugeVecMetrics = map[string]*prometheus.GaugeVec{}
	p.CounterVecMetrics = map[string]*prometheus.CounterVec{}

	p.GaugeVecMetrics["hep_size"] = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "hep_size",
		Help: "HEP packet size",
	}, []string{"type"})

	p.CounterVecMetrics["hep_packets"] = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "hep_packets",
		Help: "HEP packet type counter",
	}, []string{"type"})

	for _, tn := range p.HunterName {
		p.GaugeMetrics[tn+"_xrtp_cs"] = prometheus.NewGauge(prometheus.GaugeOpts{
			Name: tn + "_xrtp_cs",
			Help: "XRTP call setup time",
		})

		p.GaugeMetrics[tn+"_xrtp_jir"] = prometheus.NewGauge(prometheus.GaugeOpts{
			Name: tn + "_xrtp_jir",
			Help: "XRTP received jitter",
		})

		p.GaugeMetrics[tn+"_xrtp_jis"] = prometheus.NewGauge(prometheus.GaugeOpts{
			Name: tn + "_xrtp_jis",
			Help: "XRTP sent jitter",
		})

		p.GaugeMetrics[tn+"_xrtp_plr"] = prometheus.NewGauge(prometheus.GaugeOpts{
			Name: tn + "_xrtp_plr",
			Help: "XRTP received packets lost",
		})

		p.GaugeMetrics[tn+"_xrtp_pls"] = prometheus.NewGauge(prometheus.GaugeOpts{
			Name: tn + "_xrtp_pls",
			Help: "XRTP sent packets lost",
		})

		p.GaugeMetrics[tn+"_xrtp_dle"] = prometheus.NewGauge(prometheus.GaugeOpts{
			Name: tn + "_xrtp_dle",
			Help: "XRTP mean rtt",
		})

		p.GaugeMetrics[tn+"_xrtp_mos"] = prometheus.NewGauge(prometheus.GaugeOpts{
			Name: tn + "_xrtp_mos",
			Help: "XRTP mos",
		})

		p.CounterVecMetrics[tn+"_method_response"] = prometheus.NewCounterVec(prometheus.CounterOpts{
			Name: tn + "_method_response",
			Help: "SIP method and response counter",
		}, []string{"response", "method"})

	}

	for k := range p.GaugeMetrics {
		logp.Info("prometheus register gaugeMetric %s", k)
		prometheus.MustRegister(p.GaugeMetrics[k])
	}
	for k := range p.GaugeVecMetrics {
		logp.Info("prometheus register gaugeVecMetric %s", k)
		prometheus.MustRegister(p.GaugeVecMetrics[k])
	}
	for k := range p.CounterVecMetrics {
		logp.Info("prometheus register counterVecMetric %s", k)
		prometheus.MustRegister(p.CounterVecMetrics[k])
	}

	go func() {
		http.Handle("/metrics", promhttp.Handler())
		err = http.ListenAndServe(config.Setting.PromAddr, nil)
		if err != nil {
			logp.Err("%v", err)
		}
	}()
	return err
}

func (p *Prometheus) collect(hCh chan *decoder.HEP) {
	var (
		pkt *decoder.HEP
		ok  bool
	)

	for {
		pkt, ok = <-hCh
		if !ok {
			break
		}

		if pkt.ProtoType == 1 {
			p.CounterVecMetrics["hep_packets"].WithLabelValues("sip").Inc()
			p.GaugeVecMetrics["hep_size"].WithLabelValues("sip").Set(float64(len(pkt.Payload)))
		} else if pkt.ProtoType == 5 {
			p.CounterVecMetrics["hep_packets"].WithLabelValues("rtcp").Inc()
			p.GaugeVecMetrics["hep_size"].WithLabelValues("rtcp").Set(float64(len(pkt.Payload)))
		} else if pkt.ProtoType == 38 {
			p.CounterVecMetrics["hep_packets"].WithLabelValues("horaclifix").Inc()
			p.GaugeVecMetrics["hep_size"].WithLabelValues("horaclifix").Set(float64(len(pkt.Payload)))
		} else if pkt.ProtoType == 100 {
			p.CounterVecMetrics["hep_packets"].WithLabelValues("log").Inc()
			p.GaugeVecMetrics["hep_size"].WithLabelValues("log").Set(float64(len(pkt.Payload)))
		}

		if pkt.SIP != nil {

			for k, tn := range p.HunterName {
				if pkt.SrcIPString == p.HunterIP[k] || pkt.DstIPString == p.HunterIP[k] {
					p.CounterVecMetrics[tn+"_method_response"].WithLabelValues(pkt.SIP.StartLine.Method, pkt.SIP.Cseq.Method).Inc()

					if pkt.SIP.RTPStatVal != "" {
						p.dissectXRTPStats(tn, pkt.SIP.RTPStatVal)
					}
				}

			}
		}
	}
}

func (p *Prometheus) dissectXRTPStats(tn, stats string) {
	var err error
	cs, pr, ps, plr, pls, jir, jis, dle, r, mos := 0, 0, 0, 0, 0, 0, 0, 0, 0.0, 0.0
	m := make(map[string]string)
	sr := strings.Split(stats, ";")

	for _, pair := range sr {
		ss := strings.Split(pair, "=")
		if len(ss) == 2 {
			m[ss[0]] = ss[1]
		}
	}

	if v, ok := m["CS"]; ok {
		if len(v) >= 1 {
			cs, err = strconv.Atoi(v)
			if err == nil {
				p.GaugeMetrics[tn+"_xrtp_cs"].Set(float64(cs / 1000))
			} else {
				logp.Err("%v", err)
			}
		}
	}
	if v, ok := m["PR"]; ok {
		if len(v) >= 1 {
			pr, err = strconv.Atoi(v)
			if err == nil {
			} else {
				logp.Err("%v", err)
			}
		}
	}
	if v, ok := m["PS"]; ok {
		if len(v) >= 1 {
			ps, err = strconv.Atoi(v)
			if err == nil {
			} else {
				logp.Err("%v", err)
			}
		}
	}
	if v, ok := m["PL"]; ok {
		if len(v) >= 1 {
			pl := strings.Split(v, ",")
			if len(pl) == 2 {
				plr, err = strconv.Atoi(pl[0])
				if err == nil {
					p.GaugeMetrics[tn+"_xrtp_plr"].Set(float64(plr))
				} else {
					logp.Err("%v", err)
				}
				pls, err = strconv.Atoi(pl[1])
				if err == nil {
					p.GaugeMetrics[tn+"_xrtp_pls"].Set(float64(pls))
				} else {
					logp.Err("%v", err)
				}
			}
		}
	}
	if v, ok := m["JI"]; ok {
		if len(v) >= 1 {
			ji := strings.Split(v, ",")
			if len(ji) == 2 {
				jir, err = strconv.Atoi(ji[0])
				if err == nil {
					p.GaugeMetrics[tn+"_xrtp_jir"].Set(float64(jir))
				} else {
					logp.Err("%v", err)
				}
				jis, err = strconv.Atoi(ji[1])
				if err == nil {
					p.GaugeMetrics[tn+"_xrtp_jis"].Set(float64(jis))
				} else {
					logp.Err("%v", err)
				}
			}
		}
	}
	if v, ok := m["DL"]; ok {
		if len(v) >= 1 {
			dl := strings.Split(v, ",")
			if len(dl) == 3 {
				dle, err = strconv.Atoi(dl[0])
				if err == nil {
					p.GaugeMetrics[tn+"_xrtp_dle"].Set(float64(dle))
				} else {
					logp.Err("%v", err)
				}
			}
		}
	}

	if pr == 0 && ps == 0 {
		pr, ps = 1, 1
	}

	loss := ((plr + pls) * 100) / (pr + ps)
	el := (jir * 2) + (dle + 10)

	if el < 160 {
		r = 93.2 - (float64(el) / 40)
	} else {
		r = 93.2 - (float64(el-120) / 10)
	}
	r = r - (float64(loss) * 2.5)

	mos = 1 + (0.035)*r + (0.000007)*r*(r-60)*(100-r)
	p.GaugeMetrics[tn+"_xrtp_mos"].Set(mos)

}

func cutSpace(str string) string {
	return strings.Map(func(r rune) rune {
		if unicode.IsSpace(r) {
			return -1
		}
		return r
	}, str)
}
