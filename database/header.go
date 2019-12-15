package database

import (
	"strconv"

	"github.com/buger/jsonparser"
	"github.com/negbie/logp"
	"github.com/sipcapture/heplify-server/decoder"
	"github.com/valyala/bytebufferpool"
	"github.com/valyala/fasttemplate"
)

func makeProtoHeader(h *decoder.HEP, sb *bytebufferpool.ByteBuffer) string {
	sb.WriteString(`{`)
	sb.WriteString(`"protocolFamily":`)
	sb.WriteString(strconv.FormatUint(uint64(h.Version), 10))
	sb.WriteString(`,"protocol":`)
	sb.WriteString(strconv.FormatUint(uint64(h.Protocol), 10))
	sb.WriteString(`,"srcIp":"`)
	sb.WriteString(h.SrcIP)
	sb.WriteString(`","dstIp":"`)
	sb.WriteString(h.DstIP)
	sb.WriteString(`","srcPort":`)
	sb.WriteString(strconv.FormatUint(uint64(h.SrcPort), 10))
	sb.WriteString(`,"dstPort":`)
	sb.WriteString(strconv.FormatUint(uint64(h.DstPort), 10))
	sb.WriteString(`,"timeSeconds":`)
	sb.WriteString(strconv.FormatUint(uint64(h.Tsec), 10))
	sb.WriteString(`,"timeUseconds":`)
	sb.WriteString(strconv.FormatUint(uint64(h.Tmsec), 10))
	sb.WriteString(`,"payloadType":`)
	sb.WriteString(strconv.FormatUint(uint64(h.ProtoType), 10))
	sb.WriteString(`,"captureId":"`)
	sb.WriteString(h.NodeName)
	if h.NodePW != "" {
		sb.WriteString(`","capturePass":"`)
		sb.WriteString(h.NodePW)
	}
	sb.WriteString(`","correlation_id":"`)
	sb.WriteString(h.CID)
	sb.WriteString(`"}`)
	return sb.String()
}

func makeSIPDataHeader(h *decoder.HEP, sb *bytebufferpool.ByteBuffer, t *fasttemplate.Template) string {
	sb.WriteString(`{`)

	t.ExecuteFunc(sb, h.EscapeFields)

	if len(h.SIP.CHeader) > 0 {
		for k, v := range h.SIP.CustomHeader {
			sb.WriteString(`,"` + k + `":"`)
			sb.WriteString(v + `"`)
		}
	}

	sb.WriteString(`}`)
	return sb.String()
}

func makeRTCDataHeader(h *decoder.HEP, sb *bytebufferpool.ByteBuffer) string {
	sb.WriteString(`{`)
	sb.WriteString(`"node":"`)
	sb.WriteString(h.NodeName)
	sb.WriteString(`","proto":"`)
	sb.WriteString(h.ProtoString)
	sb.WriteString(`"}`)
	return sb.String()
}

var IsupPaths = [][]string{
	[]string{"cic"},
	[]string{"dpc"},
	[]string{"opc"},
	[]string{"msg_name"},
	[]string{"called_number", "num"},
	[]string{"calling_number", "num"},
}

func makeISUPDataHeader(data []byte, sb *bytebufferpool.ByteBuffer) (string, string) {
	var msg_name, called_number, calling_number, callid string
	var cic, dpc, opc int64

	jsonparser.EachKey(data, func(idx int, value []byte, vt jsonparser.ValueType, err error) {
		switch idx {
		case 0:
			if cic, err = jsonparser.ParseInt(value); err != nil {
				logp.Warn("%v", err)
			}
		case 1:
			if dpc, err = jsonparser.ParseInt(value); err != nil {
				logp.Warn("%v", err)
			}
		case 2:
			if opc, err = jsonparser.ParseInt(value); err != nil {
				logp.Warn("%v", err)
			}
		case 3:
			if msg_name, err = jsonparser.ParseString(value); err != nil {
				logp.Warn("%v", err)
			}
		case 4:
			if called_number, err = jsonparser.ParseString(value); err != nil {
				logp.Warn("%v", err)
			}
		case 5:
			if calling_number, err = jsonparser.ParseString(value); err != nil {
				logp.Warn("%v", err)
			}
		}
	}, IsupPaths...)
	scic := strconv.FormatInt(cic, 10)
	sdpc := strconv.FormatInt(dpc, 10)
	sopc := strconv.FormatInt(opc, 10)
	//snprintf("%d:%d:%d", opc < dpc ? opc : dpc, dpc < opc ? opc : dpc , cic)

	if opc < dpc {
		callid = sopc + ":"
	} else {
		callid = sdpc + ":"
	}
	if dpc < opc {
		callid += sopc + ":" + scic
	} else {
		callid += sdpc + ":" + scic
	}

	sb.WriteString(`{`)
	sb.WriteString(`"cic":`)
	sb.WriteString(scic)
	sb.WriteString(`,"dpc":`)
	sb.WriteString(sdpc)
	sb.WriteString(`,"opc":`)
	sb.WriteString(sopc)
	sb.WriteString(`,"msg_name":"`)
	sb.WriteString(msg_name)
	sb.WriteString(`","called_number":"`)
	sb.WriteString(called_number)
	sb.WriteString(`","calling_number":"`)
	sb.WriteString(calling_number)
	sb.WriteString(`","callid":"`)
	sb.WriteString(callid)
	sb.WriteString(`"}`)

	return callid, sb.String()
}
