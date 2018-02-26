package metric

import (
	"fmt"
	"net/http"
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
		}
	}

	p.TargetName = uniqueNames

	if len(p.TargetIP) != len(p.TargetName) {
		return fmt.Errorf("please give every prometheus target a IP address and a name")
	}

	for _, tn := range p.TargetName {
		if len(tn) < 2 {
			return fmt.Errorf("please give every prometheus target a unique name with at least 2 characters")
		}
	}

	p.GaugeMetrics = map[string]prometheus.Gauge{}
	p.GaugeVecMetrics = map[string]*prometheus.GaugeVec{}
	p.CounterVecMetrics = map[string]*prometheus.CounterVec{}

	p.GaugeMetrics["kind"] = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "kind",
		Help: "SIP packet kind",
	})

	p.GaugeVecMetrics["size"] = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "size",
		Help: "SIP packet size",
	}, []string{"type"})

	p.CounterVecMetrics["packets"] = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "packets",
		Help: "Packet type counter",
	}, []string{"type"})

	for _, tn := range p.TargetName {

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
			p.CounterVecMetrics["packets"].WithLabelValues("sip").Inc()
			p.GaugeVecMetrics["size"].WithLabelValues("sip").Add(float64(len(pkt.Payload) / 1048576))
		} else if pkt.ProtoType == 5 {
			p.CounterVecMetrics["packets"].WithLabelValues("rtcp").Inc()
			p.GaugeVecMetrics["size"].WithLabelValues("rtcp").Add(float64(len(pkt.Payload) / 1048576))
		} else if pkt.ProtoType == 38 {
			p.CounterVecMetrics["packets"].WithLabelValues("horaclifix").Inc()
			p.GaugeVecMetrics["size"].WithLabelValues("horaclifix").Add(float64(len(pkt.Payload) / 1048576))
		} else if pkt.ProtoType == 100 {
			p.CounterVecMetrics["packets"].WithLabelValues("log").Inc()
			p.GaugeVecMetrics["size"].WithLabelValues("log").Add(float64(len(pkt.Payload) / 1048576))
		}

		if pkt.SipMsg != nil {

			for k, tn := range p.TargetName {
				if pkt.SrcIP == p.TargetIP[k] || pkt.DstIP == p.TargetIP[k] {
					p.CounterVecMetrics[tn+"_method_response"].WithLabelValues(pkt.SipMsg.StartLine.Method, pkt.SipMsg.Cseq.Method).Inc()
				}

			}
		}
	}
}
