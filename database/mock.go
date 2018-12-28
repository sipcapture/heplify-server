package database

import (
	"strconv"

	"github.com/negbie/heplify-server"
	"github.com/negbie/heplify-server/config"
	"github.com/valyala/bytebufferpool"
)

type Mock struct {
	db         map[string]interface{}
	bulkCnt    int
	sipBulkVal []byte
}

func (m *Mock) setup() error {
	m.db = make(map[string]interface{})
	m.bulkCnt = config.Setting.DBBulk
	m.sipBulkVal = sipQueryVal(m.bulkCnt)
	return nil
}

func (m *Mock) insert(hCh chan *decoder.HEP) {
	var (
		pkt             *decoder.HEP
		ok              bool
		callCnt, defCnt int
		callRows        = make([]interface{}, 0, m.bulkCnt)
		defRows         = make([]interface{}, 0, m.bulkCnt)
		_               = defRows
		defRowsString   = make([]string, 0, m.bulkCnt)
		_               = defRowsString
	)

	addSIPRow := func(r []interface{}) []interface{} {
		r = append(r, []interface{}{
			pkt.Timestamp.Format("2006-01-02 15:04:05.999999"),
			pkt.Timestamp.UnixNano() / 1000,
			short(pkt.SIP.StartLine.Method, 50),
			short(pkt.SIP.StartLine.RespText, 100),
			short(pkt.SIP.StartLine.URI.Raw, 200),
			short(pkt.SIP.StartLine.URI.User, 100),
			short(pkt.SIP.StartLine.URI.Host, 150),
			short(pkt.SIP.FromUser, 100),
			short(pkt.SIP.FromHost, 150),
			short(pkt.SIP.FromTag, 64),
			short(pkt.SIP.ToUser, 100),
			short(pkt.SIP.ToHost, 150),
			short(pkt.SIP.ToTag, 64),
			short(pkt.SIP.PaiUser, 100),
			short(pkt.SIP.ContactUser, 120),
			short(pkt.SIP.AuthUser, 120),
			short(pkt.SIP.CallID, 120),
			short(pkt.SIP.XCallID, 120),
			short(pkt.SIP.ViaOne, 256),
			short(pkt.SIP.ViaOneBranch, 80),
			short(pkt.SIP.CseqVal, 25),
			short(pkt.SIP.DiversionVal, 256),
			pkt.SIP.ReasonVal,
			short(pkt.SIP.ContentType, 256),
			short(pkt.SIP.AuthVal, 256),
			short(pkt.SIP.UserAgent, 256),
			pkt.SrcIP,
			pkt.SrcPort,
			pkt.DstIP,
			pkt.DstPort,
			pkt.SIP.ContactHost,
			pkt.SIP.ContactPort,
			pkt.Protocol,
			pkt.Version,
			short(pkt.SIP.RTPStatVal, 256),
			pkt.ProtoType,
			pkt.NodeID,
			short(pkt.SIP.CallID, 120),
			short(pkt.Payload, 3000)}...)
		return r
	}

	for {
		select {
		case pkt, ok = <-hCh:
			if !ok {
				break
			}

			date := pkt.Timestamp.Format("2006-01-02 15:04:05.999999")
			bpp := bytebufferpool.Get()
			bpd := bytebufferpool.Get()
			if pkt.ProtoType == 1 && pkt.Payload != "" && pkt.SIP != nil {
				//pHeader := makeProtoHeader(pkt, pkt.SIP.XCallID)
				//dHeader := makeSIPDataHeader(pkt, date)

				pHeader := makeProtoHeaderString(pkt, pkt.SIP.XCallID, bpp)
				dHeader := makeSIPDataHeaderString(pkt, date, bpd)
				switch pkt.SIP.CseqMethod {
				case "INVITE":
					callRows = addSIPRow(callRows)
					callCnt++
					if callCnt == m.bulkCnt {
						m.bulkInsert(callCopy, callRows)
						callRows = []interface{}{}
						callCnt = 0
					}

				default:
					//defRows = append(defRows, []interface{}{pkt.CID, date, pHeader, dHeader, pkt.Payload}...)
					defRowsString = append(defRowsString, pkt.CID, date, pHeader, dHeader, pkt.Payload)

					defCnt++
					if defCnt == m.bulkCnt {
						//m.bulkInsert(defaultCopy, defRows)
						m.bulkInsertString(defaultCopy, defRowsString)

						defRows = []interface{}{}
						defRowsString = []string{}
						defCnt = 0
					}
				}
			}
			bytebufferpool.Put(bpp)
			bytebufferpool.Put(bpd)
		}
	}
}

func (m *Mock) bulkInsert(query string, rows []interface{}) {
	m.db[query] = rows
}

func (m *Mock) bulkInsertString(query string, rows []string) {
	m.db[query] = rows
}

func makeProtoHeaderString(h *decoder.HEP, corrID string, sb *bytebufferpool.ByteBuffer) string {
	sb.WriteString("{")
	sb.WriteString("\"protocolFamily\":")
	sb.WriteString(strconv.Itoa(int(h.Version)))
	sb.WriteString(",\"protocol\":")
	sb.WriteString(strconv.Itoa(int(h.Protocol)))
	sb.WriteString(",\"srcIp\":\"")
	sb.WriteString(h.SrcIP)
	sb.WriteString("\",\"dstIp\":\"")
	sb.WriteString(h.DstIP)
	sb.WriteString("\",\"srcPort\":")
	sb.WriteString(strconv.Itoa(int(h.SrcPort)))
	sb.WriteString(",\"dstPort\":")
	sb.WriteString(strconv.Itoa(int(h.DstPort)))
	sb.WriteString(",\"timeSeconds\":")
	sb.WriteString(strconv.Itoa(int(h.Tsec)))
	sb.WriteString(",\"timeUseconds\":")
	sb.WriteString(strconv.Itoa(int(h.Tmsec)))
	sb.WriteString(",\"payloadType\":")
	sb.WriteString(strconv.Itoa(int(h.ProtoType)))
	sb.WriteString(",\"captureId\":")
	sb.WriteString(strconv.Itoa(int(h.NodeID)))
	sb.WriteString(",\"capturePass\":\"")
	sb.WriteString(h.NodePW)
	if corrID != "" {
		sb.WriteString("\",\"correlation_id\":\"")
		sb.WriteString(corrID)
	}
	sb.WriteString("\"}")
	return sb.String()
}

func makeSIPDataHeaderString(h *decoder.HEP, date string, sb *bytebufferpool.ByteBuffer) string {
	sb.WriteString("{")
	sb.WriteString("\"create_date\":\"")
	sb.WriteString(date)
	sb.WriteString("\",\"ruri_user\":\"")
	sb.WriteString(h.SIP.StartLine.URI.User)
	sb.WriteString("\",\"from_user\":\"")
	sb.WriteString(h.SIP.FromUser)
	sb.WriteString("\",\"to_user\":\"")
	sb.WriteString(h.SIP.ToUser)
	sb.WriteString("\",\"pid_user\":\"")
	sb.WriteString(h.SIP.PaiUser)
	sb.WriteString("\",\"auth_user\":\"")
	sb.WriteString(h.SIP.AuthUser)
	sb.WriteString("\",\"sid\":\"")
	sb.WriteString(h.CID)
	sb.WriteString("\",\"method\":\"")
	sb.WriteString(h.SIP.StartLine.Method)
	sb.WriteString("\",\"source_ip\":\"")
	sb.WriteString(h.SrcIP)
	sb.WriteString("\",\"destination_ip\":\"")
	sb.WriteString(h.DstIP)
	sb.WriteString("\"}")
	return sb.String()
}
