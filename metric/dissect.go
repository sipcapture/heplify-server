package metric

import (
	"strconv"
	"strings"

	"github.com/buger/jsonparser"
	"github.com/negbie/logp"
)

func normMax(val float64) float64 {
	if val > 10000000 {
		return 0
	}
	return val
}

func (p *Prometheus) dissectRTCPXRStats(nodeID string, stats string) {
	if nlr, err := strconv.ParseFloat(extractXR("NLR=", stats), 64); err == nil {
		vqrtcpxrNLR.WithLabelValues(nodeID).Set(nlr)
	}
	if jdr, err := strconv.ParseFloat(extractXR("JDR=", stats), 64); err == nil {
		vqrtcpxrJDR.WithLabelValues(nodeID).Set(jdr)
	}
	if iaj, err := strconv.ParseFloat(extractXR("IAJ=", stats), 64); err == nil {
		vqrtcpxrIAJ.WithLabelValues(nodeID).Set(iaj)
	}
	if moslq, err := strconv.ParseFloat(extractXR("MOSLQ=", stats), 64); err == nil {
		vqrtcpxrMOSLQ.WithLabelValues(nodeID).Set(moslq)
	}
	if moscq, err := strconv.ParseFloat(extractXR("MOSCQ=", stats), 64); err == nil {
		vqrtcpxrMOSCQ.WithLabelValues(nodeID).Set(moscq)
	}
}

func (p *Prometheus) dissectXRTPStats(tn, stats string) {
	var err error
	plr, pls, jir, jis, dle, r, mos := 0, 0, 0, 0, 0, 0.0, 0.0

	packetsByType.WithLabelValues("xrtp").Inc()
	packetsBySize.WithLabelValues("xrtp").Set(float64(len(stats)))

	if cs, err := strconv.ParseFloat(extractXR("CS=", stats), 64); err == nil {
		xrtpCS.WithLabelValues(tn).Set(cs / 1000)
	}

	if plt := extractXR("PL=", stats); len(plt) >= 1 {
		if pl := strings.Split(plt, ","); len(pl) == 2 {
			if plr, err = strconv.Atoi(pl[0]); err == nil {
				xrtpPLR.WithLabelValues(tn).Set(float64(plr))
			}
			if pls, err = strconv.Atoi(pl[1]); err == nil {
				xrtpPLS.WithLabelValues(tn).Set(float64(pls))
			}
		}
	}

	if jit := extractXR("JI=", stats); len(jit) >= 1 {
		if ji := strings.Split(jit, ","); len(ji) == 2 {
			if jir, err = strconv.Atoi(ji[0]); err == nil {
				xrtpJIR.WithLabelValues(tn).Set(float64(jir))
			}
			if jis, err = strconv.Atoi(ji[1]); err == nil {
				xrtpJIS.WithLabelValues(tn).Set(float64(jis))
			}
		}
	}

	if dlt := extractXR("DL=", stats); len(dlt) >= 1 {
		if dl := strings.Split(dlt, ","); len(dl) == 3 {
			if dle, err = strconv.Atoi(dl[0]); err == nil {
				xrtpDLE.WithLabelValues(tn).Set(float64(dle))
			}
		}
	}

	pr, _ := strconv.Atoi(extractXR("PR=", stats))
	ps, _ := strconv.Atoi(extractXR("PS=", stats))
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
	xrtpMOS.WithLabelValues(tn).Set(mos)
}

func (p *Prometheus) dissectRTCPStats(nodeID string, data []byte) {
	jsonparser.EachKey(data, func(idx int, value []byte, vt jsonparser.ValueType, err error) {
		switch idx {
		case 0:
			if fractionLost, err := jsonparser.ParseFloat(value); err == nil {
				rtcpFractionLost.WithLabelValues(nodeID).Set(normMax(fractionLost))
			}
		case 1:
			if packetsLost, err := jsonparser.ParseFloat(value); err == nil {
				rtcpPacketsLost.WithLabelValues(nodeID).Set(normMax(packetsLost))
			}
		case 2:
			if iaJitter, err := jsonparser.ParseFloat(value); err == nil {
				rtcpJitter.WithLabelValues(nodeID).Set(normMax(iaJitter))
			}
		case 3:
			if dlsr, err := jsonparser.ParseFloat(value); err == nil {
				rtcpDLSR.WithLabelValues(nodeID).Set(normMax(dlsr))
			}
		case 4:
			if fractionLost, err := jsonparser.ParseFloat(value); err == nil {
				rtcpxrFractionLost.WithLabelValues(nodeID).Set(fractionLost)
			}
		case 5:
			if fractionDiscard, err := jsonparser.ParseFloat(value); err == nil {
				rtcpxrFractionDiscard.WithLabelValues(nodeID).Set(fractionDiscard)
			}
		case 6:
			if burstDensity, err := jsonparser.ParseFloat(value); err == nil {
				rtcpxrBurstDensity.WithLabelValues(nodeID).Set(burstDensity)
			}
		case 7:
			if gapDensity, err := jsonparser.ParseFloat(value); err == nil {
				rtcpxrGapDensity.WithLabelValues(nodeID).Set(gapDensity)
			}
		case 8:
			if burstDuration, err := jsonparser.ParseFloat(value); err == nil {
				rtcpxrBurstDuration.WithLabelValues(nodeID).Set(burstDuration)
			}
		case 9:
			if gapDuration, err := jsonparser.ParseFloat(value); err == nil {
				rtcpxrGapDuration.WithLabelValues(nodeID).Set(gapDuration)
			}
		case 10:
			if roundTripDelay, err := jsonparser.ParseFloat(value); err == nil {
				rtcpxrRoundTripDelay.WithLabelValues(nodeID).Set(roundTripDelay)
			}
		case 11:
			if endSystemDelay, err := jsonparser.ParseFloat(value); err == nil {
				rtcpxrEndSystemDelay.WithLabelValues(nodeID).Set(endSystemDelay)
			}
		}
	}, p.rtcpPaths...)
}

func (p *Prometheus) dissectRTPStats(nodeID string, data []byte) {
	jsonparser.EachKey(data, func(idx int, value []byte, vt jsonparser.ValueType, err error) {
		switch idx {
		case 0:
			if delta, err := jsonparser.ParseFloat(value); err == nil {
				rtpagentDelta.WithLabelValues(nodeID).Set(delta)
			}
		case 1:
			if iaJitter, err := jsonparser.ParseFloat(value); err == nil {
				rtpagentJitter.WithLabelValues(nodeID).Set(iaJitter)
			}
		case 2:
			if mos, err := jsonparser.ParseFloat(value); err == nil {
				rtpagentMOS.WithLabelValues(nodeID).Set(mos)
			}
		case 3:
			if packetsLost, err := jsonparser.ParseFloat(value); err == nil {
				rtpagentPacketsLost.WithLabelValues(nodeID).Set(packetsLost)
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
				horaclifixRtpMOS.WithLabelValues(sbcName, "inc", incRealm, outRealm).Set(incMos / 100)
			}
		case 4:
			if incRval, err := jsonparser.ParseFloat(value); err == nil {
				horaclifixRtpRVAL.WithLabelValues(sbcName, "inc", incRealm, outRealm).Set(incRval / 100)
			}
		case 5:
			if incRtpPackets, err := jsonparser.ParseFloat(value); err == nil {
				horaclifixRtpPackets.WithLabelValues(sbcName, "inc", incRealm, outRealm).Set(incRtpPackets)
			}
		case 6:
			if incRtpLostPackets, err := jsonparser.ParseFloat(value); err == nil {
				horaclifixRtpLostPackets.WithLabelValues(sbcName, "inc", incRealm, outRealm).Set(incRtpLostPackets)
			}
		case 7:
			if incRtpAvgJitter, err := jsonparser.ParseFloat(value); err == nil {
				horaclifixRtpAvgJitter.WithLabelValues(sbcName, "inc", incRealm, outRealm).Set(incRtpAvgJitter)
			}
		case 8:
			if incRtpMaxJitter, err := jsonparser.ParseFloat(value); err == nil {
				horaclifixRtpMaxJitter.WithLabelValues(sbcName, "inc", incRealm, outRealm).Set(incRtpMaxJitter)
			}
		case 9:
			if incRtcpPackets, err := jsonparser.ParseFloat(value); err == nil {
				horaclifixRtcpPackets.WithLabelValues(sbcName, "inc", incRealm, outRealm).Set(incRtcpPackets)
			}
		case 10:
			if incRtcpLostPackets, err := jsonparser.ParseFloat(value); err == nil {
				horaclifixRtcpLostPackets.WithLabelValues(sbcName, "inc", incRealm, outRealm).Set(incRtcpLostPackets)
			}
		case 11:
			if incRtcpAvgJitter, err := jsonparser.ParseFloat(value); err == nil {
				horaclifixRtcpAvgJitter.WithLabelValues(sbcName, "inc", incRealm, outRealm).Set(incRtcpAvgJitter)
			}
		case 12:
			if incRtcpMaxJitter, err := jsonparser.ParseFloat(value); err == nil {
				horaclifixRtcpMaxJitter.WithLabelValues(sbcName, "inc", incRealm, outRealm).Set(incRtcpMaxJitter)
			}
		case 13:
			if incRtcpAvgLat, err := jsonparser.ParseFloat(value); err == nil {
				horaclifixRtcpAvgLAT.WithLabelValues(sbcName, "inc", incRealm, outRealm).Set(incRtcpAvgLat)
			}
		case 14:
			if incRtcpMaxLat, err := jsonparser.ParseFloat(value); err == nil {
				horaclifixRtcpMaxLAT.WithLabelValues(sbcName, "inc", incRealm, outRealm).Set(incRtcpMaxLat)
			}
		case 15:
			if outMos, err := jsonparser.ParseFloat(value); err == nil {
				horaclifixRtpMOS.WithLabelValues(sbcName, "out", incRealm, outRealm).Set(outMos / 100)
			}
		case 16:
			if outRval, err := jsonparser.ParseFloat(value); err == nil {
				horaclifixRtpRVAL.WithLabelValues(sbcName, "out", incRealm, outRealm).Set(outRval / 100)
			}
		case 17:
			if outRtpPackets, err := jsonparser.ParseFloat(value); err == nil {
				horaclifixRtpPackets.WithLabelValues(sbcName, "out", incRealm, outRealm).Set(outRtpPackets)
			}
		case 18:
			if outRtpLostPackets, err := jsonparser.ParseFloat(value); err == nil {
				horaclifixRtpLostPackets.WithLabelValues(sbcName, "out", incRealm, outRealm).Set(outRtpLostPackets)
			}
		case 19:
			if outRtpAvgJitter, err := jsonparser.ParseFloat(value); err == nil {
				horaclifixRtpAvgJitter.WithLabelValues(sbcName, "out", incRealm, outRealm).Set(outRtpAvgJitter)
			}
		case 20:
			if outRtpMaxJitter, err := jsonparser.ParseFloat(value); err == nil {
				horaclifixRtpMaxJitter.WithLabelValues(sbcName, "out", incRealm, outRealm).Set(outRtpMaxJitter)
			}
		case 21:
			if outRtcpPackets, err := jsonparser.ParseFloat(value); err == nil {
				horaclifixRtcpPackets.WithLabelValues(sbcName, "out", incRealm, outRealm).Set(outRtcpPackets)
			}
		case 22:
			if outRtcpLostPackets, err := jsonparser.ParseFloat(value); err == nil {
				horaclifixRtcpLostPackets.WithLabelValues(sbcName, "out", incRealm, outRealm).Set(outRtcpLostPackets)
			}
		case 23:
			if outRtcpAvgJitter, err := jsonparser.ParseFloat(value); err == nil {
				horaclifixRtcpAvgJitter.WithLabelValues(sbcName, "out", incRealm, outRealm).Set(outRtcpAvgJitter)
			}
		case 24:
			if outRtcpMaxJitter, err := jsonparser.ParseFloat(value); err == nil {
				horaclifixRtcpMaxJitter.WithLabelValues(sbcName, "out", incRealm, outRealm).Set(outRtcpMaxJitter)
			}
		case 25:
			if outRtcpAvgLat, err := jsonparser.ParseFloat(value); err == nil {
				horaclifixRtcpAvgLAT.WithLabelValues(sbcName, "out", incRealm, outRealm).Set(outRtcpAvgLat)
			}
		case 26:
			if outRtcpMaxLat, err := jsonparser.ParseFloat(value); err == nil {
				horaclifixRtcpMaxLAT.WithLabelValues(sbcName, "out", incRealm, outRealm).Set(outRtcpMaxLat)
			}
		}
	}, p.horaclifixPaths...)
}
