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
	cacheSize = 80 * 1024 * 1024
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
	var st, dt, cause string
	var withQ bool
	for pkt := range hCh {
		packetsByType.WithLabelValues(pkt.ProtoString).Inc()
		packetsBySize.WithLabelValues(pkt.ProtoString).Set(float64(len(pkt.Payload)))

		if pkt.SIP != nil && pkt.ProtoType == 1 {
			withQ = false
			if pkt.SIP.ReasonVal != "" && strings.Contains(pkt.SIP.ReasonVal, "850") {
				cause = extractXR("cause=", pkt.SIP.ReasonVal)
				withQ = true
			}

			if !p.TargetEmpty {
				var ok bool
				st, ok = p.TargetMap[pkt.SrcIP]
				if ok {
					methodResponses.WithLabelValues(st, "src", "", pkt.SIP.FirstMethod, pkt.SIP.CseqMethod).Inc()
					if withQ {
						reasonCause.WithLabelValues(st, "", cause).Inc()
					}
				}
				dt, ok = p.TargetMap[pkt.DstIP]
				if ok {
					methodResponses.WithLabelValues(dt, "dst", "", pkt.SIP.FirstMethod, pkt.SIP.CseqMethod).Inc()
					if withQ {
						reasonCause.WithLabelValues(dt, "", cause).Inc()
					}
				}
			}

			if (pkt.SIP.FirstMethod == invite && pkt.SIP.CseqMethod == invite) ||
				(pkt.SIP.FirstMethod == register && pkt.SIP.CseqMethod == register) {
				ik := []byte(pkt.CID)
				if !p.cache.Has(ik) {
					sk := []byte(pkt.SrcIP + pkt.CID)
					tu := uint64(pkt.Timestamp.UnixNano())
					tb := make([]byte, 8)

					binary.BigEndian.PutUint64(tb, tu)
					p.cache.Set(ik, tb)
					p.cache.Set(sk, tb)
				}
			}

			if (pkt.SIP.CseqMethod == invite && (pkt.SIP.FirstMethod == "180" || pkt.SIP.FirstMethod == "183")) ||
				(pkt.SIP.CseqMethod == register && pkt.SIP.FirstMethod == "200") {
				did := []byte(pkt.DstIP + pkt.CID)
				if buf := p.cache.Get(nil, did); buf != nil {
					i := binary.BigEndian.Uint64(buf)
					c := pkt.Timestamp.UnixNano()
					d := uint64(c) - i

					if dt == "" {
						dt = st
					}

					if pkt.SIP.CseqMethod == invite {
						srd.WithLabelValues(dt, pkt.NodeName).Set(float64(d))
					} else {
						rrd.WithLabelValues(dt, pkt.NodeName).Set(float64(d))
						p.cache.Del([]byte(pkt.CID))
					}
					p.cache.Del(did)
				}
			}

			if p.TargetEmpty {
				k := []byte(pkt.CID + pkt.SIP.FirstMethod + pkt.SIP.CseqMethod)
				if p.cache.Has(k) {
					continue
				}
				p.cache.Set(k, nil)
				methodResponses.WithLabelValues("", "", pkt.NodeName, pkt.SIP.FirstMethod, pkt.SIP.CseqMethod).Inc()
				if withQ {
					reasonCause.WithLabelValues("", pkt.NodeName, cause).Inc()
				}
			}

			if pkt.SIP.RTPStatVal != "" {
				p.dissectXRTPStats(st, pkt.SIP.RTPStatVal)
			}

		} else if pkt.ProtoType == 5 {
			p.dissectRTCPStats(pkt.NodeName, []byte(pkt.Payload))
		} else if pkt.ProtoType == 34 {
			p.dissectRTPStats(pkt.NodeName, []byte(pkt.Payload))
		} else if pkt.ProtoType == 35 {
			p.dissectRTCPXRStats(pkt.NodeName, pkt.Payload)
		} else if pkt.ProtoType == 38 {
			p.dissectHoraclifixStats([]byte(pkt.Payload))
		} else if pkt.ProtoType == 112 {
			logSeverity.WithLabelValues(pkt.NodeName, pkt.CID, pkt.HostTag).Inc()
		}
	}
}
