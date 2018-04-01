package metric

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"unicode"

	"github.com/buger/jsonparser"
	"github.com/coocood/freecache"
	"github.com/negbie/heplify-server"
	"github.com/negbie/heplify-server/config"
	"github.com/negbie/heplify-server/logp"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

type Prometheus struct {
	TargetIP          []string
	TargetName        []string
	TargetEmpty       bool
	GaugeMetrics      map[string]prometheus.Gauge
	GaugeVecMetrics   map[string]*prometheus.GaugeVec
	CounterVecMetrics map[string]*prometheus.CounterVec
	Cache             *freecache.Cache
}

func (p *Prometheus) setup() (err error) {
	promTargetIP := cutSpace(config.Setting.PromTargetIP)
	promTargetName := cutSpace(config.Setting.PromTargetName)

	p.TargetIP = strings.Split(promTargetIP, ",")
	p.TargetName = strings.Split(promTargetName, ",")

	dedupIP := make(map[string]bool)
	dedupName := make(map[string]bool)

	uniqueIP := []string{}
	for _, ti := range p.TargetIP {
		if _, ok := dedupIP[ti]; !ok {
			dedupIP[ti] = true
			uniqueIP = append(uniqueIP, ti)
		}
	}

	uniqueNames := []string{}
	for _, tn := range p.TargetName {
		if _, ok := dedupName[tn]; !ok {
			dedupName[tn] = true
			uniqueNames = append(uniqueNames, tn)
		}
	}

	p.TargetIP = uniqueIP
	p.TargetName = uniqueNames

	if len(p.TargetIP) != len(p.TargetName) {
		return fmt.Errorf("please give every prometheus Target a unique IP address and a unique name")
	}

	if p.TargetIP[0] == "" && p.TargetName[0] == "" {
		logp.Info("Start prometheus with no targets")
		p.TargetEmpty = true
		p.Cache = freecache.NewCache(80 * 1024 * 1024)
	}

	p.GaugeMetrics = map[string]prometheus.Gauge{}
	p.GaugeVecMetrics = map[string]*prometheus.GaugeVec{}
	p.CounterVecMetrics = map[string]*prometheus.CounterVec{}

	p.CounterVecMetrics["heplify_method_response"] = prometheus.NewCounterVec(prometheus.CounterOpts{Name: "heplify_method_response", Help: "SIP method and response counter"}, []string{"target_name", "response", "method"})
	p.CounterVecMetrics["heplify_packets_total"] = prometheus.NewCounterVec(prometheus.CounterOpts{Name: "heplify_packets_total", Help: "Total packets by HEP type"}, []string{"type"})
	p.GaugeVecMetrics["heplify_packets_size"] = prometheus.NewGaugeVec(prometheus.GaugeOpts{Name: "heplify_packets_size", Help: "Packet size by HEP type"}, []string{"type"})
	p.GaugeVecMetrics["heplify_xrtp_cs"] = prometheus.NewGaugeVec(prometheus.GaugeOpts{Name: "heplify_xrtp_cs", Help: "XRTP call setup time"}, []string{"target_name"})
	p.GaugeVecMetrics["heplify_xrtp_jir"] = prometheus.NewGaugeVec(prometheus.GaugeOpts{Name: "heplify_xrtp_jir", Help: "XRTP received jitter"}, []string{"target_name"})
	p.GaugeVecMetrics["heplify_xrtp_jis"] = prometheus.NewGaugeVec(prometheus.GaugeOpts{Name: "heplify_xrtp_jis", Help: "XRTP sent jitter"}, []string{"target_name"})
	p.GaugeVecMetrics["heplify_xrtp_plr"] = prometheus.NewGaugeVec(prometheus.GaugeOpts{Name: "heplify_xrtp_plr", Help: "XRTP received packets lost"}, []string{"target_name"})
	p.GaugeVecMetrics["heplify_xrtp_pls"] = prometheus.NewGaugeVec(prometheus.GaugeOpts{Name: "heplify_xrtp_pls", Help: "XRTP sent packets lost"}, []string{"target_name"})
	p.GaugeVecMetrics["heplify_xrtp_dle"] = prometheus.NewGaugeVec(prometheus.GaugeOpts{Name: "heplify_xrtp_dle", Help: "XRTP mean rtt"}, []string{"target_name"})
	p.GaugeVecMetrics["heplify_xrtp_mos"] = prometheus.NewGaugeVec(prometheus.GaugeOpts{Name: "heplify_xrtp_mos", Help: "XRTP mos"}, []string{"target_name"})

	p.GaugeMetrics["heplify_rtcp_fraction_lost"] = prometheus.NewGauge(prometheus.GaugeOpts{Name: "heplify_rtcp_fraction_lost", Help: "RTCP fraction lost"})
	p.GaugeMetrics["heplify_rtcp_packets_lost"] = prometheus.NewGauge(prometheus.GaugeOpts{Name: "heplify_rtcp_packets_lost", Help: "RTCP packets lost"})
	p.GaugeMetrics["heplify_rtcp_jitter"] = prometheus.NewGauge(prometheus.GaugeOpts{Name: "heplify_rtcp_jitter", Help: "RTCP jitter"})
	p.GaugeMetrics["heplify_rtcp_dlsr"] = prometheus.NewGauge(prometheus.GaugeOpts{Name: "heplify_rtcp_dlsr", Help: "RTCP dlsr"})

	p.GaugeMetrics["heplify_rtcpxr_fraction_lost"] = prometheus.NewGauge(prometheus.GaugeOpts{Name: "heplify_rtcpxr_fraction_lost", Help: "RTCPXR fraction lost"})
	p.GaugeMetrics["heplify_rtcpxr_fraction_discard"] = prometheus.NewGauge(prometheus.GaugeOpts{Name: "heplify_rtcpxr_fraction_discard", Help: "RTCPXR fraction discard"})
	p.GaugeMetrics["heplify_rtcpxr_burst_density"] = prometheus.NewGauge(prometheus.GaugeOpts{Name: "heplify_rtcpxr_burst_density", Help: "RTCPXR burst density"})
	p.GaugeMetrics["heplify_rtcpxr_gap_density"] = prometheus.NewGauge(prometheus.GaugeOpts{Name: "heplify_rtcpxr_gap_density", Help: "RTCPXR gap density"})
	p.GaugeMetrics["heplify_rtcpxr_burst_duration"] = prometheus.NewGauge(prometheus.GaugeOpts{Name: "heplify_rtcpxr_burst_duration", Help: "RTCPXR burst duration"})
	p.GaugeMetrics["heplify_rtcpxr_gap_duration"] = prometheus.NewGauge(prometheus.GaugeOpts{Name: "heplify_rtcpxr_gap_duration", Help: "RTCPXR gap duration"})
	p.GaugeMetrics["heplify_rtcpxr_round_trip_delay"] = prometheus.NewGauge(prometheus.GaugeOpts{Name: "heplify_rtcpxr_round_trip_delay", Help: "RTCPXR round trip delay"})
	p.GaugeMetrics["heplify_rtcpxr_end_system_delay"] = prometheus.NewGauge(prometheus.GaugeOpts{Name: "heplify_rtcpxr_end_system_delay", Help: "RTCPXR end system delay"})

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
			p.CounterVecMetrics["heplify_packets_total"].WithLabelValues("sip").Inc()
			p.GaugeVecMetrics["heplify_packets_size"].WithLabelValues("sip").Set(float64(len(pkt.Payload)))
		} else if pkt.ProtoType == 5 {
			p.CounterVecMetrics["heplify_packets_total"].WithLabelValues("rtcp").Inc()
			p.GaugeVecMetrics["heplify_packets_size"].WithLabelValues("rtcp").Set(float64(len(pkt.Payload)))
		} else if pkt.ProtoType == 38 {
			p.CounterVecMetrics["heplify_packets_total"].WithLabelValues("horaclifix").Inc()
			p.GaugeVecMetrics["heplify_packets_size"].WithLabelValues("horaclifix").Set(float64(len(pkt.Payload)))
		} else if pkt.ProtoType == 100 {
			p.CounterVecMetrics["heplify_packets_total"].WithLabelValues("log").Inc()
			p.GaugeVecMetrics["heplify_packets_size"].WithLabelValues("log").Set(float64(len(pkt.Payload)))
		}

		if pkt.SIP != nil && pkt.ProtoType == 1 {
			if !p.TargetEmpty {
				for k, tn := range p.TargetName {
					if pkt.SrcIP == p.TargetIP[k] || pkt.DstIP == p.TargetIP[k] {
						p.CounterVecMetrics["heplify_method_response"].WithLabelValues(tn, pkt.SIP.StartLine.Method, pkt.SIP.CseqMethod).Inc()

						if pkt.SIP.RTPStatVal != "" {
							p.dissectXRTPStats(tn, pkt.SIP.RTPStatVal)
						}
					}
				}
			} else {
				_, err := p.Cache.Get([]byte(pkt.SIP.CallID + pkt.SIP.StartLine.Method + pkt.SIP.CseqMethod))
				if err == nil {
					continue
				}
				err = p.Cache.Set([]byte(pkt.SIP.CallID+pkt.SIP.StartLine.Method+pkt.SIP.CseqMethod), nil, 600)
				if err != nil {
					logp.Warn("%v", err)
				}

				p.CounterVecMetrics["heplify_method_response"].WithLabelValues("", pkt.SIP.StartLine.Method, pkt.SIP.CseqMethod).Inc()

				if pkt.SIP.RTPStatVal != "" {
					p.dissectXRTPStats("", pkt.SIP.RTPStatVal)
				}
			}
		} else if pkt.ProtoType == 5 {
			p.dissectRTCPStats([]byte(pkt.Payload))
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
				p.GaugeVecMetrics["heplify_xrtp_cs"].WithLabelValues(tn).Set(float64(cs / 1000))
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
					p.GaugeVecMetrics["heplify_xrtp_plr"].WithLabelValues(tn).Set(float64(plr))
				} else {
					logp.Err("%v", err)
				}
				pls, err = strconv.Atoi(pl[1])
				if err == nil {
					p.GaugeVecMetrics["heplify_xrtp_pls"].WithLabelValues(tn).Set(float64(pls))
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
					p.GaugeVecMetrics["heplify_xrtp_jir"].WithLabelValues(tn).Set(float64(jir))
				} else {
					logp.Err("%v", err)
				}
				jis, err = strconv.Atoi(ji[1])
				if err == nil {
					p.GaugeVecMetrics["heplify_xrtp_jis"].WithLabelValues(tn).Set(float64(jis))
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
					p.GaugeVecMetrics["heplify_xrtp_dle"].WithLabelValues(tn).Set(float64(dle))
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
	p.GaugeVecMetrics["heplify_xrtp_mos"].WithLabelValues(tn).Set(mos)

}

func (p *Prometheus) dissectRTCPStats(data []byte) {
	jsonparser.EachKey(data, func(idx int, value []byte, vt jsonparser.ValueType, err error) {
		switch idx {
		case 0:
			if fraction_lost, err := jsonparser.ParseFloat(value); err == nil {
				p.GaugeMetrics["heplify_rtcp_fraction_lost"].Set(normMax(fraction_lost))
			}
		case 1:
			if packets_lost, err := jsonparser.ParseFloat(value); err == nil {
				p.GaugeMetrics["heplify_rtcp_packets_lost"].Set(normMax(packets_lost))
			}
		case 2:
			if ia_jitter, err := jsonparser.ParseFloat(value); err == nil {
				p.GaugeMetrics["heplify_rtcp_jitter"].Set(normMax(ia_jitter))
			}
		case 3:
			if dlsr, err := jsonparser.ParseFloat(value); err == nil {
				p.GaugeMetrics["heplify_rtcp_dlsr"].Set(normMax(dlsr))
			}
		case 4:
			if fraction_lost, err := jsonparser.ParseFloat(value); err == nil {
				p.GaugeMetrics["heplify_rtcpxr_fraction_lost"].Set(fraction_lost)
			}
		case 5:
			if fraction_discard, err := jsonparser.ParseFloat(value); err == nil {
				p.GaugeMetrics["heplify_rtcpxr_fraction_discard"].Set(fraction_discard)
			}
		case 6:
			if burst_density, err := jsonparser.ParseFloat(value); err == nil {
				p.GaugeMetrics["heplify_rtcpxr_burst_density"].Set(burst_density)
			}
		case 7:
			if gap_density, err := jsonparser.ParseFloat(value); err == nil {
				p.GaugeMetrics["heplify_rtcpxr_gap_density"].Set(gap_density)
			}
		case 8:
			if burst_duration, err := jsonparser.ParseFloat(value); err == nil {
				p.GaugeMetrics["heplify_rtcpxr_burst_duration"].Set(burst_duration)
			}
		case 9:
			if gap_duration, err := jsonparser.ParseFloat(value); err == nil {
				p.GaugeMetrics["heplify_rtcpxr_gap_duration"].Set(gap_duration)
			}
		case 10:
			if round_trip_delay, err := jsonparser.ParseFloat(value); err == nil {
				p.GaugeMetrics["heplify_rtcpxr_round_trip_delay"].Set(round_trip_delay)
			}
		case 11:
			if end_system_delay, err := jsonparser.ParseFloat(value); err == nil {
				p.GaugeMetrics["heplify_rtcpxr_end_system_delay"].Set(end_system_delay)
			}
		}
	}, paths...)
}

var paths = [][]string{
	[]string{"report_blocks", "[0]", "fraction_lost"},
	[]string{"report_blocks", "[0]", "packets_lost"},
	[]string{"report_blocks", "[0]", "ia_jitter"},
	[]string{"report_blocks", "[0]", "dlsr"},
	[]string{"report_blocks_xr", "fraction_lost"},
	[]string{"report_blocks_xr", "fraction_discard"},
	[]string{"report_blocks_xr", "burst_density"},
	[]string{"report_blocks_xr", "gap_density"},
	[]string{"report_blocks_xr", "burst_duration"},
	[]string{"report_blocks_xr", "gap_duration"},
	[]string{"report_blocks_xr", "round_trip_delay"},
	[]string{"report_blocks_xr", "end_system_delay"},
}

func cutSpace(str string) string {
	return strings.Map(func(r rune) rune {
		if unicode.IsSpace(r) {
			return -1
		}
		return r
	}, str)
}

func normMax(val float64) float64 {
	if val > 10000000 {
		return 0
	}
	return val
}
