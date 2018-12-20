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

	"github.com/coocood/freecache"
	"github.com/negbie/heplify-server"
	"github.com/negbie/heplify-server/config"
	"github.com/negbie/logp"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	packetsByType = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "heplify_packets_total",
		Help: "Total packets by HEP type"},
		[]string{"type"})
	packetsBySize = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "heplify_packets_size",
		Help: "Packet size by HEP type"},
		[]string{"type"})
	methodResponses = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "heplify_method_response",
		Help: "SIP method and response counter"},
		[]string{"target_name", "direction", "node_id", "response", "method"})
	reasonCause = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "heplify_reason_isup_total",
		Help: "ISUP Q.850 cause from reason header"},
		[]string{"cause", "node_id"})
)

type Prometheus struct {
	TargetEmpty       bool
	TargetIP          []string
	TargetName        []string
	TargetMap         map[string]string
	TargetConf        *sync.RWMutex
	Cache             *freecache.Cache
	GaugeVecMetrics   map[string]*prometheus.GaugeVec
	CounterVecMetrics map[string]*prometheus.CounterVec
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
			logp.Info("expose metrics without or unbalanced targets")
			p.TargetIP[0] = ""
			p.TargetName[0] = ""
			p.TargetEmpty = true
			p.Cache = freecache.NewCache(60 * 1024 * 1024)
		} else {
			for i := range p.TargetName {
				logp.Info("prometheus tag assignment %d: %s -> %s", i+1, p.TargetIP[i], p.TargetName[i])
			}
			p.TargetMap = make(map[string]string)
			for i := 0; i < len(p.TargetName); i++ {
				p.TargetMap[p.TargetIP[i]] = p.TargetName[i]
			}
		}
	} else {
		logp.Info("please give every PromTargetIP a unique IP and PromTargetName a unique name")
		return fmt.Errorf("faulty PromTargetIP or PromTargetName")
	}

	p.GaugeVecMetrics = map[string]*prometheus.GaugeVec{}
	p.CounterVecMetrics = map[string]*prometheus.CounterVec{}

	p.GaugeVecMetrics["heplify_xrtp_cs"] = prometheus.NewGaugeVec(prometheus.GaugeOpts{Name: "heplify_xrtp_cs", Help: "XRTP call setup time"}, []string{"target_name"})
	p.GaugeVecMetrics["heplify_xrtp_jir"] = prometheus.NewGaugeVec(prometheus.GaugeOpts{Name: "heplify_xrtp_jir", Help: "XRTP received jitter"}, []string{"target_name"})
	p.GaugeVecMetrics["heplify_xrtp_jis"] = prometheus.NewGaugeVec(prometheus.GaugeOpts{Name: "heplify_xrtp_jis", Help: "XRTP sent jitter"}, []string{"target_name"})
	p.GaugeVecMetrics["heplify_xrtp_plr"] = prometheus.NewGaugeVec(prometheus.GaugeOpts{Name: "heplify_xrtp_plr", Help: "XRTP received packets lost"}, []string{"target_name"})
	p.GaugeVecMetrics["heplify_xrtp_pls"] = prometheus.NewGaugeVec(prometheus.GaugeOpts{Name: "heplify_xrtp_pls", Help: "XRTP sent packets lost"}, []string{"target_name"})
	p.GaugeVecMetrics["heplify_xrtp_dle"] = prometheus.NewGaugeVec(prometheus.GaugeOpts{Name: "heplify_xrtp_dle", Help: "XRTP mean rtt"}, []string{"target_name"})
	p.GaugeVecMetrics["heplify_xrtp_mos"] = prometheus.NewGaugeVec(prometheus.GaugeOpts{Name: "heplify_xrtp_mos", Help: "XRTP mos"}, []string{"target_name"})

	p.GaugeVecMetrics["heplify_rtcp_fraction_lost"] = prometheus.NewGaugeVec(prometheus.GaugeOpts{Name: "heplify_rtcp_fraction_lost", Help: "RTCP fraction lost"}, []string{"node_id"})
	p.GaugeVecMetrics["heplify_rtcp_packets_lost"] = prometheus.NewGaugeVec(prometheus.GaugeOpts{Name: "heplify_rtcp_packets_lost", Help: "RTCP packets lost"}, []string{"node_id"})
	p.GaugeVecMetrics["heplify_rtcp_jitter"] = prometheus.NewGaugeVec(prometheus.GaugeOpts{Name: "heplify_rtcp_jitter", Help: "RTCP jitter"}, []string{"node_id"})
	p.GaugeVecMetrics["heplify_rtcp_dlsr"] = prometheus.NewGaugeVec(prometheus.GaugeOpts{Name: "heplify_rtcp_dlsr", Help: "RTCP dlsr"}, []string{"node_id"})

	p.GaugeVecMetrics["heplify_rtcpxr_fraction_lost"] = prometheus.NewGaugeVec(prometheus.GaugeOpts{Name: "heplify_rtcpxr_fraction_lost", Help: "RTCPXR fraction lost"}, []string{"node_id"})
	p.GaugeVecMetrics["heplify_rtcpxr_fraction_discard"] = prometheus.NewGaugeVec(prometheus.GaugeOpts{Name: "heplify_rtcpxr_fraction_discard", Help: "RTCPXR fraction discard"}, []string{"node_id"})
	p.GaugeVecMetrics["heplify_rtcpxr_burst_density"] = prometheus.NewGaugeVec(prometheus.GaugeOpts{Name: "heplify_rtcpxr_burst_density", Help: "RTCPXR burst density"}, []string{"node_id"})
	p.GaugeVecMetrics["heplify_rtcpxr_burst_duration"] = prometheus.NewGaugeVec(prometheus.GaugeOpts{Name: "heplify_rtcpxr_burst_duration", Help: "RTCPXR burst duration"}, []string{"node_id"})
	p.GaugeVecMetrics["heplify_rtcpxr_gap_density"] = prometheus.NewGaugeVec(prometheus.GaugeOpts{Name: "heplify_rtcpxr_gap_density", Help: "RTCPXR gap density"}, []string{"node_id"})
	p.GaugeVecMetrics["heplify_rtcpxr_gap_duration"] = prometheus.NewGaugeVec(prometheus.GaugeOpts{Name: "heplify_rtcpxr_gap_duration", Help: "RTCPXR gap duration"}, []string{"node_id"})
	p.GaugeVecMetrics["heplify_rtcpxr_round_trip_delay"] = prometheus.NewGaugeVec(prometheus.GaugeOpts{Name: "heplify_rtcpxr_round_trip_delay", Help: "RTCPXR round trip delay"}, []string{"node_id"})
	p.GaugeVecMetrics["heplify_rtcpxr_end_system_delay"] = prometheus.NewGaugeVec(prometheus.GaugeOpts{Name: "heplify_rtcpxr_end_system_delay", Help: "RTCPXR end system delay"}, []string{"node_id"})

	p.GaugeVecMetrics["heplify_vqrtcpxr_nlr"] = prometheus.NewGaugeVec(prometheus.GaugeOpts{Name: "heplify_vqrtcpxr_nlr", Help: "VQ-RTCPXR network packet loss rate"}, []string{"node_id"})
	p.GaugeVecMetrics["heplify_vqrtcpxr_jdr"] = prometheus.NewGaugeVec(prometheus.GaugeOpts{Name: "heplify_vqrtcpxr_jdr", Help: "VQ-RTCPXR jitter buffer discard rate"}, []string{"node_id"})
	p.GaugeVecMetrics["heplify_vqrtcpxr_iaj"] = prometheus.NewGaugeVec(prometheus.GaugeOpts{Name: "heplify_vqrtcpxr_iaj", Help: "VQ-RTCPXR interarrival jitter"}, []string{"node_id"})
	p.GaugeVecMetrics["heplify_vqrtcpxr_moslq"] = prometheus.NewGaugeVec(prometheus.GaugeOpts{Name: "heplify_vqrtcpxr_moslq", Help: "VQ-RTCPXR MOS listening voice quality"}, []string{"node_id"})
	p.GaugeVecMetrics["heplify_vqrtcpxr_moscq"] = prometheus.NewGaugeVec(prometheus.GaugeOpts{Name: "heplify_vqrtcpxr_moscq", Help: "VQ-RTCPXR MOS conversation voice quality"}, []string{"node_id"})

	if config.Setting.RTPAgentStats {
		p.GaugeVecMetrics["heplify_rtpagent_delta"] = prometheus.NewGaugeVec(prometheus.GaugeOpts{Name: "heplify_rtpagent_delta", Help: "RTPAgent delta"}, []string{"node_id"})
		p.GaugeVecMetrics["heplify_rtpagent_jitter"] = prometheus.NewGaugeVec(prometheus.GaugeOpts{Name: "heplify_rtpagent_jitter", Help: "RTPAgent jitter"}, []string{"node_id"})
		p.GaugeVecMetrics["heplify_rtpagent_mos"] = prometheus.NewGaugeVec(prometheus.GaugeOpts{Name: "heplify_rtpagent_mos", Help: "RTPAgent mos"}, []string{"node_id"})
		p.GaugeVecMetrics["heplify_rtpagent_packets_lost"] = prometheus.NewGaugeVec(prometheus.GaugeOpts{Name: "heplify_rtpagent_packets_lost", Help: "RTPAgent packets lost"}, []string{"node_id"})
		p.GaugeVecMetrics["heplify_rtpagent_rfactor"] = prometheus.NewGaugeVec(prometheus.GaugeOpts{Name: "heplify_rtpagent_rfactor", Help: "RTPAgent rfactor"}, []string{"node_id"})
		p.GaugeVecMetrics["heplify_rtpagent_skew"] = prometheus.NewGaugeVec(prometheus.GaugeOpts{Name: "heplify_rtpagent_skew", Help: "RTPAgent skew"}, []string{"node_id"})

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

func (p *Prometheus) expose(hCh chan *decoder.HEP) {

	for {
		select {
		case pkt, ok := <-hCh:
			if !ok {
				break
			}

			nodeID := strconv.Itoa(int(pkt.NodeID))
			labelType := decoder.HEPTypeString(pkt.ProtoType)

			packetsByType.WithLabelValues(labelType).Inc()
			packetsBySize.WithLabelValues(labelType).Set(float64(len(pkt.Payload)))

			if pkt.SIP != nil && pkt.ProtoType == 1 {
				if !p.TargetEmpty {
					st, ok := p.TargetMap[pkt.SrcIP]
					if ok {
						methodResponses.WithLabelValues(st, "src", nodeID, pkt.SIP.StartLine.Method, pkt.SIP.CseqMethod).Inc()
						if pkt.SIP.RTPStatVal != "" {
							p.dissectXRTPStats(st, pkt.SIP.RTPStatVal)
						}
					}
					dt, ok := p.TargetMap[pkt.DstIP]
					if ok {
						methodResponses.WithLabelValues(dt, "dst", nodeID, pkt.SIP.StartLine.Method, pkt.SIP.CseqMethod).Inc()
						if pkt.SIP.RTPStatVal != "" {
							p.dissectXRTPStats(dt, pkt.SIP.RTPStatVal)
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

					methodResponses.WithLabelValues(
						"", "", nodeID, pkt.SIP.StartLine.Method, pkt.SIP.CseqMethod).Inc()

					if pkt.SIP.RTPStatVal != "" {
						p.dissectXRTPStats("", pkt.SIP.RTPStatVal)
					}
				}
				if pkt.SIP.ReasonVal != "" && strings.Contains(pkt.SIP.ReasonVal, "850") {
					reasonCause.WithLabelValues(extractXR("cause=", pkt.SIP.ReasonVal), nodeID).Inc()
				}
			} else if pkt.ProtoType == 5 {
				p.dissectRTCPStats(nodeID, []byte(pkt.Payload))
			} else if pkt.ProtoType == 34 && config.Setting.RTPAgentStats {
				p.dissectRTPStats(nodeID, []byte(pkt.Payload))
			} else if pkt.ProtoType == 35 {
				p.dissectRTCPXRStats(nodeID, pkt.Payload)
			} else if pkt.ProtoType == 38 && config.Setting.HoraclifixStats {
				p.dissectHoraclifixStats([]byte(pkt.Payload))
			}
		}
	}
}
