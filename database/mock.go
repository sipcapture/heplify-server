package database

import (
	"sync"
	"time"

	"github.com/negbie/logp"
	"github.com/sipcapture/heplify-server/decoder"
	"github.com/valyala/bytebufferpool"
	"golang.org/x/sync/syncmap"
)

type Mock struct {
	db      *sync.Map
	bulkCnt int
}

func (m *Mock) setup() error {
	m.db = new(syncmap.Map)
	m.bulkCnt = 200
	return nil
}

func (m *Mock) insert(hCh chan *decoder.HEP) {
	callCnt := 0
	callRowsString := make([]string, 0, m.bulkCnt)
	t := buildTemplate()
	bb := bytebufferpool.Get()
	defer bytebufferpool.Put(bb)

	for pkt := range hCh {
		date := pkt.Timestamp.Format(time.RFC3339Nano)
		if pkt.ProtoType == 1 && pkt.Payload != "" && pkt.SIP != nil {
			pHeader := makeProtoHeader(pkt, bb)
			dHeader := makeSIPDataHeader(pkt, bb, t)
			callRowsString = append(callRowsString, pkt.SID, date, pHeader, dHeader, pkt.Payload)
			callCnt++
			if callCnt == m.bulkCnt {
				m.bulkInsert(callCopy, callRowsString)
				callRowsString = []string{}
				callCnt = 0
			}
		}
	}
}

func (m *Mock) bulkInsert(query string, rows []string) {
	logp.Debug("sql", "%s\n\n%v\n\n", query, rows)
	m.db.Store(query, rows)
}
