package database

import (
	"bytes"
	"database/sql"
	"fmt"
	"math/rand"
	"net/url"
	"runtime"
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
	db       *sql.DB
	dbDriver string
	bulkCnt  int
	bulkVal  string
}

var queryVal = `(sid,create_date,protocol_header,data_header,raw) VALUES `
var queryValCnt = 5

func (s *SQLHomer7) setup() error {
	var err error
	var addr = strings.Split(config.Setting.DBAddr, ":")
	s.dbDriver = config.Setting.DBDriver

	if len(addr) != 2 {
		err = fmt.Errorf("faulty database address: %v, format should be localhost:3306", config.Setting.DBAddr)
		return err
	}
	if addr[1] == "3306" && s.dbDriver == "postgres" {
		err = fmt.Errorf("don't use port: %s, for db driver: %s", addr[1], s.dbDriver)
		return err
	} else if addr[1] == "5432" && s.dbDriver == "mysql" {
		err = fmt.Errorf("don't use port: %s, for db driver: %s", addr[1], s.dbDriver)
		return err
	}

	if config.Setting.DBRotate {
		b := packr.NewBox("./files/homer7")
		r := NewRotator(&b)
		r.Rotate()
	}

	if s.dbDriver == "mysql" {
		if s.db, err = sql.Open(s.dbDriver, config.Setting.DBUser+":"+config.Setting.DBPass+"@tcp("+addr[0]+":"+addr[1]+")/"+config.Setting.DBDataTable+"?"+url.QueryEscape("charset=utf8mb4&parseTime=true")); err != nil {
			s.db.Close()
			return err
		}
	} else if s.dbDriver == "postgres" {
		if s.db, err = sql.Open(s.dbDriver, "sslmode=disable connect_timeout=2 host="+addr[0]+" port="+addr[1]+" dbname="+config.Setting.DBDataTable+" user="+config.Setting.DBUser+" password="+config.Setting.DBPass); err != nil {
			s.db.Close()
			return err
		}
	}
	if err = s.db.Ping(); err != nil {
		s.db.Close()
		return err
	}

	s.db.SetMaxOpenConns(runtime.NumCPU() * 4)
	s.db.SetMaxIdleConns(runtime.NumCPU())

	s.bulkCnt = config.Setting.DBBulk

	if s.bulkCnt < 1 {
		s.bulkCnt = 1
	}

	s.bulkVal = s.createQueryValues(s.bulkCnt, queryVal)

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
		timer      = config.Setting.DBTimer
	)

	if timer < 0 {
		timer = 0
	}
	rand.Seed(time.Now().UTC().UnixNano())
	tr := rand.Intn(timer+4-3) + 3
	ticker := time.NewTicker(time.Duration(tr) * time.Second)
	if timer == 0 {
		logp.Info("disable timed db inserts")
		ticker.Stop()
	}

	for {
		select {
		case pkt, ok = <-hCh:
			if !ok {
				break
			}

			if pkt.SIP.StartLine == nil {
				logp.Warn("received empty SIP startline inside packet: %v", pkt)
				continue
			}

			date = pkt.Timestamp.Format("2006-01-02 15:04:05.999999")
			pHeader = formProtocolHeader(pkt)
			dHeader = formDataHeader(pkt, date)

			if pkt.ProtoType == 1 && pkt.Payload != "" {
				switch pkt.SIP.CseqMethod {
				case "INVITE", "UPDATE", "BYE", "ACK", "PRACK", "REFER", "CANCEL", "INFO":
					callRows = append(callRows, []interface{}{pkt.CID, date, pHeader, dHeader, pkt.Payload}...)
					callCnt++
					if callCnt == s.bulkCnt {
						s.bulkInsert("call", callRows, s.bulkVal)
						callRows = []interface{}{}
						callCnt = 0
					}
				case "REGISTER":
					regRows = append(regRows, []interface{}{pkt.CID, date, pHeader, dHeader, pkt.Payload}...)
					regCnt++
					if regCnt == s.bulkCnt {
						s.bulkInsert("register", regRows, s.bulkVal)
						regRows = []interface{}{}
						regCnt = 0
					}
				default:
					defRows = append(defRows, []interface{}{pkt.CID, date, pHeader, dHeader, pkt.Payload}...)
					defCnt++
					if defCnt == s.bulkCnt {
						s.bulkInsert("default", defRows, s.bulkVal)
						defRows = []interface{}{}
						defCnt = 0
					}
				}
			} else if pkt.ProtoType >= 2 && pkt.Payload != "" && pkt.CID != "" {
				switch pkt.ProtoType {
				case 5:
					rtcpRows = append(rtcpRows, []interface{}{pkt.CID, date, pHeader, dHeader, pkt.Payload}...)
					rtcpCnt++
					if rtcpCnt == s.bulkCnt {
						s.bulkInsert("rtcp", rtcpRows, s.bulkVal)
						rtcpRows = []interface{}{}
						rtcpCnt = 0
					}
				case 34, 35, 38:
					reportRows = append(reportRows, []interface{}{pkt.CID, date, pHeader, dHeader, pkt.Payload}...)
					reportCnt++
					if reportCnt == s.bulkCnt {
						s.bulkInsert("report", reportRows, s.bulkVal)
						reportRows = []interface{}{}
						reportCnt = 0
					}
				case 53:
					dnsRows = append(dnsRows, []interface{}{pkt.CID, date, pHeader, dHeader, pkt.Payload}...)
					dnsCnt++
					if dnsCnt == s.bulkCnt {
						s.bulkInsert("dns", dnsRows, s.bulkVal)
						dnsRows = []interface{}{}
						dnsCnt = 0
					}
				case 100:
					logRows = append(logRows, []interface{}{pkt.CID, date, pHeader, dHeader, pkt.Payload}...)
					logCnt++
					if logCnt == s.bulkCnt {
						s.bulkInsert("log", logRows, s.bulkVal)
						logRows = []interface{}{}
						logCnt = 0
					}
				}
			}
		case <-ticker.C:
			if callCnt > 1 {
				l := len(callRows)
				s.bulkInsert("call", callRows[:l], s.createQueryValues(l/queryValCnt, queryVal))
				callRows = []interface{}{}
				callCnt = 0
			}
			if regCnt > 1 {
				l := len(regRows)
				s.bulkInsert("register", regRows[:l], s.createQueryValues(l/queryValCnt, queryVal))
				regRows = []interface{}{}
				regCnt = 0
			}
			if defCnt > 1 {
				l := len(defRows)
				s.bulkInsert("default", defRows[:l], s.createQueryValues(l/queryValCnt, queryVal))
				defRows = []interface{}{}
				defCnt = 0
			}
			if rtcpCnt > 1 {
				l := len(rtcpRows)
				s.bulkInsert("rtcp", rtcpRows[:l], s.createQueryValues(l/queryValCnt, queryVal))
				rtcpRows = []interface{}{}
				rtcpCnt = 0
			}
			if reportCnt > 1 {
				l := len(reportRows)
				s.bulkInsert("report", reportRows[:l], s.createQueryValues(l/queryValCnt, queryVal))
				reportRows = []interface{}{}
				reportCnt = 0
			}
			if dnsCnt > 1 {
				l := len(dnsRows)
				s.bulkInsert("dns", dnsRows[:l], s.createQueryValues(l/queryValCnt, queryVal))
				dnsRows = []interface{}{}
				dnsCnt = 0
			}
			if logCnt > 1 {
				l := len(logRows)
				s.bulkInsert("log", logRows[:l], s.createQueryValues(l/queryValCnt, queryVal))
				logRows = []interface{}{}
				logCnt = 0
			}
		}
	}
}

func (s *SQLHomer7) bulkInsert(query string, rows []interface{}, values string) {
	if s.dbDriver == "mysql" {
		tableDate := time.Now().UTC().Format("20060102")
		switch query {
		case "call":
			query = "INSERT INTO hep_proto_1_call_" + tableDate + values
		case "register":
			query = "INSERT INTO hep_proto_1_register_" + tableDate + values
		case "default":
			query = "INSERT INTO hep_proto_1_default_" + tableDate + values
		case "rtcp":
			query = "INSERT INTO hep_proto_5_default_" + tableDate + values
		case "report":
			query = "INSERT INTO hep_proto_35_default_" + tableDate + values
		case "dns":
			query = "INSERT INTO hep_proto_53_default_" + tableDate + values
		case "log":
			query = "INSERT INTO hep_proto_100_default_" + tableDate + values
		}
		_, err := s.db.Exec(query, rows...)
		if err != nil {
			logp.Err("%v", err)
		}
	} else if s.dbDriver == "postgres" {
		switch query {
		case "call":
			query = "COPY hep_proto_1_call(sid,create_date,protocol_header,data_header,raw) FROM STDIN ENCODING 'WIN1252'"
		case "register":
			query = "COPY hep_proto_1_register(sid,create_date,protocol_header,data_header,raw) FROM STDIN ENCODING 'WIN1252'"
		case "default":
			query = "COPY hep_proto_1_default(sid,create_date,protocol_header,data_header,raw) FROM STDIN ENCODING 'WIN1252'"
		case "rtcp":
			query = "COPY hep_proto_5_default(sid,create_date,protocol_header,data_header,raw) FROM STDIN ENCODING 'WIN1252'"
		case "report":
			query = "COPY hep_proto_35_default(sid,create_date,protocol_header,data_header,raw) FROM STDIN ENCODING 'WIN1252'"
		case "dns":
			query = "COPY hep_proto_53_default(sid,create_date,protocol_header,data_header,raw) FROM STDIN ENCODING 'WIN1252'"
		case "log":
			query = "COPY hep_proto_100_default(sid,create_date,protocol_header,data_header,raw) FROM STDIN ENCODING 'WIN1252'"
		}

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
						logp.Err("%s", strconv.Quote(s))
					}
				}
			}
			logp.Err("%v", err)
		}
	}

	logp.Debug("sql", "%s\n\n%v\n\n", query, rows)

}

func (s *SQLHomer7) createQueryValues(count int, values string) string {
	if s.dbDriver == "mysql" {
		for i := 0; i < count; i++ {
			values += `(?,?,?,?,?),`
		}
		values = values[:len(values)-1]
	}
	return values
}

func formProtocolHeader(h *decoder.HEP) []byte {
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

func formDataHeader(h *decoder.HEP, date string) []byte {
	var b bytes.Buffer
	if h.ProtoType == 1 {
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
	} else {
		b.WriteString("{")
		b.WriteString("\"create_date\":\"")
		b.WriteString(date)
		b.WriteString("\",\"sid\":\"")
		b.WriteString(h.CID)
		b.WriteString("\"}")
	}
	return b.Bytes()
}
