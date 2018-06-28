package metric

import (
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"sync"
	"syscall"

	"github.com/buger/jsonparser"
	"github.com/coocood/freecache"
	"github.com/negbie/heplify-server"
	"github.com/negbie/heplify-server/config"
	"github.com/negbie/logp"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

type Prometheus struct {
	TargetIP          []string
	TargetName        []string
	TargetEmpty       bool
	TargetConf        *sync.RWMutex
	CvMethodResponse  *prometheus.CounterVec
	CvPacketsTotal    *prometheus.CounterVec
	GvPacketsSize     *prometheus.GaugeVec
	GaugeMetrics      map[string]prometheus.Gauge
	GaugeVecMetrics   map[string]*prometheus.GaugeVec
	CounterVecMetrics map[string]*prometheus.CounterVec
	Cache             *freecache.Cache
	horaclifixPaths   [][]string
	rtpPaths          [][]string
	rtcpPaths         [][]string
}

func (p *Prometheus) setup() (err error) {
	p.TargetConf = new(sync.RWMutex)
	p.TargetIP = strings.Split(cutSpace(config.Setting.PromTargetIP), ",")
	p.TargetName = strings.Split(cutSpace(config.Setting.PromTargetName), ",")

	s := make(chan os.Signal, 1)
	signal.Notify(s, syscall.SIGHUP)
	go func() {
		for {
			<-s
			p.loadPromConf()
		}
	}()

	if len(p.TargetIP) == len(p.TargetName) && p.TargetIP != nil && p.TargetName != nil {
		if len(p.TargetIP[0]) == 0 || len(p.TargetName[0]) == 0 {
			logp.Info("expose metrics with no or unbalanced targets")
			p.TargetIP[0] = ""
			p.TargetName[0] = ""
			p.TargetEmpty = true
			p.Cache = freecache.NewCache(60 * 1024 * 1024)
		} else {
			logp.Info("start prometheus with PromTargetIP: %#v", p.TargetIP)
			logp.Info("start prometheus with PromTargetName: %#v", p.TargetName)
		}
	} else {
		logp.Info("please give every PromTargetIP a unique IP and PromTargetName a unique name")
		return fmt.Errorf("faulty PromTargetIP or PromTargetName")
	}

	p.GaugeMetrics = map[string]prometheus.Gauge{}
	p.GaugeVecMetrics = map[string]*prometheus.GaugeVec{}
	p.CounterVecMetrics = map[string]*prometheus.CounterVec{}

	p.CvMethodResponse = prometheus.NewCounterVec(prometheus.CounterOpts{Name: "heplify_method_response", Help: "SIP method and response counter"}, []string{"target_name", "direction", "response", "method"})
	p.CvPacketsTotal = prometheus.NewCounterVec(prometheus.CounterOpts{Name: "heplify_packets_total", Help: "Total packets by HEP type"}, []string{"type"})
	p.GvPacketsSize = prometheus.NewGaugeVec(prometheus.GaugeOpts{Name: "heplify_packets_size", Help: "Packet size by HEP type"}, []string{"type"})
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

	if config.Setting.RTPAgentStats {
		p.GaugeMetrics["heplify_rtpagent_delta"] = prometheus.NewGauge(prometheus.GaugeOpts{Name: "heplify_rtpagent_delta", Help: "RTPAgent delta"})
		p.GaugeMetrics["heplify_rtpagent_jitter"] = prometheus.NewGauge(prometheus.GaugeOpts{Name: "heplify_rtpagent_jitter", Help: "RTPAgent jitter"})
		p.GaugeMetrics["heplify_rtpagent_mos"] = prometheus.NewGauge(prometheus.GaugeOpts{Name: "heplify_rtpagent_mos", Help: "RTPAgent mos"})
		p.GaugeMetrics["heplify_rtpagent_packets_lost"] = prometheus.NewGauge(prometheus.GaugeOpts{Name: "heplify_rtpagent_packets_lost", Help: "RTPAgent packets lost"})
		p.GaugeMetrics["heplify_rtpagent_rfactor"] = prometheus.NewGauge(prometheus.GaugeOpts{Name: "heplify_rtpagent_rfactor", Help: "RTPAgent rfactor"})
		p.GaugeMetrics["heplify_rtpagent_skew"] = prometheus.NewGauge(prometheus.GaugeOpts{Name: "heplify_rtpagent_skew", Help: "RTPAgent skew"})

		p.rtpPaths = [][]string{
			[]string{"DELTA"},
			[]string{"JITTER"},
			[]string{"MOS"},
			[]string{"PACKET_LOSS"},
			[]string{"RFACTOR"},
			[]string{"SKEW"},
		}
	}

	if config.Setting.HoraclifixStats {
		p.GaugeVecMetrics["horaclifix_rtp_mos"] = prometheus.NewGaugeVec(prometheus.GaugeOpts{Name: "horaclifix_rtp_mos", Help: "Incoming RTP MOS"}, []string{"sbc_name", "direction", "inc_realm", "out_realm"})
		p.GaugeVecMetrics["horaclifix_rtp_rval"] = prometheus.NewGaugeVec(prometheus.GaugeOpts{Name: "horaclifix_rtp_rval", Help: "Incoming RTP rVal"}, []string{"sbc_name", "direction", "inc_realm", "out_realm"})
		p.GaugeVecMetrics["horaclifix_rtp_packets"] = prometheus.NewGaugeVec(prometheus.GaugeOpts{Name: "horaclifix_rtp_packets", Help: "Incoming RTP packets"}, []string{"sbc_name", "direction", "inc_realm", "out_realm"})
		p.GaugeVecMetrics["horaclifix_rtp_lost_packets"] = prometheus.NewGaugeVec(prometheus.GaugeOpts{Name: "horaclifix_rtp_lost_packets", Help: "Incoming RTP lostPackets"}, []string{"sbc_name", "direction", "inc_realm", "out_realm"})
		p.GaugeVecMetrics["horaclifix_rtp_avg_jitter"] = prometheus.NewGaugeVec(prometheus.GaugeOpts{Name: "horaclifix_rtp_avg_jitter", Help: "Incoming RTP avgJitter"}, []string{"sbc_name", "direction", "inc_realm", "out_realm"})
		p.GaugeVecMetrics["horaclifix_rtp_max_jitter"] = prometheus.NewGaugeVec(prometheus.GaugeOpts{Name: "horaclifix_rtp_max_jitter", Help: "Incoming RTP maxJitter"}, []string{"sbc_name", "direction", "inc_realm", "out_realm"})
		p.GaugeVecMetrics["horaclifix_rtcp_packets"] = prometheus.NewGaugeVec(prometheus.GaugeOpts{Name: "horaclifix_rtcp_packets", Help: "Incoming RTCP packets"}, []string{"sbc_name", "direction", "inc_realm", "out_realm"})
		p.GaugeVecMetrics["horaclifix_rtcp_lost_packets"] = prometheus.NewGaugeVec(prometheus.GaugeOpts{Name: "horaclifix_rtcp_lost_packets", Help: "Incoming RTCP lostPackets"}, []string{"sbc_name", "direction", "inc_realm", "out_realm"})
		p.GaugeVecMetrics["horaclifix_rtcp_avg_jitter"] = prometheus.NewGaugeVec(prometheus.GaugeOpts{Name: "horaclifix_rtcp_avg_jitter", Help: "Incoming RTCP avgJitter"}, []string{"sbc_name", "direction", "inc_realm", "out_realm"})
		p.GaugeVecMetrics["horaclifix_rtcp_max_jitter"] = prometheus.NewGaugeVec(prometheus.GaugeOpts{Name: "horaclifix_rtcp_max_jitter", Help: "Incoming RTCP maxJitter"}, []string{"sbc_name", "direction", "inc_realm", "out_realm"})
		p.GaugeVecMetrics["horaclifix_rtcp_avg_lat"] = prometheus.NewGaugeVec(prometheus.GaugeOpts{Name: "horaclifix_rtcp_avg_lat", Help: "Incoming RTCP avgLat"}, []string{"sbc_name", "direction", "inc_realm", "out_realm"})
		p.GaugeVecMetrics["horaclifix_rtcp_max_lat"] = prometheus.NewGaugeVec(prometheus.GaugeOpts{Name: "horaclifix_rtcp_max_lat", Help: "Incoming RTCP maxLat"}, []string{"sbc_name", "direction", "inc_realm", "out_realm"})

		p.horaclifixPaths = [][]string{
			[]string{"NAME"},
			[]string{"INC_REALM"},
			[]string{"OUT_REALM"},
			//[]string{"INC_ID"},
			//[]string{"OUT_ID"},
			[]string{"INC_MOS"},
			[]string{"INC_RVAL"},
			//[]string{"INC_RTP_BYTE"},
			[]string{"INC_RTP_PK"},
			[]string{"INC_RTP_PK_LOSS"},
			[]string{"INC_RTP_AVG_JITTER"},
			[]string{"INC_RTP_MAX_JITTER"},
			//[]string{"INC_RTCP_BYTE"},
			[]string{"INC_RTCP_PK"},
			[]string{"INC_RTCP_PK_LOSS"},
			[]string{"INC_RTCP_AVG_JITTER"},
			[]string{"INC_RTCP_MAX_JITTER"},
			[]string{"INC_RTCP_AVG_LAT"},
			[]string{"INC_RTCP_MAX_LAT"},
			//[]string{"CALLER_VLAN"},
			//[]string{"CALLER_SRC_IP"},
			//[]string{"CALLER_SRC_PORT"},
			//[]string{"CALLER_DST_IP"},
			//[]string{"CALLER_DST_PORT"},
			[]string{"OUT_MOS"},
			[]string{"OUT_RVAL"},
			//[]string{"OUT_RTP_BYTE"},
			[]string{"OUT_RTP_PK"},
			[]string{"OUT_RTP_PK_LOSS"},
			[]string{"OUT_RTP_AVG_JITTER"},
			[]string{"OUT_RTP_MAX_JITTER"},
			//[]string{"OUT_RTCP_BYTE"},
			[]string{"OUT_RTCP_PK"},
			[]string{"OUT_RTCP_PK_LOSS"},
			[]string{"OUT_RTCP_AVG_JITTER"},
			[]string{"OUT_RTCP_MAX_JITTER"},
			[]string{"OUT_RTCP_AVG_LAT"},
			[]string{"OUT_RTCP_MAX_LAT"},
			//[]string{"CALLEE_VLAN"},
			//[]string{"CALLEE_SRC_IP"},
			//[]string{"CALLEE_SRC_PORT"},
			//[]string{"CALLEE_DST_IP"},
			//[]string{"CALLEE_DST_PORT"},
			//[]string{"MEDIA_TYPE"},
		}
	}

	p.rtcpPaths = [][]string{
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

	prometheus.MustRegister(p.CvMethodResponse)
	prometheus.MustRegister(p.CvPacketsTotal)
	prometheus.MustRegister(p.GvPacketsSize)

	for k := range p.GaugeMetrics {
		prometheus.MustRegister(p.GaugeMetrics[k])
	}
	for k := range p.GaugeVecMetrics {
		prometheus.MustRegister(p.GaugeVecMetrics[k])
	}
	for k := range p.CounterVecMetrics {
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
		pkt       *decoder.HEP
		ok        bool
		direction string
	)

	for {
		select {
		case pkt, ok = <-hCh:
			if !ok {
				break
			}

			if pkt.ProtoType == 1 {
				p.CvPacketsTotal.WithLabelValues("sip").Inc()
				p.GvPacketsSize.WithLabelValues("sip").Set(float64(len(pkt.Payload)))
			} else if pkt.ProtoType == 5 {
				p.CvPacketsTotal.WithLabelValues("rtcp").Inc()
				p.GvPacketsSize.WithLabelValues("rtcp").Set(float64(len(pkt.Payload)))
			} else if pkt.ProtoType == 38 {
				p.CvPacketsTotal.WithLabelValues("horaclifix").Inc()
				p.GvPacketsSize.WithLabelValues("horaclifix").Set(float64(len(pkt.Payload)))
			} else if pkt.ProtoType == 100 {
				p.CvPacketsTotal.WithLabelValues("log").Inc()
				p.GvPacketsSize.WithLabelValues("log").Set(float64(len(pkt.Payload)))
			} else {
				pt := strconv.Itoa(int(pkt.ProtoType))
				p.CvPacketsTotal.WithLabelValues(pt).Inc()
				p.GvPacketsSize.WithLabelValues(pt).Set(float64(len(pkt.Payload)))
			}

			if pkt.SIP != nil && pkt.ProtoType == 1 {
				if !p.TargetEmpty {
					for k, tn := range p.TargetName {
						if pkt.SrcIP == p.TargetIP[k] || pkt.DstIP == p.TargetIP[k] {
							direction = "src"
							if pkt.DstIP == p.TargetIP[k] {
								direction = "dst"
							}
							p.CvMethodResponse.WithLabelValues(tn, direction, pkt.SIP.StartLine.Method, pkt.SIP.CseqMethod).Inc()

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

					p.CvMethodResponse.WithLabelValues("", "", pkt.SIP.StartLine.Method, pkt.SIP.CseqMethod).Inc()

					if pkt.SIP.RTPStatVal != "" {
						p.dissectXRTPStats("", pkt.SIP.RTPStatVal)
					}
				}
			} else if pkt.ProtoType == 5 {
				p.dissectRTCPStats([]byte(pkt.Payload))
			} else if pkt.ProtoType == 34 && config.Setting.RTPAgentStats {
				p.dissectRTPStats([]byte(pkt.Payload))
			} else if pkt.ProtoType == 38 && config.Setting.HoraclifixStats {
				p.dissectHoraclifixStats([]byte(pkt.Payload))
			}
		}
	}
}

func (p *Prometheus) dissectXRTPStats(tn, stats string) {
	var err error
	cs, pr, ps, plr, pls, jir, jis, dle, r, mos := 0, 0, 0, 0, 0, 0, 0, 0, 0.0, 0.0

	p.CvPacketsTotal.WithLabelValues("xrtp").Inc()
	p.GvPacketsSize.WithLabelValues("xrtp").Set(float64(len(stats)))

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
	if mos < 1 || mos > 5 {
		mos = 1
	}
	p.GaugeVecMetrics["heplify_xrtp_mos"].WithLabelValues(tn).Set(mos)

}

func (p *Prometheus) dissectRTCPStats(data []byte) {
	jsonparser.EachKey(data, func(idx int, value []byte, vt jsonparser.ValueType, err error) {
		switch idx {
		case 0:
			if fractionLost, err := jsonparser.ParseFloat(value); err == nil {
				p.GaugeMetrics["heplify_rtcp_fraction_lost"].Set(normMax(fractionLost))
			}
		case 1:
			if packetsLost, err := jsonparser.ParseFloat(value); err == nil {
				p.GaugeMetrics["heplify_rtcp_packets_lost"].Set(normMax(packetsLost))
			}
		case 2:
			if iaJitter, err := jsonparser.ParseFloat(value); err == nil {
				p.GaugeMetrics["heplify_rtcp_jitter"].Set(normMax(iaJitter))
			}
		case 3:
			if dlsr, err := jsonparser.ParseFloat(value); err == nil {
				p.GaugeMetrics["heplify_rtcp_dlsr"].Set(normMax(dlsr))
			}
		case 4:
			if fractionLost, err := jsonparser.ParseFloat(value); err == nil {
				p.GaugeMetrics["heplify_rtcpxr_fraction_lost"].Set(fractionLost)
			}
		case 5:
			if fractionDiscard, err := jsonparser.ParseFloat(value); err == nil {
				p.GaugeMetrics["heplify_rtcpxr_fraction_discard"].Set(fractionDiscard)
			}
		case 6:
			if burstDensity, err := jsonparser.ParseFloat(value); err == nil {
				p.GaugeMetrics["heplify_rtcpxr_burst_density"].Set(burstDensity)
			}
		case 7:
			if gapDensity, err := jsonparser.ParseFloat(value); err == nil {
				p.GaugeMetrics["heplify_rtcpxr_gap_density"].Set(gapDensity)
			}
		case 8:
			if burstDuration, err := jsonparser.ParseFloat(value); err == nil {
				p.GaugeMetrics["heplify_rtcpxr_burst_duration"].Set(burstDuration)
			}
		case 9:
			if gapDuration, err := jsonparser.ParseFloat(value); err == nil {
				p.GaugeMetrics["heplify_rtcpxr_gap_duration"].Set(gapDuration)
			}
		case 10:
			if roundTripDelay, err := jsonparser.ParseFloat(value); err == nil {
				p.GaugeMetrics["heplify_rtcpxr_round_trip_delay"].Set(roundTripDelay)
			}
		case 11:
			if endSystemDelay, err := jsonparser.ParseFloat(value); err == nil {
				p.GaugeMetrics["heplify_rtcpxr_end_system_delay"].Set(endSystemDelay)
			}
		}
	}, p.rtcpPaths...)
}

func (p *Prometheus) dissectRTPStats(data []byte) {
	jsonparser.EachKey(data, func(idx int, value []byte, vt jsonparser.ValueType, err error) {
		switch idx {
		case 0:
			if delta, err := jsonparser.ParseFloat(value); err == nil {
				p.GaugeMetrics["heplify_rtpagent_delta"].Set(delta)
			}
		case 1:
			if iaJitter, err := jsonparser.ParseFloat(value); err == nil {
				p.GaugeMetrics["heplify_rtpagent_jitter"].Set(iaJitter)
			}
		case 2:
			if mos, err := jsonparser.ParseFloat(value); err == nil {
				p.GaugeMetrics["heplify_rtpagent_mos"].Set(mos)
			}
		case 3:
			if packetsLost, err := jsonparser.ParseFloat(value); err == nil {
				p.GaugeMetrics["heplify_rtpagent_packets_lost"].Set(packetsLost)
			}
		case 4:
			if rfactor, err := jsonparser.ParseFloat(value); err == nil {
				p.GaugeMetrics["heplify_rtpagent_rfactor"].Set(rfactor)
			}
		case 5:
			if skew, err := jsonparser.ParseFloat(value); err == nil {
				p.GaugeMetrics["heplify_rtpagent_skew"].Set(skew)
			}
		}
	}, p.rtpPaths...)
}

func (p *Prometheus) dissectHoraclifixStats(data []byte) {
	var sbcName, incRealm, outRealm string

	jsonparser.EachKey(data, func(idx int, value []byte, vt jsonparser.ValueType, err error) {
		switch idx {
		case 0:
			if sbcName, err = jsonparser.ParseString(value); err != nil {
				logp.Warn("could not decode sbcName %s from horaclifix report", string(sbcName))
				return
			}
		case 1:
			if incRealm, err = jsonparser.ParseString(value); err != nil {
				logp.Warn("could not decode incRealm %s from horaclifix report", string(incRealm))
				return
			}
		case 2:
			if outRealm, err = jsonparser.ParseString(value); err != nil {
				logp.Warn("could not decode outRealm %s from horaclifix report", string(outRealm))
				return
			}
		case 3:
			if incMos, err := jsonparser.ParseFloat(value); err == nil {
				p.GaugeVecMetrics["horaclifix_rtp_mos"].WithLabelValues(sbcName, "inc", incRealm, outRealm).Set(incMos / 100)
			}
		case 4:
			if incRval, err := jsonparser.ParseFloat(value); err == nil {
				p.GaugeVecMetrics["horaclifix_rtp_rval"].WithLabelValues(sbcName, "inc", incRealm, outRealm).Set(incRval / 100)
			}
		case 5:
			if incRtpPackets, err := jsonparser.ParseFloat(value); err == nil {
				p.GaugeVecMetrics["horaclifix_rtp_packets"].WithLabelValues(sbcName, "inc", incRealm, outRealm).Set(incRtpPackets)
			}
		case 6:
			if incRtpLostPackets, err := jsonparser.ParseFloat(value); err == nil {
				p.GaugeVecMetrics["horaclifix_rtp_lost_packets"].WithLabelValues(sbcName, "inc", incRealm, outRealm).Set(incRtpLostPackets)
			}
		case 7:
			if incRtpAvgJitter, err := jsonparser.ParseFloat(value); err == nil {
				p.GaugeVecMetrics["horaclifix_rtp_avg_jitter"].WithLabelValues(sbcName, "inc", incRealm, outRealm).Set(incRtpAvgJitter)
			}
		case 8:
			if incRtpMaxJitter, err := jsonparser.ParseFloat(value); err == nil {
				p.GaugeVecMetrics["horaclifix_rtp_max_jitter"].WithLabelValues(sbcName, "inc", incRealm, outRealm).Set(incRtpMaxJitter)
			}
		case 9:
			if incRtcpPackets, err := jsonparser.ParseFloat(value); err == nil {
				p.GaugeVecMetrics["horaclifix_rtcp_packets"].WithLabelValues(sbcName, "inc", incRealm, outRealm).Set(incRtcpPackets)
			}
		case 10:
			if incRtcpLostPackets, err := jsonparser.ParseFloat(value); err == nil {
				p.GaugeVecMetrics["horaclifix_rtcp_lost_packets"].WithLabelValues(sbcName, "inc", incRealm, outRealm).Set(incRtcpLostPackets)
			}
		case 11:
			if incRtcpAvgJitter, err := jsonparser.ParseFloat(value); err == nil {
				p.GaugeVecMetrics["horaclifix_rtcp_avg_jitter"].WithLabelValues(sbcName, "inc", incRealm, outRealm).Set(incRtcpAvgJitter)
			}
		case 12:
			if incRtcpMaxJitter, err := jsonparser.ParseFloat(value); err == nil {
				p.GaugeVecMetrics["horaclifix_rtcp_max_jitter"].WithLabelValues(sbcName, "inc", incRealm, outRealm).Set(incRtcpMaxJitter)
			}
		case 13:
			if incRtcpAvgLat, err := jsonparser.ParseFloat(value); err == nil {
				p.GaugeVecMetrics["horaclifix_rtcp_avg_lat"].WithLabelValues(sbcName, "inc", incRealm, outRealm).Set(incRtcpAvgLat)
			}
		case 14:
			if incRtcpMaxLat, err := jsonparser.ParseFloat(value); err == nil {
				p.GaugeVecMetrics["horaclifix_rtcp_max_lat"].WithLabelValues(sbcName, "inc", incRealm, outRealm).Set(incRtcpMaxLat)
			}
		case 15:
			if outMos, err := jsonparser.ParseFloat(value); err == nil {
				p.GaugeVecMetrics["horaclifix_rtp_mos"].WithLabelValues(sbcName, "out", incRealm, outRealm).Set(outMos / 100)
			}
		case 16:
			if outRval, err := jsonparser.ParseFloat(value); err == nil {
				p.GaugeVecMetrics["horaclifix_rtp_rval"].WithLabelValues(sbcName, "out", incRealm, outRealm).Set(outRval / 100)
			}
		case 17:
			if outRtpPackets, err := jsonparser.ParseFloat(value); err == nil {
				p.GaugeVecMetrics["horaclifix_rtp_packets"].WithLabelValues(sbcName, "out", incRealm, outRealm).Set(outRtpPackets)
			}
		case 18:
			if outRtpLostPackets, err := jsonparser.ParseFloat(value); err == nil {
				p.GaugeVecMetrics["horaclifix_rtp_lost_packets"].WithLabelValues(sbcName, "out", incRealm, outRealm).Set(outRtpLostPackets)
			}
		case 19:
			if outRtpAvgJitter, err := jsonparser.ParseFloat(value); err == nil {
				p.GaugeVecMetrics["horaclifix_rtp_avg_jitter"].WithLabelValues(sbcName, "out", incRealm, outRealm).Set(outRtpAvgJitter)
			}
		case 20:
			if outRtpMaxJitter, err := jsonparser.ParseFloat(value); err == nil {
				p.GaugeVecMetrics["horaclifix_rtp_max_jitter"].WithLabelValues(sbcName, "out", incRealm, outRealm).Set(outRtpMaxJitter)
			}
		case 21:
			if outRtcpPackets, err := jsonparser.ParseFloat(value); err == nil {
				p.GaugeVecMetrics["horaclifix_rtcp_packets"].WithLabelValues(sbcName, "out", incRealm, outRealm).Set(outRtcpPackets)
			}
		case 22:
			if outRtcpLostPackets, err := jsonparser.ParseFloat(value); err == nil {
				p.GaugeVecMetrics["horaclifix_rtcp_lost_packets"].WithLabelValues(sbcName, "out", incRealm, outRealm).Set(outRtcpLostPackets)
			}
		case 23:
			if outRtcpAvgJitter, err := jsonparser.ParseFloat(value); err == nil {
				p.GaugeVecMetrics["horaclifix_rtcp_avg_jitter"].WithLabelValues(sbcName, "out", incRealm, outRealm).Set(outRtcpAvgJitter)
			}
		case 24:
			if outRtcpMaxJitter, err := jsonparser.ParseFloat(value); err == nil {
				p.GaugeVecMetrics["horaclifix_rtcp_max_jitter"].WithLabelValues(sbcName, "out", incRealm, outRealm).Set(outRtcpMaxJitter)
			}
		case 25:
			if outRtcpAvgLat, err := jsonparser.ParseFloat(value); err == nil {
				p.GaugeVecMetrics["horaclifix_rtcp_avg_lat"].WithLabelValues(sbcName, "out", incRealm, outRealm).Set(outRtcpAvgLat)
			}
		case 26:
			if outRtcpMaxLat, err := jsonparser.ParseFloat(value); err == nil {
				p.GaugeVecMetrics["horaclifix_rtcp_max_lat"].WithLabelValues(sbcName, "out", incRealm, outRealm).Set(outRtcpMaxLat)
			}
		}
	}, p.horaclifixPaths...)
}
