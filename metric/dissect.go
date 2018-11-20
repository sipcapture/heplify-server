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
		p.GaugeVecMetrics["heplify_vqrtcpxr_nlr"].WithLabelValues(nodeID).Set(nlr)
	}
	if jdr, err := strconv.ParseFloat(extractXR("JDR=", stats), 64); err == nil {
		p.GaugeVecMetrics["heplify_vqrtcpxr_jdr"].WithLabelValues(nodeID).Set(jdr)
	}
	if iaj, err := strconv.ParseFloat(extractXR("IAJ=", stats), 64); err == nil {
		p.GaugeVecMetrics["heplify_vqrtcpxr_iaj"].WithLabelValues(nodeID).Set(iaj)
	}
	if moslq, err := strconv.ParseFloat(extractXR("MOSLQ=", stats), 64); err == nil {
		p.GaugeVecMetrics["heplify_vqrtcpxr_moslq"].WithLabelValues(nodeID).Set(moslq)
	}
	if moscq, err := strconv.ParseFloat(extractXR("MOSCQ=", stats), 64); err == nil {
		p.GaugeVecMetrics["heplify_vqrtcpxr_moscq"].WithLabelValues(nodeID).Set(moscq)
	}
}

func (p *Prometheus) dissectXRTPStats(tn, stats string) {
	var err error
	plr, pls, jir, jis, dle, r, mos := 0, 0, 0, 0, 0, 0.0, 0.0

	p.CvPacketsTotal.WithLabelValues("xrtp").Inc()
	p.GvPacketsSize.WithLabelValues("xrtp").Set(float64(len(stats)))

	cs, err := strconv.ParseFloat(extractXR("CS=", stats), 64)
	if err == nil {
		p.GaugeVecMetrics["heplify_xrtp_cs"].WithLabelValues(tn).Set(cs / 1000)
	} else {
		logp.Err("%v", err)
	}

	if plt := extractXR("PL=", stats); len(plt) >= 1 {
		if pl := strings.Split(plt, ","); len(pl) == 2 {
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

	if jit := extractXR("JI=", stats); len(jit) >= 1 {
		if ji := strings.Split(jit, ","); len(ji) == 2 {
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

	if dlt := extractXR("DL=", stats); len(dlt) >= 1 {
		if dl := strings.Split(dlt, ","); len(dl) == 3 {
			dle, err = strconv.Atoi(dl[0])
			if err == nil {
				p.GaugeVecMetrics["heplify_xrtp_dle"].WithLabelValues(tn).Set(float64(dle))
			} else {
				logp.Err("%v", err)
			}
		}
	}

	pr, err := strconv.Atoi(extractXR("PR=", stats))
	if err != nil {
		logp.Err("%v", err)
	}
	ps, err := strconv.Atoi(extractXR("PS=", stats))
	if err != nil {
		logp.Err("%v", err)
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

func (p *Prometheus) dissectRTCPStats(nodeID string, data []byte) {
	jsonparser.EachKey(data, func(idx int, value []byte, vt jsonparser.ValueType, err error) {
		switch idx {
		case 0:
			if fractionLost, err := jsonparser.ParseFloat(value); err == nil {
				p.GaugeVecMetrics["heplify_rtcp_fraction_lost"].WithLabelValues(nodeID).Set(normMax(fractionLost))
			}
		case 1:
			if packetsLost, err := jsonparser.ParseFloat(value); err == nil {
				p.GaugeVecMetrics["heplify_rtcp_packets_lost"].WithLabelValues(nodeID).Set(normMax(packetsLost))
			}
		case 2:
			if iaJitter, err := jsonparser.ParseFloat(value); err == nil {
				p.GaugeVecMetrics["heplify_rtcp_jitter"].WithLabelValues(nodeID).Set(normMax(iaJitter))
			}
		case 3:
			if dlsr, err := jsonparser.ParseFloat(value); err == nil {
				p.GaugeVecMetrics["heplify_rtcp_dlsr"].WithLabelValues(nodeID).Set(normMax(dlsr))
			}
		case 4:
			if fractionLost, err := jsonparser.ParseFloat(value); err == nil {
				p.GaugeVecMetrics["heplify_rtcpxr_fraction_lost"].WithLabelValues(nodeID).Set(fractionLost)
			}
		case 5:
			if fractionDiscard, err := jsonparser.ParseFloat(value); err == nil {
				p.GaugeVecMetrics["heplify_rtcpxr_fraction_discard"].WithLabelValues(nodeID).Set(fractionDiscard)
			}
		case 6:
			if burstDensity, err := jsonparser.ParseFloat(value); err == nil {
				p.GaugeVecMetrics["heplify_rtcpxr_burst_density"].WithLabelValues(nodeID).Set(burstDensity)
			}
		case 7:
			if gapDensity, err := jsonparser.ParseFloat(value); err == nil {
				p.GaugeVecMetrics["heplify_rtcpxr_gap_density"].WithLabelValues(nodeID).Set(gapDensity)
			}
		case 8:
			if burstDuration, err := jsonparser.ParseFloat(value); err == nil {
				p.GaugeVecMetrics["heplify_rtcpxr_burst_duration"].WithLabelValues(nodeID).Set(burstDuration)
			}
		case 9:
			if gapDuration, err := jsonparser.ParseFloat(value); err == nil {
				p.GaugeVecMetrics["heplify_rtcpxr_gap_duration"].WithLabelValues(nodeID).Set(gapDuration)
			}
		case 10:
			if roundTripDelay, err := jsonparser.ParseFloat(value); err == nil {
				p.GaugeVecMetrics["heplify_rtcpxr_round_trip_delay"].WithLabelValues(nodeID).Set(roundTripDelay)
			}
		case 11:
			if endSystemDelay, err := jsonparser.ParseFloat(value); err == nil {
				p.GaugeVecMetrics["heplify_rtcpxr_end_system_delay"].WithLabelValues(nodeID).Set(endSystemDelay)
			}
		}
	}, p.rtcpPaths...)
}

func (p *Prometheus) dissectRTPStats(nodeID string, data []byte) {
	jsonparser.EachKey(data, func(idx int, value []byte, vt jsonparser.ValueType, err error) {
		switch idx {
		case 0:
			if delta, err := jsonparser.ParseFloat(value); err == nil {
				p.GaugeVecMetrics["heplify_rtpagent_delta"].WithLabelValues(nodeID).Set(delta)
			}
		case 1:
			if iaJitter, err := jsonparser.ParseFloat(value); err == nil {
				p.GaugeVecMetrics["heplify_rtpagent_jitter"].WithLabelValues(nodeID).Set(iaJitter)
			}
		case 2:
			if mos, err := jsonparser.ParseFloat(value); err == nil {
				p.GaugeVecMetrics["heplify_rtpagent_mos"].WithLabelValues(nodeID).Set(mos)
			}
		case 3:
			if packetsLost, err := jsonparser.ParseFloat(value); err == nil {
				p.GaugeVecMetrics["heplify_rtpagent_packets_lost"].WithLabelValues(nodeID).Set(packetsLost)
			}
		case 4:
			if rfactor, err := jsonparser.ParseFloat(value); err == nil {
				p.GaugeVecMetrics["heplify_rtpagent_rfactor"].WithLabelValues(nodeID).Set(rfactor)
			}
		case 5:
			if skew, err := jsonparser.ParseFloat(value); err == nil {
				p.GaugeVecMetrics["heplify_rtpagent_skew"].WithLabelValues(nodeID).Set(skew)
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
