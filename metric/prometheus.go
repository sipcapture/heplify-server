package metric

import (
	"fmt"
	"net/http"

	"github.com/negbie/heplify-server"
	"github.com/negbie/heplify-server/config"
	"github.com/negbie/heplify-server/logp"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

type Prometheus struct {
	GaugeMetrics      map[string]prometheus.Gauge
	GaugeVecMetrics   map[string]*prometheus.GaugeVec
	CounterVecMetrics map[string]*prometheus.CounterVec
}

func (p *Prometheus) setup() error {
	var err error

	p.GaugeMetrics = map[string]prometheus.Gauge{}

	p.GaugeMetrics["kind"] = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "kind",
		Help: "SIP packet kind",
	})

	p.GaugeVecMetrics = map[string]*prometheus.GaugeVec{}

	p.GaugeVecMetrics["size"] = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "size",
		Help: "SIP packet size",
	}, []string{"type"})

	p.CounterVecMetrics = map[string]*prometheus.CounterVec{}

	p.CounterVecMetrics["method_response"] = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "method_response",
		Help: "SIP method and response counter",
	}, []string{"method", "response"})

	p.CounterVecMetrics["packets"] = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "packets",
		Help: "Packet type counter",
	}, []string{"type"})

	for k := range p.GaugeMetrics {
		prometheus.MustRegister(p.GaugeMetrics[k])
	}
	for k := range p.GaugeVecMetrics {
		prometheus.MustRegister(p.GaugeVecMetrics[k])
	}
	for k := range p.CounterVecMetrics {
		fmt.Println(k)
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

			p.CounterVecMetrics["method_response"].WithLabelValues(pkt.SipMsg.StartLine.Method, pkt.SipMsg.Cseq.Method).Inc()

		}

	}
}
