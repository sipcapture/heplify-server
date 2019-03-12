package database

import (
	"sync"

	"github.com/negbie/heplify-server/config"
	"github.com/negbie/heplify-server/decoder"
	"github.com/valyala/bytebufferpool"
)

type Mock struct {
	db         map[string]interface{}
	bulkCnt    int
	sipBulkVal []byte
}

func (m *Mock) setup(wg *sync.WaitGroup) error {
	m.db = make(map[string]interface{})
	m.bulkCnt = config.Setting.DBBulk
	m.sipBulkVal = sipQueryVal(m.bulkCnt)
	return nil
}

func (m *Mock) insert(hCh chan *decoder.HEP) {
	var (
		pkt             *decoder.HEP
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
			short(pkt.SIP.FirstMethod, 50),
			short(pkt.SIP.FirstRespText, 100),
			short(pkt.SIP.URIRaw, 200),
			short(pkt.SIP.URIUser, 100),
			short(pkt.SIP.URIHost, 150),
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
			short(pkt.CID, 120),
			short(pkt.Payload, 3000)}...)
		return r
	}

	for pkt := range hCh {
		//date := pkt.Timestamp.Format("2006-01-02 15:04:05.999999")
		date := pkt.Timestamp.String()
		bpp := bytebufferpool.Get()
		bpd := bytebufferpool.Get()
		if pkt.ProtoType == 1 && pkt.Payload != "" && pkt.SIP != nil {
			pHeader := makeProtoHeader(pkt, pkt.SIP.XCallID, bpp)
			dHeader := makeSIPDataHeader(pkt, bpd)
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

func (m *Mock) bulkInsert(query string, rows []interface{}) {
	m.db[query] = rows
}

func (m *Mock) bulkInsertString(query string, rows []string) {
	m.db[query] = rows
}
