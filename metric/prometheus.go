package metric

import (
	"encoding/binary"
	"fmt"
	"strings"
	"time"

	//!!!!!
	"strconv"
	//!!!!!
	"sync"

	"github.com/VictoriaMetrics/fastcache"
	"github.com/negbie/logp"
	"github.com/sipcapture/heplify-server/config"
	"github.com/sipcapture/heplify-server/decoder"
)

const (
	invite   = "INVITE"
	register = "REGISTER"
	//!!cacheSize = 60 * 1024 * 1024
	cacheSize = 120 * 1024 * 1024
	bye       = "BYE"
)

type Prometheus struct {
	TargetEmpty       bool
	TargetIP          []string
	TargetName        []string
	TargetMap         map[string]string
	TargetConf        *sync.RWMutex
	cache             *fastcache.Cache
	muLastSync        *sync.RWMutex
	lastSyncDate      int64
	expirationTimeout int64
}

func (p *Prometheus) cleaner() {
	ticker := time.Tick(15 * time.Second)
	for range ticker {
		p.reset()
	}
}

func (p *Prometheus) reset() {
	p.muLastSync.RLock()
	defer p.muLastSync.RUnlock()
	if time.Now().Unix()-p.lastSyncDate < p.expirationTimeout {
		return
	}
	xrtpeMOS.Reset()
}

func (p *Prometheus) setup() (err error) {
	p.TargetConf = new(sync.RWMutex)
	p.muLastSync = new(sync.RWMutex)
	p.expirationTimeout = 60 // sec
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
	go p.cleaner()
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
			//!!!!!!!!!!!!!!!!!!!!!!
			// xrtpeMOS stats
			if !skip && (pkt.SIP.FirstMethod == bye && pkt.SIP.CseqMethod == bye) {
				if val, ok := pkt.SIP.CustomHeader["X-RTPE-MOS"]; ok && len(pkt.SIP.CustomHeader["X-RTPE-MOS"]) > 0 {
					if mos, err := strconv.ParseFloat(val, 64); err == nil {
						q_cc_name := ""
						if val, ok := pkt.SIP.CustomHeader["X-CC-Name"]; ok && len(pkt.SIP.CustomHeader["X-CC-Name"]) > 0 {
							q_cc_name = val
						}
						q_id_user := ""
						if val, ok := pkt.SIP.CustomHeader["X-ID-User"]; ok && len(pkt.SIP.CustomHeader["X-ID-User"]) > 0 {
							q_id_user = val
						}
						q_queue_name := ""
						if val, ok := pkt.SIP.CustomHeader["X-Queue-Name"]; ok && len(pkt.SIP.CustomHeader["X-Queue-Name"]) > 0 {
							q_queue_name = val
						}
						xrtpeMOS.WithLabelValues(q_queue_name, q_cc_name, q_id_user).Set(mos)
						p.muLastSync.Lock()
						p.lastSyncDate = time.Now().Unix()
						p.muLastSync.Unlock()
					}
				}

			}
			//!!!!!!!!!!!!!!!!!!!!!!

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
