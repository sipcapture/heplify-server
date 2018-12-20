package database

import (
	"bytes"
	"database/sql"
	"fmt"
	"strconv"
	"strings"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/gobuffalo/packr"
	_ "github.com/lib/pq"
	"github.com/negbie/heplify-server"
	"github.com/negbie/heplify-server/config"
	"github.com/negbie/logp"
)

type SQLHomer7 struct {
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

var queryVal = `(sid,create_date,protocol_header,data_header,raw) VALUES `
var queryValCnt = 5

func (s *SQLHomer7) setup() error {
	cs, err := ConnectString(config.Setting.DBDataTable)
	if err != nil {
		return err
	}

	if config.Setting.DBRotate {
		b := packr.NewBox("./files/homer7")
		r := NewRotator(&b)
		r.Rotate()
	}

	if config.Setting.DBDriver == "mysql" {
		return fmt.Errorf("homer7 has only postgres support")
	} else if config.Setting.DBDriver == "postgres" {
		if s.db, err = sql.Open(config.Setting.DBDriver, cs); err != nil {
			s.db.Close()
			return err
		}
	}
	if err = s.db.Ping(); err != nil {
		s.db.Close()
		return err
	}

	s.db.SetMaxOpenConns(config.Setting.DBWorker * 4)
	s.db.SetMaxIdleConns(config.Setting.DBWorker)

	s.bulkCnt = config.Setting.DBBulk

	if s.bulkCnt < 1 {
		s.bulkCnt = 1
	}

	logp.Info("%s connection established\n", config.Setting.DBDriver)
	return nil
}

func (s *SQLHomer7) insert(hCh chan *decoder.HEP) {
	var (
		callCnt, regCnt, defCnt, dnsCnt, logCnt, rtcpCnt, reportCnt int

		pkt        *decoder.HEP
		date       string
		pHeader    []byte
		dHeader    []byte
		ok         bool
		callRows   = make([]interface{}, 0, s.bulkCnt)
		regRows    = make([]interface{}, 0, s.bulkCnt)
		defRows    = make([]interface{}, 0, s.bulkCnt)
		dnsRows    = make([]interface{}, 0, s.bulkCnt)
		logRows    = make([]interface{}, 0, s.bulkCnt)
		rtcpRows   = make([]interface{}, 0, s.bulkCnt)
		reportRows = make([]interface{}, 0, s.bulkCnt)
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
		case pkt, ok = <-hCh:
			if !ok {
				break
			}

			date = pkt.Timestamp.Format("2006-01-02 15:04:05.999999")
			pHeader = makeProtoHeader(pkt)

			if pkt.ProtoType == 1 && pkt.Payload != "" && pkt.SIP != nil {
				dHeader = makeSIPDataHeader(pkt, date)
				switch pkt.SIP.CseqMethod {
				case "INVITE", "UPDATE", "BYE", "ACK", "PRACK", "REFER", "CANCEL", "INFO":
					callRows = append(callRows, []interface{}{pkt.CID, date, pHeader, dHeader, pkt.Payload}...)
					callCnt++
					if callCnt == s.bulkCnt {
						s.bulkInsert(callCopy, callRows)
						callRows = []interface{}{}
						callCnt = 0
					}
				case "REGISTER":
					regRows = append(regRows, []interface{}{pkt.CID, date, pHeader, dHeader, pkt.Payload}...)
					regCnt++
					if regCnt == s.bulkCnt {
						s.bulkInsert(registerCopy, regRows)
						regRows = []interface{}{}
						regCnt = 0
					}
				default:
					defRows = append(defRows, []interface{}{pkt.CID, date, pHeader, dHeader, pkt.Payload}...)
					defCnt++
					if defCnt == s.bulkCnt {
						s.bulkInsert(defaultCopy, defRows)
						defRows = []interface{}{}
						defCnt = 0
					}
				}
			} else if pkt.ProtoType >= 2 && pkt.Payload != "" && pkt.CID != "" {
				dHeader = makeRTCDataHeader(pkt, date)
				switch pkt.ProtoType {
				case 5:
					rtcpRows = append(rtcpRows, []interface{}{pkt.CID, date, pHeader, dHeader, pkt.Payload}...)
					rtcpCnt++
					if rtcpCnt == s.bulkCnt {
						s.bulkInsert(rtcpCopy, rtcpRows)
						rtcpRows = []interface{}{}
						rtcpCnt = 0
					}
				case 53:
					dnsRows = append(dnsRows, []interface{}{pkt.CID, date, pHeader, dHeader, pkt.Payload}...)
					dnsCnt++
					if dnsCnt == s.bulkCnt {
						s.bulkInsert(dnsCopy, dnsRows)
						dnsRows = []interface{}{}
						dnsCnt = 0
					}
				case 100:
					logRows = append(logRows, []interface{}{pkt.CID, date, pHeader, dHeader, pkt.Payload}...)
					logCnt++
					if logCnt == s.bulkCnt {
						s.bulkInsert(logCopy, logRows)
						logRows = []interface{}{}
						logCnt = 0
					}
				default:
					stop()
					timer.Reset(1e9)
					reportRows = append(reportRows, []interface{}{pkt.CID, date, pHeader, dHeader, pkt.Payload}...)
					reportCnt++
					if reportCnt == s.bulkCnt {
						s.bulkInsert(reportCopy, reportRows)
						reportRows = []interface{}{}
						reportCnt = 0
					}
				}
			}
		case <-timer.C:
			timer.Reset(maxWait)
			if callCnt > 0 {
				l := len(callRows)
				s.bulkInsert(callCopy, callRows[:l])
				callRows = []interface{}{}
				callCnt = 0
			}
			if regCnt > 0 {
				l := len(regRows)
				s.bulkInsert(registerCopy, regRows[:l])
				regRows = []interface{}{}
				regCnt = 0
			}
			if defCnt > 0 {
				l := len(defRows)
				s.bulkInsert(defaultCopy, defRows[:l])
				defRows = []interface{}{}
				defCnt = 0
			}
			if rtcpCnt > 0 {
				l := len(rtcpRows)
				s.bulkInsert(rtcpCopy, rtcpRows[:l])
				rtcpRows = []interface{}{}
				rtcpCnt = 0
			}
			if reportCnt > 0 {
				l := len(reportRows)
				s.bulkInsert(reportCopy, reportRows[:l])
				reportRows = []interface{}{}
				reportCnt = 0
			}
			if dnsCnt > 0 {
				l := len(dnsRows)
				s.bulkInsert(dnsCopy, dnsRows[:l])
				dnsRows = []interface{}{}
				dnsCnt = 0
			}
			if logCnt > 0 {
				l := len(logRows)
				s.bulkInsert(logCopy, logRows[:l])
				logRows = []interface{}{}
				logCnt = 0
			}
		}
	}
}

func (s *SQLHomer7) bulkInsert(query string, rows []interface{}) {
	tx, err := s.db.Begin()
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
		for i := 0; i < len(rows); i++ {
			s, ok := rows[i].(string)
			if ok {
				if strings.Contains(s, "\x00") {
					logp.Err("%q", s)
				}
			}
		}
		logp.Err("%v", err)
	}

	//logp.Debug("sql", "%s\n\n%v\n\n", query, rows)
}

func makeProtoHeader(h *decoder.HEP) []byte {
	var b bytes.Buffer
	b.WriteString("{")
	b.WriteString("\"protocolFamily\":")
	b.WriteString(strconv.Itoa(int(h.Version)))
	b.WriteString(",\"protocol\":")
	b.WriteString(strconv.Itoa(int(h.Protocol)))
	b.WriteString(",\"srcIp\":\"")
	b.WriteString(h.SrcIP)
	b.WriteString("\",\"dstIp\":\"")
	b.WriteString(h.DstIP)
	b.WriteString("\",\"srcPort\":")
	b.WriteString(strconv.Itoa(int(h.SrcPort)))
	b.WriteString(",\"dstPort\":")
	b.WriteString(strconv.Itoa(int(h.DstPort)))
	b.WriteString(",\"timeSeconds\":")
	b.WriteString(strconv.Itoa(int(h.Tsec)))
	b.WriteString(",\"timeUseconds\":")
	b.WriteString(strconv.Itoa(int(h.Tmsec)))
	b.WriteString(",\"payloadType\":")
	b.WriteString(strconv.Itoa(int(h.ProtoType)))
	b.WriteString(",\"captureId\":")
	b.WriteString(strconv.Itoa(int(h.NodeID)))
	b.WriteString(",\"capturePass\":\"")
	b.WriteString(h.NodePW)
	b.WriteString("\"}")
	return b.Bytes()
}

func makeRTCDataHeader(h *decoder.HEP, date string) []byte {
	var b bytes.Buffer
	b.WriteString("{")
	b.WriteString("\"create_date\":\"")
	b.WriteString(date)
	b.WriteString("\",\"sid\":\"")
	b.WriteString(h.CID)
	b.WriteString("\"}")
	return b.Bytes()
}

func makeSIPDataHeader(h *decoder.HEP, date string) []byte {
	var b bytes.Buffer
	b.WriteString("{")
	b.WriteString("\"create_date\":\"")
	b.WriteString(date)
	b.WriteString("\",\"ruri_user\":\"")
	b.WriteString(h.SIP.StartLine.URI.User)
	b.WriteString("\",\"from_user\":\"")
	b.WriteString(h.SIP.FromUser)
	b.WriteString("\",\"to_user\":\"")
	b.WriteString(h.SIP.ToUser)
	b.WriteString("\",\"pid_user\":\"")
	b.WriteString(h.SIP.PaiUser)
	b.WriteString("\",\"auth_user\":\"")
	b.WriteString(h.SIP.AuthUser)
	b.WriteString("\",\"sid\":\"")
	b.WriteString(h.CID)
	b.WriteString("\",\"method\":\"")
	b.WriteString(h.SIP.StartLine.Method)
	b.WriteString("\",\"source_ip\":\"")
	b.WriteString(h.SrcIP)
	b.WriteString("\",\"destination_ip\":\"")
	b.WriteString(h.DstIP)
	b.WriteString("\"}")
	return b.Bytes()
}
