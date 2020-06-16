package metric

import (
	"encoding/binary"
	"fmt"
	"strings"
	"sync"

	"github.com/VictoriaMetrics/fastcache"
	"github.com/negbie/logp"
	"github.com/sipcapture/heplify-server/config"
	"github.com/sipcapture/heplify-server/decoder"
)

const (
	invite    = "INVITE"
	register  = "REGISTER"
	cacheSize = 60 * 1024 * 1024
)

type Prometheus struct {
	TargetEmpty bool
	TargetIP    []string
	TargetName  []string
	TargetMap   map[string]string
	TargetConf  *sync.RWMutex
	cache       *fastcache.Cache
}

func (p *Prometheus) setup() (err error) {
	p.TargetConf = new(sync.RWMutex)
	p.TargetIP = strings.Split(cutSpace(config.Setting.PromTargetIP), ",")
	p.TargetName = strings.Split(cutSpace(config.Setting.PromTargetName), ",")
	p.cache = fastcache.New(cacheSize)

	if len(p.TargetIP) == len(p.TargetName) && p.TargetIP != nil && p.TargetName != nil {
		if len(p.TargetIP[0]) == 0 || len(p.TargetName[0]) == 0 {
			logp.Info("expose metrics without or unbalanced targets")
			p.TargetIP[0] = ""
			p.TargetName[0] = ""
			p.TargetEmpty = true
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

	return err
}

func (p *Prometheus) expose(hCh chan *decoder.HEP) {
	for pkt := range hCh {
		packetsByType.WithLabelValues(pkt.NodeName, pkt.ProtoString).Inc()
		packetsBySize.WithLabelValues(pkt.NodeName, pkt.ProtoString).Set(float64(len(pkt.Payload)))

		var srcTarget, dstTarget string
		if pkt.SIP != nil && pkt.ProtoType == 1 {
			if !p.TargetEmpty {
				var srcHit, dstHit bool
				srcTarget, srcHit = p.TargetMap[pkt.SrcIP]
				if srcHit {
					methodResponses.WithLabelValues(srcTarget, "src", "", pkt.SIP.FirstMethod, pkt.SIP.CseqMethod).Inc()

					if pkt.SIP.ReasonVal != "" && strings.Contains(pkt.SIP.ReasonVal, "850") {
						reasonCause.WithLabelValues(srcTarget, extractXR("cause=", pkt.SIP.ReasonVal), pkt.SIP.FirstMethod).Inc()
					}
				}
				dstTarget, dstHit = p.TargetMap[pkt.DstIP]
				if dstHit {
					methodResponses.WithLabelValues(dstTarget, "dst", "", pkt.SIP.FirstMethod, pkt.SIP.CseqMethod).Inc()
				}
				if !srcHit && !dstHit {
					methodResponses.WithLabelValues("unknown", "", "", pkt.SIP.FirstMethod, pkt.SIP.CseqMethod).Inc()
				}
			}

			skip := false
			if dstTarget == "" && srcTarget == "" && !p.TargetEmpty {
				skip = true
			}

			callID := pkt.SID
			for {
				if strings.HasSuffix(callID, "_b2b-1") {
					callID = callID[:len(callID)-6]
					continue
				}
				break
			}

			if !skip && ((pkt.SIP.FirstMethod == invite && pkt.SIP.CseqMethod == invite) ||
				(pkt.SIP.FirstMethod == register && pkt.SIP.CseqMethod == register)) {
				ptn := pkt.Timestamp.UnixNano()
				sid := []byte(callID)
				buf := p.cache.Get(nil, sid)
				if buf == nil || buf != nil && (uint64(ptn) < binary.BigEndian.Uint64(buf)) {
					sk := []byte(pkt.SrcIP + callID)
					tb := make([]byte, 8)

					binary.BigEndian.PutUint64(tb, uint64(ptn))
					p.cache.Set(sid, tb)
					p.cache.Set(sk, tb)
				}
			}

			if !skip && ((pkt.SIP.CseqMethod == invite || pkt.SIP.CseqMethod == register) &&
				(pkt.SIP.FirstMethod == "180" ||
					pkt.SIP.FirstMethod == "181" ||
					pkt.SIP.FirstMethod == "182" ||
					pkt.SIP.FirstMethod == "183" ||
					pkt.SIP.FirstMethod == "200")) {
				ptn := pkt.Timestamp.UnixNano()
				did := []byte(pkt.DstIP + callID)
				if buf := p.cache.Get(nil, did); buf != nil {
					d := uint64(ptn) - binary.BigEndian.Uint64(buf)

					if dstTarget == "" {
						dstTarget = srcTarget
					}

					if pkt.SIP.CseqMethod == invite {
						srd.WithLabelValues(dstTarget, pkt.NodeName).Set(float64(d))
					} else {
						rrd.WithLabelValues(dstTarget, pkt.NodeName).Set(float64(d))
						p.cache.Del([]byte(callID))
					}
					p.cache.Del(did)
				}
			}

			if p.TargetEmpty {
				k := []byte(callID + pkt.SIP.FirstMethod + pkt.SIP.CseqMethod)
				if p.cache.Has(k) {
					continue
				}
				p.cache.Set(k, nil)
				methodResponses.WithLabelValues(pkt.TargetName, "", pkt.NodeName, pkt.SIP.FirstMethod, pkt.SIP.CseqMethod).Inc()

				if pkt.SIP.ReasonVal != "" && strings.Contains(pkt.SIP.ReasonVal, "850") {
					reasonCause.WithLabelValues(srcTarget, extractXR("cause=", pkt.SIP.ReasonVal), pkt.SIP.FirstMethod).Inc()
				}
			}

			if pkt.SIP.RTPStatVal != "" {
				p.dissectXRTPStats(srcTarget, pkt.SIP.RTPStatVal)
			}

		} else if pkt.ProtoType == 5 {
			p.dissectRTCPStats(pkt.NodeName, []byte(pkt.Payload))
		} else if pkt.ProtoType == 34 {
			p.dissectRTPStats(pkt.NodeName, []byte(pkt.Payload))
		} else if pkt.ProtoType == 35 {
			p.dissectRTCPXRStats(pkt.NodeName, pkt.Payload)
		} else if pkt.ProtoType == 38 {
			p.dissectHoraclifixStats([]byte(pkt.Payload))
		}
	}
}
