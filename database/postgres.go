package database

import (
	"database/sql"
	"strconv"
	"time"

	"github.com/gobuffalo/packr"
	_ "github.com/lib/pq"
	"github.com/negbie/heplify-server/config"
	"github.com/negbie/heplify-server/decoder"
	"github.com/negbie/logp"
	"github.com/valyala/bytebufferpool"
)

type Postgres struct {
	db      *sql.DB
	bulkCnt int
}

const (
	callCopy     = "COPY hep_proto_1_call(sid,create_date,protocol_header,data_header,raw) FROM STDIN"
	registerCopy = "COPY hep_proto_1_register(sid,create_date,protocol_header,data_header,raw) FROM STDIN"
	defaultCopy  = "COPY hep_proto_1_default(sid,create_date,protocol_header,data_header,raw) FROM STDIN"
	rtcpCopy     = "COPY hep_proto_5_default(sid,create_date,protocol_header,data_header,raw) FROM STDIN"
	reportCopy   = "COPY hep_proto_35_default(sid,create_date,protocol_header,data_header,raw) FROM STDIN"
	dnsCopy      = "COPY hep_proto_53_default(sid,create_date,protocol_header,data_header,raw) FROM STDIN"
	logCopy      = "COPY hep_proto_100_default(sid,create_date,protocol_header,data_header,raw) FROM STDIN"
)

//var queryVal = `(sid,create_date,protocol_header,data_header,raw) VALUES `
//var queryValCnt = 5

func (p *Postgres) setup() error {
	cs, err := connectString(config.Setting.DBDataTable)
	if err != nil {
		return err
	}

	if config.Setting.DBRotate {
		b := packr.NewBox("./files/homer7")
		r := NewRotator(&b)
		r.Rotate()
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

	if p.bulkCnt < 1 {
		p.bulkCnt = 1
	}

	logp.Info("%s connection established\n", config.Setting.DBDriver)
	return nil
}

func (p *Postgres) insert(hCh chan *decoder.HEP) {
	var (
		callCnt, regCnt, defCnt, dnsCnt, logCnt, rtcpCnt, reportCnt int

		callRows   = make([]string, 0, p.bulkCnt)
		regRows    = make([]string, 0, p.bulkCnt)
		defRows    = make([]string, 0, p.bulkCnt)
		dnsRows    = make([]string, 0, p.bulkCnt)
		logRows    = make([]string, 0, p.bulkCnt)
		rtcpRows   = make([]string, 0, p.bulkCnt)
		reportRows = make([]string, 0, p.bulkCnt)
		maxWait    = time.Duration(config.Setting.DBTimer) * time.Second
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

	for {
		select {
		case pkt, ok := <-hCh:
			if !ok {
				break
			}

			date := pkt.Timestamp.Format(time.RFC3339Nano)
			bpp := bytebufferpool.Get()
			bpd := bytebufferpool.Get()

			if pkt.ProtoType == 1 && pkt.Payload != "" && pkt.SIP != nil {
				pHeader := makeProtoHeader(pkt, pkt.SIP.XCallID, bpp)
				dHeader := makeSIPDataHeader(pkt, bpd)
				switch pkt.SIP.CseqMethod {
				case "INVITE", "UPDATE", "BYE", "ACK", "PRACK", "REFER", "CANCEL", "INFO":
					callRows = append(callRows, pkt.CID, date, pHeader, dHeader, pkt.Payload)
					callCnt++
					if callCnt == p.bulkCnt {
						p.bulkInsert(callCopy, callRows)
						callRows = []string{}
						callCnt = 0
					}
				case "REGISTER":
					regRows = append(regRows, pkt.CID, date, pHeader, dHeader, pkt.Payload)
					regCnt++
					if regCnt == p.bulkCnt {
						p.bulkInsert(registerCopy, regRows)
						regRows = []string{}
						regCnt = 0
					}
				default:
					defRows = append(defRows, pkt.CID, date, pHeader, dHeader, pkt.Payload)
					defCnt++
					if defCnt == p.bulkCnt {
						p.bulkInsert(defaultCopy, defRows)
						defRows = []string{}
						defCnt = 0
					}
				}
			} else if pkt.ProtoType >= 2 && pkt.Payload != "" && pkt.CID != "" {
				pHeader := makeProtoHeader(pkt, "", bpp)
				dHeader := makeRTCDataHeader(pkt, bpd)
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
					reportRows = append(reportRows, pkt.CID, date, pHeader, dHeader, pkt.Payload)
					reportCnt++
					if reportCnt == p.bulkCnt {
						p.bulkInsert(reportCopy, reportRows)
						reportRows = []string{}
						reportCnt = 0
					}
				}
			}
			bytebufferpool.Put(bpp)
			bytebufferpool.Put(bpd)
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

func makeProtoHeader(h *decoder.HEP, corrID string, sb *bytebufferpool.ByteBuffer) string {
	sb.WriteString("{")
	sb.WriteString("\"protocolFamily\":")
	sb.WriteString(strconv.FormatUint(uint64(h.Version), 10))
	sb.WriteString(",\"protocol\":")
	sb.WriteString(strconv.FormatUint(uint64(h.Protocol), 10))
	sb.WriteString(",\"srcIp\":\"")
	sb.WriteString(h.SrcIP)
	sb.WriteString("\",\"dstIp\":\"")
	sb.WriteString(h.DstIP)
	sb.WriteString("\",\"srcPort\":")
	sb.WriteString(strconv.FormatUint(uint64(h.SrcPort), 10))
	sb.WriteString(",\"dstPort\":")
	sb.WriteString(strconv.FormatUint(uint64(h.DstPort), 10))
	sb.WriteString(",\"timeSeconds\":")
	sb.WriteString(strconv.FormatUint(uint64(h.Tsec), 10))
	sb.WriteString(",\"timeUseconds\":")
	sb.WriteString(strconv.FormatUint(uint64(h.Tmsec), 10))
	sb.WriteString(",\"payloadType\":")
	sb.WriteString(strconv.FormatUint(uint64(h.ProtoType), 10))
	sb.WriteString(",\"captureId\":")
	sb.WriteString(h.Node)
	sb.WriteString(",\"capturePass\":\"")
	sb.WriteString(h.NodePW)
	if corrID != "" {
		sb.WriteString("\",\"correlation_id\":\"")
		sb.WriteString(corrID)
	}
	sb.WriteString("\"}")
	return sb.String()
}

func makeRTCDataHeader(h *decoder.HEP, sb *bytebufferpool.ByteBuffer) string {
	sb.WriteString("{")
	sb.WriteString("\"host\":\"")
	sb.WriteString(h.Host)
	sb.WriteString("\",\"tag\":\"")
	sb.WriteString(h.Tag)
	sb.WriteString("\"}")
	return sb.String()
}

func makeSIPDataHeader(h *decoder.HEP, sb *bytebufferpool.ByteBuffer) string {
	sb.WriteString("{")
	sb.WriteString("\"ruri_domain\":\"")
	sb.WriteString(h.SIP.URIHost)
	sb.WriteString("\",\"ruri_user\":\"")
	sb.WriteString(h.SIP.URIUser)
	sb.WriteString("\",\"from_user\":\"")
	sb.WriteString(h.SIP.FromUser)
	sb.WriteString("\",\"to_user\":\"")
	sb.WriteString(h.SIP.ToUser)
	sb.WriteString("\",\"pid_user\":\"")
	sb.WriteString(h.SIP.PaiUser)
	sb.WriteString("\",\"auth_user\":\"")
	sb.WriteString(h.SIP.AuthUser)
	sb.WriteString("\",\"callid\":\"")
	sb.WriteString(h.SIP.CallID)
	sb.WriteString("\",\"method\":\"")
	sb.WriteString(h.SIP.FirstMethod)
	sb.WriteString("\",\"from_tag\":\"")
	sb.WriteString(h.SIP.FromTag)
	sb.WriteString("\",\"to_tag\":\"")
	sb.WriteString(h.SIP.ToTag)
	sb.WriteString("\"}")
	return sb.String()
}
