package database

import (
	"strconv"

	"github.com/buger/jsonparser"
	"github.com/negbie/logp"
	"github.com/sipcapture/heplify-server/decoder"
	"github.com/valyala/bytebufferpool"
	"github.com/valyala/fasttemplate"
)

func makeProtoHeader(h *decoder.HEP, bb *bytebufferpool.ByteBuffer) string {
	bb.Reset()
	bb.WriteString(`{`)
	bb.WriteString(`"protocolFamily":`)
	bb.WriteString(strconv.FormatUint(uint64(h.Version), 10))
	bb.WriteString(`,"protocol":`)
	bb.WriteString(strconv.FormatUint(uint64(h.Protocol), 10))
	bb.WriteString(`,"srcIp":"`)
	bb.WriteString(h.SrcIP)
	bb.WriteString(`","dstIp":"`)
	bb.WriteString(h.DstIP)
	bb.WriteString(`","srcPort":`)
	bb.WriteString(strconv.FormatUint(uint64(h.SrcPort), 10))
	bb.WriteString(`,"dstPort":`)
	bb.WriteString(strconv.FormatUint(uint64(h.DstPort), 10))
	bb.WriteString(`,"timeSeconds":`)
	bb.WriteString(strconv.FormatUint(uint64(h.Timestamp.Unix()), 10))
	bb.WriteString(`,"timeUseconds":`)
	bb.WriteString(strconv.FormatUint(uint64(h.Timestamp.Nanosecond()/1000), 10))
	bb.WriteString(`,"payloadType":`)
	bb.WriteString(strconv.FormatUint(uint64(h.ProtoType), 10))
	bb.WriteString(`,"captureId":"`)
	bb.WriteString(h.NodeName)
	if h.NodePW != "" {
		bb.WriteString(`","capturePass":"`)
		bb.WriteString(h.NodePW)
	}
	bb.WriteString(`","correlation_id":"`)
	decoder.WriteJSONString(bb, h.CID)
	bb.WriteString(`"}`)
	return bb.String()
}

func makeSIPDataHeader(h *decoder.HEP, bb *bytebufferpool.ByteBuffer, t *fasttemplate.Template) string {
	bb.Reset()
	bb.WriteString(`{`)

	t.ExecuteFunc(bb, h.EscapeFields)

	if len(h.SIP.CHeader) > 0 || h.SIP.CustomHeader != nil {
		for k, v := range h.SIP.CustomHeader {
			bb.WriteString(`,"` + k + `":"`)
			bb.WriteString(v + `"`)
		}
	}

	bb.WriteString(`}`)
	return bb.String()
}

func makeRTCDataHeader(h *decoder.HEP, bb *bytebufferpool.ByteBuffer) string {
	bb.Reset()
	bb.WriteString(`{`)
	bb.WriteString(`"node":"`)
	bb.WriteString(h.NodeName)
	bb.WriteString(`","proto":"`)
	bb.WriteString(h.ProtoString)
	bb.WriteString(`"}`)
	return bb.String()
}

var IsupPaths = [][]string{
	[]string{"cic"},
	[]string{"dpc"},
	[]string{"opc"},
	[]string{"msg_name"},
	[]string{"called_number", "num"},
	[]string{"calling_number", "num"},
}

func makeISUPDataHeader(data []byte, bb *bytebufferpool.ByteBuffer) (string, string) {
	bb.Reset()
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

	bb.WriteString(`{`)
	bb.WriteString(`"cic":`)
	bb.WriteString(scic)
	bb.WriteString(`,"dpc":`)
	bb.WriteString(sdpc)
	bb.WriteString(`,"opc":`)
	bb.WriteString(sopc)
	bb.WriteString(`,"msg_name":"`)
	bb.WriteString(msg_name)
	bb.WriteString(`","called_number":"`)
	bb.WriteString(called_number)
	bb.WriteString(`","calling_number":"`)
	bb.WriteString(calling_number)
	bb.WriteString(`","callid":"`)
	bb.WriteString(callid)
	bb.WriteString(`"}`)

	return callid, bb.String()
}
