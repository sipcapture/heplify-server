package metric

import (
	"encoding/binary"
	"fmt"
	"net"
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

type cidrTarget struct {
	net  *net.IPNet
	name string
}

type Prometheus struct {
	TargetEmpty bool
	TargetIP    []string
	TargetName  []string
	TargetMap   map[string]string
	TargetCIDRs []cidrTarget
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
			if err = p.setTargets(p.TargetIP, p.TargetName); err != nil {
				return err
			}
		}
	} else {
		logp.Info("please give every PromTargetIP a unique IP and PromTargetName a unique name")
		return fmt.Errorf("faulty PromTargetIP or PromTargetName")
	}

	return err
}

func (p *Prometheus) setTargets(ips, names []string) error {
	exact := make(map[string]string)
	var cidrs []cidrTarget
	for i := 0; i < len(names); i++ {
		ip := ips[i]
		name := names[i]
		if strings.Contains(ip, "/") {
			_, network, err := net.ParseCIDR(ip)
			if err != nil {
				return fmt.Errorf("invalid PromTargetIP CIDR %q: %w", ip, err)
			}
			cidrs = append(cidrs, cidrTarget{net: network, name: name})
			continue
		}
		exact[ip] = name
	}
	p.TargetMap = exact
	p.TargetCIDRs = cidrs
	return nil
}

func (p *Prometheus) matchTarget(ipStr string) (string, bool) {
	if name, ok := p.TargetMap[ipStr]; ok {
		return name, true
	}
	if len(p.TargetCIDRs) == 0 {
		return "", false
	}
	ip := net.ParseIP(ipStr)
	if ip == nil {
		return "", false
	}
	for _, c := range p.TargetCIDRs {
		if c.net.Contains(ip) {
			return c.name, true
		}
	}
	return "", false
}

func (p *Prometheus) expose(hCh chan *decoder.HEP) {
	for pkt := range hCh {
		packetsByType.WithLabelValues(pkt.NodeName, pkt.ProtoString).Inc()
		packetsBySize.WithLabelValues(pkt.NodeName, pkt.ProtoString).Set(float64(len(pkt.Payload)))

		var srcTarget, dstTarget string
		var srcHit, dstHit bool

		if !p.TargetEmpty {
			p.TargetConf.RLock()
			srcTarget, srcHit = p.matchTarget(pkt.SrcIP)
			dstTarget, dstHit = p.matchTarget(pkt.DstIP)
			p.TargetConf.RUnlock()
		}

		if pkt.SIP != nil && pkt.ProtoType == 1 {
			if !p.TargetEmpty {
				if srcHit {
					methodResponses.WithLabelValues(srcTarget, "src", pkt.NodeName, pkt.SIP.FirstMethod, pkt.SIP.CseqMethod).Inc()

					if pkt.SIP.ReasonVal != "" && strings.Contains(pkt.SIP.ReasonVal, "850") {
						reasonCause.WithLabelValues(srcTarget, extractXR("cause=", pkt.SIP.ReasonVal), pkt.SIP.FirstMethod).Inc()
					}
				}
				if dstHit {
					methodResponses.WithLabelValues(dstTarget, "dst", pkt.NodeName, pkt.SIP.FirstMethod, pkt.SIP.CseqMethod).Inc()
				}
				if !srcHit && !dstHit {
					methodResponses.WithLabelValues("unknown", "", pkt.NodeName, pkt.SIP.FirstMethod, pkt.SIP.CseqMethod).Inc()
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
			if srcHit {
				p.dissectRTCPStats(srcTarget, "src", pkt.NodeName, []byte(pkt.Payload))
			}
			if dstHit {
				p.dissectRTCPStats(dstTarget, "dst", pkt.NodeName, []byte(pkt.Payload))
			}
			if !srcHit && !dstHit {
				p.dissectRTCPStats("unknown", "", pkt.NodeName, []byte(pkt.Payload))
			}
		} else if pkt.ProtoType == 34 {
			p.dissectRTPStats(pkt.NodeName, []byte(pkt.Payload))
		} else if pkt.ProtoType == 35 {
			p.dissectRTCPXRStats(pkt.NodeName, pkt.Payload)
		} else if pkt.ProtoType == 38 {
			p.dissectHoraclifixStats([]byte(pkt.Payload))
		}
	}
}
