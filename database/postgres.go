package database

import (
	"database/sql"
	"time"

	_ "github.com/lib/pq"
	"github.com/negbie/logp"
	"github.com/sipcapture/heplify-server/config"
	"github.com/sipcapture/heplify-server/decoder"
	"github.com/valyala/bytebufferpool"
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

var pgRawTables = []string{
	"hep_proto_1_call",
	"hep_proto_1_registration",
	"hep_proto_1_default",
	"hep_proto_5_default",
	"hep_proto_35_default",
	"hep_proto_53_default",
	"hep_proto_54_default",
	"hep_proto_100_default",
}

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

	p.ensureRawBytea()
	startDBHealthMonitor(p.db, "postgres")

	logp.Info("%s connection established\n", config.Setting.DBDriver)
	return nil
}

// ensureRawBytea migrates existing varchar/text raw columns to bytea for binary-safe storage (#540).
// New installs already get bytea from rotator schema. Best-effort; failures are logged only.
func (p *Postgres) ensureRawBytea() {
	for _, table := range pgRawTables {
		var dataType string
		err := p.db.QueryRow(`
			SELECT data_type FROM information_schema.columns
			WHERE table_schema = current_schema()
			  AND table_name = $1 AND column_name = 'raw'`, table).Scan(&dataType)
		if err != nil {
			continue
		}
		if dataType == "bytea" {
			continue
		}
		q := `ALTER TABLE ` + table + ` ALTER COLUMN raw TYPE bytea USING convert_to(coalesce(raw::text, ''), 'UTF8')`
		if _, err := p.db.Exec(q); err != nil {
			logp.Warn("could not migrate %s.raw to bytea (manual ALTER may be required): %v", table, err)
			continue
		}
		logp.Info("migrated %s.raw to bytea", table)
	}
}

func (p *Postgres) insert(hCh chan *decoder.HEP) {
	var (
		callCnt, regCnt, defCnt, dnsCnt, logCnt, rtcpCnt, isupCnt, reportCnt int

		callRows   = make([]any, 0, p.bulkCnt*5)
		regRows    = make([]any, 0, p.bulkCnt*5)
		defRows    = make([]any, 0, p.bulkCnt*5)
		dnsRows    = make([]any, 0, p.bulkCnt*5)
		logRows    = make([]any, 0, p.bulkCnt*5)
		isupRows   = make([]any, 0, p.bulkCnt*5)
		rtcpRows   = make([]any, 0, p.bulkCnt*5)
		reportRows = make([]any, 0, p.bulkCnt*5)
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

	t := buildTemplate()
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
			raw := []byte(pkt.Payload)

			if pkt.ProtoType == 1 && pkt.Payload != "" && pkt.SIP != nil {
				pHeader := makeProtoHeader(pkt, bb)
				dHeader := makeSIPDataHeader(pkt, bb, t)
				switch pkt.SIP.Profile {
				case "call":
					callRows = append(callRows, pkt.SID, date, pHeader, dHeader, raw)
					callCnt++
					if callCnt == p.bulkCnt {
						p.bulkInsert(callCopy, callRows)
						callRows = callRows[:0]
						callCnt = 0
					}
				case "registration":
					regRows = append(regRows, pkt.SID, date, pHeader, dHeader, raw)
					regCnt++
					if regCnt == p.bulkCnt {
						p.bulkInsert(registerCopy, regRows)
						regRows = regRows[:0]
						regCnt = 0
					}
				default:
					defRows = append(defRows, pkt.SID, date, pHeader, dHeader, raw)
					defCnt++
					if defCnt == p.bulkCnt {
						p.bulkInsert(defaultCopy, defRows)
						defRows = defRows[:0]
						defCnt = 0
					}
				}
			} else if pkt.ProtoType == 54 && pkt.Payload != "" {
				pHeader := makeProtoHeader(pkt, bb)
				sid, dHeader := makeISUPDataHeader([]byte(pkt.Payload), bb)

				isupRows = append(isupRows, sid, date, pHeader, dHeader, raw)
				isupCnt++
				if isupCnt == p.bulkCnt {
					p.bulkInsert(isupCopy, isupRows)
					isupRows = isupRows[:0]
					isupCnt = 0
				}

			} else if pkt.ProtoType >= 2 && pkt.Payload != "" && pkt.CID != "" {
				pHeader := makeProtoHeader(pkt, bb)
				dHeader := makeRTCDataHeader(pkt, bb)
				switch pkt.ProtoType {
				case 5:
					rtcpRows = append(rtcpRows, pkt.CID, date, pHeader, dHeader, raw)
					rtcpCnt++
					if rtcpCnt == p.bulkCnt {
						p.bulkInsert(rtcpCopy, rtcpRows)
						rtcpRows = rtcpRows[:0]
						rtcpCnt = 0
					}
				case 53:
					dnsRows = append(dnsRows, pkt.CID, date, pHeader, dHeader, raw)
					dnsCnt++
					if dnsCnt == p.bulkCnt {
						p.bulkInsert(dnsCopy, dnsRows)
						dnsRows = dnsRows[:0]
						dnsCnt = 0
					}
				case 100:
					logRows = append(logRows, pkt.CID, date, pHeader, dHeader, raw)
					logCnt++
					if logCnt == p.bulkCnt {
						p.bulkInsert(logCopy, logRows)
						logRows = logRows[:0]
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
						reportRows = append(reportRows, pkt.CID, date, pHeader, dHeader, raw)
					}

					reportCnt++
					if reportCnt == p.bulkCnt {
						p.bulkInsert(reportCopy, reportRows)
						reportRows = reportRows[:0]
						reportCnt = 0
					}
				}
			}
		case <-timer.C:
			timer.Reset(maxWait)
			if callCnt > 0 {
				p.bulkInsert(callCopy, callRows)
				callRows = callRows[:0]
				callCnt = 0
			}
			if regCnt > 0 {
				p.bulkInsert(registerCopy, regRows)
				regRows = regRows[:0]
				regCnt = 0
			}
			if defCnt > 0 {
				p.bulkInsert(defaultCopy, defRows)
				defRows = defRows[:0]
				defCnt = 0
			}
			if rtcpCnt > 0 {
				p.bulkInsert(rtcpCopy, rtcpRows)
				rtcpRows = rtcpRows[:0]
				rtcpCnt = 0
			}
			if reportCnt > 0 {
				p.bulkInsert(reportCopy, reportRows)
				reportRows = reportRows[:0]
				reportCnt = 0
			}
			if dnsCnt > 0 {
				p.bulkInsert(dnsCopy, dnsRows)
				dnsRows = dnsRows[:0]
				dnsCnt = 0
			}
			if logCnt > 0 {
				p.bulkInsert(logCopy, logRows)
				logRows = logRows[:0]
				logCnt = 0
			}
			if isupCnt > 0 {
				p.bulkInsert(isupCopy, isupRows)
				isupRows = isupRows[:0]
				isupCnt = 0
			}
		}
	}
}

func (p *Postgres) bulkInsert(query string, rows []any) {
	tx, err := p.db.Begin()
	if err != nil || tx == nil {
		logp.Err("%v", err)
		recordDBInsertError("postgres")
		return
	}

	stmt, err := tx.Prepare(query)
	if err != nil {
		logp.Err("%v", err)
		recordDBInsertError("postgres")
		err := tx.Rollback()
		if err != nil {
			logp.Err("%v", err)
		}
		return
	}

	hadErr := false
	for i := 0; i < len(rows); i = i + 5 {
		_, err = stmt.Exec(rows[i], rows[i+1], rows[i+2], rows[i+3], rows[i+4])
		if err != nil {
			logp.Err("%v", err)
			hadErr = true
			continue
		}
	}

	_, err = stmt.Exec()
	if err != nil {
		logp.Err("%v", err)
		hadErr = true
	}
	err = stmt.Close()
	if err != nil {
		logp.Err("%v", err)
		hadErr = true
	}
	err = tx.Commit()
	if err != nil {
		logp.Err("%v", err)
		hadErr = true
	}

	if hadErr {
		recordDBInsertError("postgres")
	} else {
		recordDBInsertSuccess("postgres")
	}

	logp.Debug("sql", "%s\n\n%v\n\n", query, rows)
}
