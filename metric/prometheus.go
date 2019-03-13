package metric

import (
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/coocood/freecache"
	lru "github.com/hashicorp/golang-lru"
	"github.com/negbie/logp"
	"github.com/sipcapture/heplify-server/config"
	"github.com/sipcapture/heplify-server/decoder"
)

type Prometheus struct {
	TargetEmpty bool
	TargetIP    []string
	TargetName  []string
	TargetMap   map[string]string
	TargetConf  *sync.RWMutex
	cache       *freecache.Cache
	lruID       *lru.Cache
}

func (p *Prometheus) setup() (err error) {
	p.TargetConf = new(sync.RWMutex)
	p.TargetIP = strings.Split(cutSpace(config.Setting.PromTargetIP), ",")
	p.TargetName = strings.Split(cutSpace(config.Setting.PromTargetName), ",")

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

	p.lruID, err = lru.New(1e5)
	if err != nil {
		return err
	}

	return err
}

func (p *Prometheus) expose(hCh chan *decoder.HEP) {
	for pkt := range hCh {
		labelType := decoder.HEPTypeString(pkt.ProtoType)

		packetsByType.WithLabelValues(labelType).Inc()
		packetsBySize.WithLabelValues(labelType).Set(float64(len(pkt.Payload)))

		if pkt.SIP != nil && pkt.ProtoType == 1 {
			var st, dt string
			if !p.TargetEmpty {
				var ok bool
				st, ok = p.TargetMap[pkt.SrcIP]
				if ok {
					methodResponses.WithLabelValues(st, "src", "", pkt.SIP.FirstMethod, pkt.SIP.CseqMethod).Inc()
				}
				dt, ok = p.TargetMap[pkt.DstIP]
				if ok {
					methodResponses.WithLabelValues(dt, "dst", "", pkt.SIP.FirstMethod, pkt.SIP.CseqMethod).Inc()
				}
			} else {
				_, err := p.cache.Get([]byte(pkt.CID + pkt.SIP.FirstMethod + pkt.SIP.CseqMethod))
				if err == nil {
					continue
				}
				err = p.cache.Set([]byte(pkt.CID+pkt.SIP.FirstMethod+pkt.SIP.CseqMethod), nil, 600)
				if err != nil {
					logp.Warn("%v", err)
				}
				methodResponses.WithLabelValues("", "", pkt.Node, pkt.SIP.FirstMethod, pkt.SIP.CseqMethod).Inc()
			}

			p.requestDelay(st, dt, pkt.CID, pkt.SIP.FirstMethod, pkt.SIP.CseqMethod, pkt.Timestamp)

			if pkt.SIP.RTPStatVal != "" {
				p.dissectXRTPStats(st, pkt.SIP.RTPStatVal)
			}
			if pkt.SIP.ReasonVal != "" && strings.Contains(pkt.SIP.ReasonVal, "850") {
				reasonCause.WithLabelValues(extractXR("cause=", pkt.SIP.ReasonVal), pkt.Node).Inc()
			}
		} else if pkt.ProtoType == 5 {
			p.dissectRTCPStats(pkt.Node, []byte(pkt.Payload))
		} else if pkt.ProtoType == 34 {
			p.dissectRTPStats(pkt.Node, []byte(pkt.Payload))
		} else if pkt.ProtoType == 35 {
			p.dissectRTCPXRStats(pkt.Node, pkt.Payload)
		} else if pkt.ProtoType == 38 {
			p.dissectHoraclifixStats([]byte(pkt.Payload))
		} else if pkt.ProtoType == 112 {
			logSeverity.WithLabelValues(pkt.Node, pkt.CID, pkt.Host).Inc()
		} else if pkt.ProtoType == 1032 {
			p.dissectJanusStats([]byte(pkt.Payload))
		}
	}
}

func (p *Prometheus) requestDelay(st, dt, cid, sm, cm string, ts time.Time) {
	if !p.TargetEmpty && st == "" {
		return
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
