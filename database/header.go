package database

import (
	"strconv"

	"github.com/buger/jsonparser"
	"github.com/negbie/logp"
	"github.com/sipcapture/heplify-server/decoder"
	"github.com/valyala/bytebufferpool"
	"github.com/valyala/fasttemplate"
)

func makeProtoHeader(hep *decoder.HEP, bb *bytebufferpool.ByteBuffer) string {
	bb.Reset()
	bb.WriteString(`{`)
	bb.WriteString(`"protocolFamily":`)
	bb.WriteString(strconv.FormatUint(uint64(hep.IPVersion), 10))
	bb.WriteString(`,"protocol":`)
	bb.WriteString(strconv.FormatUint(uint64(hep.TransportProto), 10))
	bb.WriteString(`,"srcIp":"`)
	bb.WriteString(hep.SourceIP)
	bb.WriteString(`","dstIp":"`)
	bb.WriteString(hep.DestIP)
	bb.WriteString(`","srcPort":`)
	bb.WriteString(strconv.FormatUint(uint64(hep.SourcePort), 10))
	bb.WriteString(`,"dstPort":`)
	bb.WriteString(strconv.FormatUint(uint64(hep.DestPort), 10))
	bb.WriteString(`,"timeSeconds":`)
	bb.WriteString(strconv.FormatUint(uint64(hep.AssembledTimestamp.Unix()), 10))
	bb.WriteString(`,"timeUseconds":`)
	bb.WriteString(strconv.FormatUint(uint64(hep.AssembledTimestamp.Nanosecond()/1000), 10))
	bb.WriteString(`,"payloadType":`)
	bb.WriteString(strconv.FormatUint(uint64(hep.AppProto), 10))
	bb.WriteString(`,"captureId":"`)
	bb.WriteString(hep.NodeName)
	if hep.NodePW != "" {
		bb.WriteString(`","capturePass":"`)
		bb.WriteString(hep.NodePW)
	}
	bb.WriteString(`","correlation_id":"`)
	decoder.WriteJSONString(bb, hep.CorrelationID)
	bb.WriteString(`"}`)
	return bb.String()
}

func makeSIPDataHeader(hep *decoder.HEP, bb *bytebufferpool.ByteBuffer, t *fasttemplate.Template) string {
	bb.Reset()
	bb.WriteString(`{`)

	t.ExecuteFunc(bb, hep.EscapeFields)

	if len(hep.SIP.CHeader) > 0 {
		for k, v := range hep.SIP.CustomHeader {
			bb.WriteString(`,"` + k + `":"`)
			bb.WriteString(v + `"`)
		}
	}

	bb.WriteString(`}`)
	return bb.String()
}

func makeRTCDataHeader(hep *decoder.HEP, bb *bytebufferpool.ByteBuffer) string {
	bb.Reset()
	bb.WriteString(`{`)
	bb.WriteString(`"node":"`)
	bb.WriteString(hep.NodeName)
	bb.WriteString(`","proto":"`)
	bb.WriteString(hep.ProtoAsString)
	bb.WriteString(`"}`)
	return bb.String()
}

var IsupPaths = [][]string{
	[]string{"cic"}, // Circuit Identification Code
	[]string{"dpc"}, // Destination Point Code
	[]string{"opc"}, // Originating Point Code
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
			// Circuit Identification Code
			if cic, err = jsonparser.ParseInt(value); err != nil {
				logp.Warn("%v", err)
			}
		case 1: 
			// Destination Point Code
			if dpc, err = jsonparser.ParseInt(value); err != nil {
				logp.Warn("%v", err)
			}
		case 2:
			// Originating Point Code
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
