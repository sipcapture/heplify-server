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
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

type Prometheus struct {
	TargetEmpty     bool
	TargetIP        []string
	TargetName      []string
	TargetMap       map[string]string
	TargetConf      *sync.RWMutex
	Cache           *freecache.Cache
	horaclifixPaths [][]string
	rtpPaths        [][]string
	rtcpPaths       [][]string
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

	p.horaclifixPaths = [][]string{
		[]string{"NAME"},
		[]string{"INC_REALM"},
		[]string{"OUT_REALM"},
		[]string{"INC_MOS"},
		[]string{"INC_RVAL"},
		[]string{"INC_RTP_PK"},
		[]string{"INC_RTP_PK_LOSS"},
		[]string{"INC_RTP_AVG_JITTER"},
		[]string{"INC_RTP_MAX_JITTER"},
		[]string{"INC_RTCP_PK"},
		[]string{"INC_RTCP_PK_LOSS"},
		[]string{"INC_RTCP_AVG_JITTER"},
		[]string{"INC_RTCP_MAX_JITTER"},
		[]string{"INC_RTCP_AVG_LAT"},
		[]string{"INC_RTCP_MAX_LAT"},
		[]string{"OUT_MOS"},
		[]string{"OUT_RVAL"},
		[]string{"OUT_RTP_PK"},
		[]string{"OUT_RTP_PK_LOSS"},
		[]string{"OUT_RTP_AVG_JITTER"},
		[]string{"OUT_RTP_MAX_JITTER"},
		[]string{"OUT_RTCP_PK"},
		[]string{"OUT_RTCP_PK_LOSS"},
		[]string{"OUT_RTCP_AVG_JITTER"},
		[]string{"OUT_RTCP_MAX_JITTER"},
		[]string{"OUT_RTCP_AVG_LAT"},
		[]string{"OUT_RTCP_MAX_LAT"},
	}
	p.rtpPaths = [][]string{
		[]string{"DELTA"},
		[]string{"JITTER"},
		[]string{"MOS"},
		[]string{"PACKET_LOSS"},
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
			} else if pkt.ProtoType == 34 {
				p.dissectRTPStats(nodeID, []byte(pkt.Payload))
			} else if pkt.ProtoType == 35 {
				p.dissectRTCPXRStats(nodeID, pkt.Payload)
			} else if pkt.ProtoType == 38 {
				p.dissectHoraclifixStats([]byte(pkt.Payload))
			}
		}
	}
}
