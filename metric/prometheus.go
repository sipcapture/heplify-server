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
	"time"

	"github.com/coocood/freecache"
	"github.com/hashicorp/golang-lru"
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
	cache           *freecache.Cache
	lruID           *lru.Cache
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
			p.cache = freecache.NewCache(60 * 1024 * 1024)
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

	p.lruID, err = lru.New(1e5)
	if err != nil {
		return err
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
				var st, dt string
				if !p.TargetEmpty {
					var ok bool
					st, ok = p.TargetMap[pkt.SrcIP]
					if ok {
						methodResponses.WithLabelValues(st, "src", nodeID, pkt.SIP.StartLine.Method, pkt.SIP.CseqMethod).Inc()
					}
					dt, ok = p.TargetMap[pkt.DstIP]
					if ok {
						methodResponses.WithLabelValues(dt, "dst", nodeID, pkt.SIP.StartLine.Method, pkt.SIP.CseqMethod).Inc()
					}
				} else {
					_, err := p.cache.Get([]byte(pkt.SIP.CallID + pkt.SIP.StartLine.Method + pkt.SIP.CseqMethod))
					if err == nil {
						continue
					}
					err = p.cache.Set([]byte(pkt.SIP.CallID+pkt.SIP.StartLine.Method+pkt.SIP.CseqMethod), nil, 600)
					if err != nil {
						logp.Warn("%v", err)
					}
					methodResponses.WithLabelValues("", "", nodeID, pkt.SIP.StartLine.Method, pkt.SIP.CseqMethod).Inc()
				}

				p.requestDelay(st, dt, pkt.SIP.CallID, pkt.SIP.StartLine.Method, pkt.SIP.CseqMethod, pkt.Timestamp)

				if pkt.SIP.RTPStatVal != "" {
					p.dissectXRTPStats(st, pkt.SIP.RTPStatVal)
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

func (p *Prometheus) requestDelay(st, dt, cid, sm, cm string, ts time.Time) {
	if !p.TargetEmpty && st == "" {
		return
	}
	for {
		if strings.HasSuffix(cid, "_b2b-1") {
			cid = cid[:len(cid)-6]
			continue
		}
		break
	}

	//TODO: tweak performance avoid double lru add
	if (sm == "INVITE" && cm == "INVITE") || (sm == "REGISTER" && cm == "REGISTER") {
		_, ok := p.lruID.Get(cid)
		if !ok {
			p.lruID.Add(cid, ts)
			p.lruID.Add(st+cid, ts)
		}
	}

	if (cm == "INVITE" || cm == "REGISTER") && (sm == "180" || sm == "183" || sm == "200") {
		did := dt + cid
		t, ok := p.lruID.Get(did)
		if ok {
			if cm == "INVITE" {
				srd.WithLabelValues(st, dt).Set(float64(ts.Sub(t.(time.Time))))
			} else {
				rrd.WithLabelValues(st, dt).Set(float64(ts.Sub(t.(time.Time))))
			}
			p.lruID.Remove(cid)
			p.lruID.Remove(did)
		}
	}
}
