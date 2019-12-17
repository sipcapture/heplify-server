package database

import (
	"database/sql"
	"time"

	_ "github.com/lib/pq"
	"github.com/negbie/logp"
	"github.com/sipcapture/heplify-server/config"
	"github.com/sipcapture/heplify-server/decoder"
	"github.com/valyala/bytebufferpool"
	"github.com/valyala/fasttemplate"
)

type Postgres struct {
	db              *sql.DB
	dbTimer         time.Duration
	bulkCnt         int
	forceHEPPayload []int
}

const (
	callCopy     = "COPY hep_proto_1_call(sid,create_date,protocol_header,data_header,raw) FROM STDIN"
	registerCopy = "COPY hep_proto_1_registration(sid,create_date,protocol_header,data_header,raw) FROM STDIN"
	defaultCopy  = "COPY hep_proto_1_default(sid,create_date,protocol_header,data_header,raw) FROM STDIN"
	rtcpCopy     = "COPY hep_proto_5_default(sid,create_date,protocol_header,data_header,raw) FROM STDIN"
	reportCopy   = "COPY hep_proto_35_default(sid,create_date,protocol_header,data_header,raw) FROM STDIN"
	dnsCopy      = "COPY hep_proto_53_default(sid,create_date,protocol_header,data_header,raw) FROM STDIN"
	isupCopy     = "COPY hep_proto_54_default(sid,create_date,protocol_header,data_header,raw) FROM STDIN"
	logCopy      = "COPY hep_proto_100_default(sid,create_date,protocol_header,data_header,raw) FROM STDIN"
)

func (p *Postgres) setup() error {
	cs, err := ConnectString(config.Setting.DBDataTable)
	if err != nil {
		return err
	}

	if p.db, err = sql.Open(config.Setting.DBDriver, cs); err != nil {
		p.db.Close()
		return err
	}

	if err = p.db.Ping(); err != nil {
		p.db.Close()
		return err
	}

	p.db.SetMaxOpenConns(config.Setting.DBWorker * 4)
	p.db.SetMaxIdleConns(config.Setting.DBWorker)

	p.bulkCnt = config.Setting.DBBulk

	/* force JSON payload to data header */
	p.forceHEPPayload = config.Setting.ForceHEPPayload

	if p.bulkCnt < 1 {
		p.bulkCnt = 1
	}
	p.dbTimer = time.Duration(config.Setting.DBTimer) * time.Second

	logp.Info("%s connection established\n", config.Setting.DBDriver)
	return nil
}

func (p *Postgres) insert(hCh chan *decoder.HEP) {
	var (
		callCnt, regCnt, defCnt, dnsCnt, logCnt, rtcpCnt, isupCnt, reportCnt int

		callRows   = make([]string, 0, p.bulkCnt)
		regRows    = make([]string, 0, p.bulkCnt)
		defRows    = make([]string, 0, p.bulkCnt)
		dnsRows    = make([]string, 0, p.bulkCnt)
		logRows    = make([]string, 0, p.bulkCnt)
		isupRows   = make([]string, 0, p.bulkCnt)
		rtcpRows   = make([]string, 0, p.bulkCnt)
		reportRows = make([]string, 0, p.bulkCnt)
		maxWait    = p.dbTimer
	)

	timer := time.NewTimer(maxWait)
	stop := func() {
		if !timer.Stop() {
			select {
			case <-timer.C:
			default:
			}
		}
	}
	defer stop()

	var dataTemplate string
	for _, v := range config.Setting.SIPHeader {
		dataTemplate += "\"" + v + "\":\"{{" + v + "}}\","
	}

	if len(dataTemplate) > 0 {
		dataTemplate = dataTemplate[:len(dataTemplate)-1]
	}

	t := fasttemplate.New(dataTemplate, "{{", "}}")

	bb := bytebufferpool.Get()
	defer bytebufferpool.Put(bb)

	for {
		select {
		case pkt, ok := <-hCh:
			if !ok {
				if p.db != nil {
					p.db.Close()
				}
				return
			}

			date := pkt.Timestamp.Format(time.RFC3339Nano)

			if pkt.ProtoType == 1 && pkt.Payload != "" && pkt.SIP != nil {
				pHeader := makeProtoHeader(pkt, bb)
				dHeader := makeSIPDataHeader(pkt, bb, t)
				switch pkt.SIP.CseqMethod {
				case "INVITE", "ACK", "BYE", "CANCEL", "UPDATE", "PRACK", "REFER", "INFO":
					callRows = append(callRows, pkt.SID, date, pHeader, dHeader, pkt.Payload)
					callCnt++
					if callCnt == p.bulkCnt {
						p.bulkInsert(callCopy, callRows)
						callRows = []string{}
						callCnt = 0
					}
				case "REGISTER":
					regRows = append(regRows, pkt.SID, date, pHeader, dHeader, pkt.Payload)
					regCnt++
					if regCnt == p.bulkCnt {
						p.bulkInsert(registerCopy, regRows)
						regRows = []string{}
						regCnt = 0
					}
				default:
					defRows = append(defRows, pkt.SID, date, pHeader, dHeader, pkt.Payload)
					defCnt++
					if defCnt == p.bulkCnt {
						p.bulkInsert(defaultCopy, defRows)
						defRows = []string{}
						defCnt = 0
					}
				}
			} else if pkt.ProtoType == 54 && pkt.Payload != "" {
				pHeader := makeProtoHeader(pkt, bb)
				sid, dHeader := makeISUPDataHeader([]byte(pkt.Payload), bb)

				isupRows = append(isupRows, sid, date, pHeader, dHeader, pkt.Payload)
				isupCnt++
				if isupCnt == p.bulkCnt {
					p.bulkInsert(isupCopy, isupRows)
					isupRows = []string{}
					isupCnt = 0
				}

			} else if pkt.ProtoType >= 2 && pkt.Payload != "" && pkt.CID != "" {
				pHeader := makeProtoHeader(pkt, bb)
				dHeader := makeRTCDataHeader(pkt, bb)
				switch pkt.ProtoType {
				case 5:
					rtcpRows = append(rtcpRows, pkt.CID, date, pHeader, dHeader, pkt.Payload)
					rtcpCnt++
					if rtcpCnt == p.bulkCnt {
						p.bulkInsert(rtcpCopy, rtcpRows)
						rtcpRows = []string{}
						rtcpCnt = 0
					}
				case 53:
					dnsRows = append(dnsRows, pkt.CID, date, pHeader, dHeader, pkt.Payload)
					dnsCnt++
					if dnsCnt == p.bulkCnt {
						p.bulkInsert(dnsCopy, dnsRows)
						dnsRows = []string{}
						dnsCnt = 0
					}
				case 100:
					logRows = append(logRows, pkt.CID, date, pHeader, dHeader, pkt.Payload)
					logCnt++
					if logCnt == p.bulkCnt {
						p.bulkInsert(logCopy, logRows)
						logRows = []string{}
						logCnt = 0
					}
				default:
					stop()
					timer.Reset(1e9)
					var ForcePayload = false

					for _, v := range p.forceHEPPayload {
						if pkt.ProtoType == uint32(v) {
							ForcePayload = true
							break
						}
					}

					if ForcePayload {
						reportRows = append(reportRows, pkt.CID, date, pHeader, pkt.Payload, dHeader)
					} else {
						reportRows = append(reportRows, pkt.CID, date, pHeader, dHeader, pkt.Payload)
					}

					reportCnt++
					if reportCnt == p.bulkCnt {
						p.bulkInsert(reportCopy, reportRows)
						reportRows = []string{}
						reportCnt = 0
					}
				}
			}
		case <-timer.C:
			timer.Reset(maxWait)
			if callCnt > 0 {
				l := len(callRows)
				p.bulkInsert(callCopy, callRows[:l])
				callRows = []string{}
				callCnt = 0
			}
			if regCnt > 0 {
				l := len(regRows)
				p.bulkInsert(registerCopy, regRows[:l])
				regRows = []string{}
				regCnt = 0
			}
			if defCnt > 0 {
				l := len(defRows)
				p.bulkInsert(defaultCopy, defRows[:l])
				defRows = []string{}
				defCnt = 0
			}
			if rtcpCnt > 0 {
				l := len(rtcpRows)
				p.bulkInsert(rtcpCopy, rtcpRows[:l])
				rtcpRows = []string{}
				rtcpCnt = 0
			}
			if reportCnt > 0 {
				l := len(reportRows)
				p.bulkInsert(reportCopy, reportRows[:l])
				reportRows = []string{}
				reportCnt = 0
			}
			if dnsCnt > 0 {
				l := len(dnsRows)
				p.bulkInsert(dnsCopy, dnsRows[:l])
				dnsRows = []string{}
				dnsCnt = 0
			}
			if logCnt > 0 {
				l := len(logRows)
				p.bulkInsert(logCopy, logRows[:l])
				logRows = []string{}
				logCnt = 0
			}
			if isupCnt > 0 {
				l := len(isupRows)
				p.bulkInsert(isupCopy, isupRows[:l])
				isupRows = []string{}
				isupCnt = 0
			}
		}
	}
}

func (p *Postgres) bulkInsert(query string, rows []string) {
	tx, err := p.db.Begin()
	if err != nil || tx == nil {
		logp.Err("%v", err)
		return
	}

	stmt, err := tx.Prepare(query)
	if err != nil {
		logp.Err("%v", err)
		err := tx.Rollback()
		if err != nil {
			logp.Err("%v", err)
		}
		return
	}

	for i := 0; i < len(rows); i = i + 5 {
		_, err = stmt.Exec(rows[i], rows[i+1], rows[i+2], rows[i+3], rows[i+4])
		if err != nil {
			logp.Err("%v", err)
			continue
		}
	}

	_, err = stmt.Exec()
	if err != nil {
		logp.Err("%v", err)
	}
	err = stmt.Close()
	if err != nil {
		logp.Err("%v", err)
	}
	err = tx.Commit()
	if err != nil {
		logp.Err("%v", err)
	}

	//logp.Debug("sql", "%s\n\n%v\n\n", query, rows)
}
