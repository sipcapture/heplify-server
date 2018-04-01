package database

import (
	"bytes"
	"fmt"
	"net/url"
	"strconv"
	"strings"
	"time"

	raven "github.com/getsentry/raven-go"
	_ "github.com/go-sql-driver/mysql"
	"github.com/gobuffalo/packr"
	"github.com/gocraft/dbr"
	_ "github.com/lib/pq"
	"github.com/negbie/heplify-server"
	"github.com/negbie/heplify-server/config"
	"github.com/negbie/heplify-server/logp"
)

type SQLHomer7 struct {
	//dbc     *sql.DB
	dbc     *dbr.Connection
	dbs     *dbr.Session
	bulkCnt int
	bulkVal string
}

var queryVal = `(cid,create_date,protocol_header,data_header,raw) VALUES `
var queryValCnt = 5

func (s *SQLHomer7) setup() error {
	var err error
	var addr = strings.Split(config.Setting.DBAddr, ":")

	if len(addr) != 2 {
		err = fmt.Errorf("faulty database address: %v, format should be localhost:3306", config.Setting.DBAddr)
		return err
	}
	if addr[1] == "3306" && config.Setting.DBDriver == "postgres" {
		err = fmt.Errorf("don't use port: %s, for db driver: %s", addr[1], config.Setting.DBDriver)
		return err
	} else if addr[1] == "5432" && config.Setting.DBDriver == "mysql" {
		err = fmt.Errorf("don't use port: %s, for db driver: %s", addr[1], config.Setting.DBDriver)
		return err
	}

	if config.Setting.DBRotate {
		b := packr.NewBox("./files/homer7")
		r := NewRotator(&b)
		r.Rotate()
	}

	if config.Setting.DBDriver == "mysql" {
		if s.dbc, err = dbr.Open(config.Setting.DBDriver, config.Setting.DBUser+":"+config.Setting.DBPass+"@tcp("+addr[0]+":"+addr[1]+")/"+config.Setting.DBDataTable+"?"+url.QueryEscape("charset=utf8mb4&parseTime=true"), nil); err != nil {
			s.dbc.Close()
			return err
		}
	} else if config.Setting.DBDriver == "postgres" {
		if s.dbc, err = dbr.Open(config.Setting.DBDriver, " host="+addr[0]+" port="+addr[1]+" dbname="+config.Setting.DBDataTable+" user="+config.Setting.DBUser+" password="+config.Setting.DBPass+" sslmode=disable", nil); err != nil {
			s.dbc.Close()
			return err
		}
	}
	if err = s.dbc.Ping(); err != nil {
		s.dbc.Close()
		return err
	}

	s.dbc.SetMaxOpenConns(80)
	s.dbc.SetMaxIdleConns(40)
	s.dbs = s.dbc.NewSession(nil)

	s.bulkCnt = config.Setting.DBBulk

	if s.bulkCnt < 1 {
		s.bulkCnt = 1
	}

	s.bulkVal = s.createQueryValues(s.bulkCnt, queryVal)

	logp.Info("%s output address: %s, bulk size: %d\n", config.Setting.DBDriver, config.Setting.DBAddr, config.Setting.DBBulk)
	return nil
}

func (s *SQLHomer7) insert(hCh chan *decoder.HEP) {
	var (
		regCnt, callCnt, dnsCnt, logCnt, rtcpCnt, reportCnt int

		pkt        *decoder.HEP
		date       string
		pHeader    []byte
		dHeader    []byte
		ok         bool
		regRows    = make([]interface{}, 0, s.bulkCnt)
		callRows   = make([]interface{}, 0, s.bulkCnt)
		dnsRows    = make([]interface{}, 0, s.bulkCnt)
		logRows    = make([]interface{}, 0, s.bulkCnt)
		rtcpRows   = make([]interface{}, 0, s.bulkCnt)
		reportRows = make([]interface{}, 0, s.bulkCnt)
		timer      = config.Setting.DBTimer
	)

	if timer < 0 {
		timer = 0
	}
	ticker := time.NewTicker(time.Duration(timer+1) * time.Second)
	if timer == 0 {
		ticker.Stop()
	}

	for {
		select {
		case pkt, ok = <-hCh:
			if !ok {
				break
			}

			date = pkt.Timestamp.Format("2006-01-02 15:04:05")
			pHeader = formProtocolHeader(pkt)
			dHeader = formDataHeader(pkt, date)

			if pkt.ProtoType == 1 && pkt.Payload != "" && pkt.CID != "" {
				if pkt.SIP.CseqMethod == "REGISTER" {
					regRows = append(regRows, []interface{}{pkt.CID, date, pHeader, dHeader, pkt.Payload}...)
					regCnt++
					if regCnt == s.bulkCnt {
						s.bulkInsert("register", regRows, s.bulkVal)
						regRows = []interface{}{}
						regCnt = 0
					}
				} else {
					callRows = append(callRows, []interface{}{pkt.CID, date, pHeader, dHeader, pkt.Payload}...)
					callCnt++
					if callCnt == s.bulkCnt {
						s.bulkInsert("call", callRows, s.bulkVal)
						callRows = []interface{}{}
						callCnt = 0
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
				case 38:
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
			if regCnt > 1 {
				l := len(regRows)
				s.bulkInsert("register", regRows[:l], s.createQueryValues(l/queryValCnt, queryVal))
				regRows = []interface{}{}
				regCnt = 0
			}
			if callCnt > 1 {
				l := len(callRows)
				s.bulkInsert("call", callRows[:l], s.createQueryValues(l/queryValCnt, queryVal))
				callRows = []interface{}{}
				callCnt = 0
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
	if config.Setting.DBDriver == "mysql" {
		switch query {
		case "call":
			query = "INSERT INTO hep_proto_1_call_" + time.Now().Format("20060102") + values
		case "register":
			query = "INSERT INTO hep_proto_1_register_" + time.Now().Format("20060102") + values
		case "rtcp":
			query = "INSERT INTO hep_proto_5_rtcp_" + time.Now().Format("20060102") + values
		case "report":
			query = "INSERT INTO hep_proto_35_report_" + time.Now().Format("20060102") + values
		case "dns":
			query = "INSERT INTO hep_proto_53_dns_" + time.Now().Format("20060102") + values
		case "log":
			query = "INSERT INTO hep_proto_100_logs_" + time.Now().Format("20060102") + values
		}
	} else if config.Setting.DBDriver == "postgres" {
		switch query {
		case "call":
			query = "INSERT INTO hep_proto_1_call" + values
		case "register":
			query = "INSERT INTO hep_proto_1_register" + values
		case "rtcp":
			query = "INSERT INTO hep_proto_5_rtcp" + values
		case "report":
			query = "INSERT INTO hep_proto_35_report" + values
		case "dns":
			query = "INSERT INTO hep_proto_53_dns" + values
		case "log":
			query = "INSERT INTO hep_proto_100_logs" + values
		}
	}

	logp.Debug("sql", "%s\n\n%v\n\n", query, rows)

	_, err := s.dbs.Exec(query, rows...)
	if err != nil {
		logp.Err("%v", err)
		if config.Setting.SentryDSN != "" {
			raven.CaptureError(err, nil)
		}
	}
}

func (s *SQLHomer7) createQueryValues(count int, values string) string {
	for i := 0; i < count; i++ {
		if config.Setting.DBDriver == "mysql" {
			values += `(?,?,?,?,?),`
		} else if config.Setting.DBDriver == "postgres" {
			values += fmt.Sprintf("($%d,$%d,$%d,$%d,$%d),",
				i*queryValCnt+1, i*queryValCnt+2, i*queryValCnt+3, i*queryValCnt+4, i*queryValCnt+5)
		}
	}
	values = values[:len(values)-1]
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
		b.WriteString("\",\"cid\":\"")
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
		b.WriteString("\",\"cid\":\"")
		b.WriteString(h.CID)
		b.WriteString("\"}")
	}
	return b.Bytes()
}
