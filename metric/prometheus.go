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
	TargetIP          []string
	TargetName        []string
	GaugeMetrics      map[string]prometheus.Gauge
	GaugeVecMetrics   map[string]*prometheus.GaugeVec
	CounterVecMetrics map[string]*prometheus.CounterVec
}

func cutSpace(str string) string {
	return strings.Map(func(r rune) rune {
		if unicode.IsSpace(r) {
			return -1
		}
		return r
	}, str)
}

func (p *Prometheus) setup() error {
	var err error

	promTargetIP := cutSpace(config.Setting.PromTargetIP)
	promTargetName := cutSpace(config.Setting.PromTargetName)

	p.TargetIP = strings.Split(promTargetIP, ",")
	p.TargetName = strings.Split(promTargetName, ",")

	dup := make(map[string]bool)

	uniqueNames := []string{}
	for _, tn := range p.TargetName {
		if _, v := dup[tn]; !v {
			dup[tn] = true
			uniqueNames = append(uniqueNames, tn)
		} else {
			return fmt.Errorf("please give every prometheus target a unique name")
		}
		if len(tn) < 2 {
			return fmt.Errorf("please give every prometheus target name at least 2 characters")
		}
	}

	uniqueIP := []string{}
	for _, ti := range p.TargetIP {
		if _, ok := dup[ti]; !ok {
			dup[ti] = true
			uniqueIP = append(uniqueIP, ti)
		} else {
			return fmt.Errorf("please give every prometheus target a unique IP")
		}
		if len(ti) < 7 {
			return fmt.Errorf("please give every prometheus target IP at least 7 characters")
		}
	}

	p.TargetIP = uniqueIP
	p.TargetName = uniqueNames

	if len(p.TargetIP) != len(p.TargetName) {
		return fmt.Errorf("please give every prometheus target a IP address and a unique name")
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

	for _, tn := range p.TargetName {
		p.GaugeMetrics[tn+"call_setup_time"] = prometheus.NewGauge(prometheus.GaugeOpts{
			Name: tn + "call_setup_time",
			Help: "SIP call setup time",
		})

		p.CounterVecMetrics[tn+"_method_response"] = prometheus.NewCounterVec(prometheus.CounterOpts{
			Name: tn + "_method_response",
			Help: "SIP method and response counter",
		}, []string{"method", "response"})

	}

	for k := range p.GaugeMetrics {
		logp.Info("register prometheus gaugeMetric %s", k)
		prometheus.MustRegister(p.GaugeMetrics[k])
	}
	for k := range p.GaugeVecMetrics {
		logp.Info("register prometheus gaugeVecMetric %s", k)
		prometheus.MustRegister(p.GaugeVecMetrics[k])
	}
	for k := range p.CounterVecMetrics {
		logp.Info("register prometheus counterVecMetric %s", k)
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

func (p *Prometheus) collect(mCh chan *decoder.HEPPacket) {
	var (
		pkt *decoder.HEPPacket
		ok  bool
	)

	for {
		pkt, ok = <-mCh
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

		if pkt.SipMsg != nil {

			for k, tn := range p.TargetName {
				if pkt.SrcIP == p.TargetIP[k] || pkt.DstIP == p.TargetIP[k] {
					p.CounterVecMetrics[tn+"_method_response"].WithLabelValues(pkt.SipMsg.StartLine.Method, pkt.SipMsg.Cseq.Method).Inc()

					if pkt.SipMsg.RTPStatVal != "" {
						p.dissectStats(tn, pkt.SipMsg.RTPStatVal)
					}
				}

			}
		}
	}
}

func (p *Prometheus) dissectStats(tn, stats string) {
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
			cs, err := strconv.Atoi(v)
			if err == nil {
				p.GaugeMetrics[tn+"call_setup_time"].Set(float64(cs / 1000))
			} else {
				logp.Err("%v", err)
			}
		}
	}
}
