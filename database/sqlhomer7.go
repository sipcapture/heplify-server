package database

import (
	"fmt"
	"net/url"
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

func (s *SQLHomer7) setup() error {
	var err error
	addr := strings.Split(config.Setting.DBAddr, ":")

	if len(addr) != 2 {
		err = fmt.Errorf("faulty database address: %v, format should be localhost:3306", config.Setting.DBAddr)
		return err
	}

	if config.Setting.DBRotate {
		b := packr.NewBox("./files")
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

	s.bulkVal = s.createQueryValues(s.bulkCnt, sipVal)

	logp.Info("%s output address: %s, bulk size: %d\n", config.Setting.DBDriver, config.Setting.DBAddr, config.Setting.DBBulk)
	return nil
}

func (s *SQLHomer7) insert(hCh chan *decoder.HEP) {
	var (
		regCnt, callCnt, dnsCnt, logCnt, rtcpCnt, reportCnt int

		pkt        *decoder.HEP
		ts         string
		tsNano     int64
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

			ts = pkt.Timestamp.Format("2006-01-02 15:04:05")
			tsNano = pkt.Timestamp.UnixNano() / 1000

			if pkt.ProtoType == 1 && pkt.Payload != "" && pkt.SIP != nil {

				if pkt.SIP.CseqMethod == "REGISTER" {
					regRows = append(regRows, []interface{}{
						ts,
						tsNano,
						short(pkt.SIP.CallID, 120),
						short(pkt.Payload, 3000)}...)

					regCnt++
					if regCnt == s.bulkCnt {
						s.bulkInsert("register", regRows, s.bulkVal)
						regRows = []interface{}{}
						regCnt = 0
					}
				} else {
					callRows = append(callRows, []interface{}{
						ts,
						tsNano,
						short(pkt.SIP.CallID, 120),
						short(pkt.Payload, 3000)}...)

					callCnt++
					if callCnt == s.bulkCnt {
						s.bulkInsert("call", callRows, s.bulkVal)
						callRows = []interface{}{}
						callCnt = 0
					}
				}
			} else if pkt.ProtoType >= 2 && pkt.ProtoType <= 200 && pkt.CorrelationID != "" {
				switch pkt.ProtoType {
				case 5:
					rtcpRows = append(rtcpRows, []interface{}{
						ts,
						tsNano,
						pkt.CorrelationID,
						pkt.SrcIPString, pkt.SrcPort, pkt.DstIPString, pkt.DstPort,
						pkt.Protocol, pkt.Version, pkt.ProtoType, pkt.NodeID, pkt.Payload}...)

					rtcpCnt++
					if rtcpCnt == s.bulkCnt {
						s.bulkInsert("rtcp", rtcpRows, s.bulkVal)
						rtcpRows = []interface{}{}
						rtcpCnt = 0
					}
				case 38:
					reportRows = append(reportRows, []interface{}{
						ts,
						tsNano,
						pkt.CorrelationID,
						pkt.SrcIPString, pkt.SrcPort, pkt.DstIPString, pkt.DstPort,
						pkt.Protocol, pkt.Version, pkt.ProtoType, pkt.NodeID, pkt.Payload}...)

					reportCnt++
					if reportCnt == s.bulkCnt {
						s.bulkInsert("report", reportRows, s.bulkVal)
						reportRows = []interface{}{}
						reportCnt = 0
					}
				case 53:
					dnsRows = append(dnsRows, []interface{}{
						ts,
						tsNano,
						pkt.CorrelationID,
						pkt.SrcIPString, pkt.SrcPort, pkt.DstIPString, pkt.DstPort,
						pkt.Protocol, pkt.Version, pkt.ProtoType, pkt.NodeID, pkt.Payload}...)

					dnsCnt++
					if dnsCnt == s.bulkCnt {
						s.bulkInsert("dns", dnsRows, s.bulkVal)
						dnsRows = []interface{}{}
						dnsCnt = 0
					}
				case 100:
					logRows = append(logRows, []interface{}{
						ts,
						tsNano,
						pkt.CorrelationID,
						pkt.SrcIPString, pkt.SrcPort, pkt.DstIPString, pkt.DstPort,
						pkt.Protocol, pkt.Version, pkt.ProtoType, pkt.NodeID, pkt.Payload}...)

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
				s.bulkInsert("register", regRows[:l], s.createQueryValues(l/sipValCnt, sipVal))
				regRows = []interface{}{}
				regCnt = 0
			}
			if callCnt > 1 {
				l := len(callRows)
				s.bulkInsert("call", callRows[:l], s.createQueryValues(l/sipValCnt, sipVal))
				callRows = []interface{}{}
				callCnt = 0
			}
			if rtcpCnt > 1 {
				l := len(rtcpRows)
				s.bulkInsert("rtcp", rtcpRows[:l], s.createQueryValues(l/rtcValCnt, rtcVal))
				rtcpRows = []interface{}{}
				rtcpCnt = 0
			}
			if reportCnt > 1 {
				l := len(reportRows)
				s.bulkInsert("report", reportRows[:l], s.createQueryValues(l/rtcValCnt, rtcVal))
				reportRows = []interface{}{}
				reportCnt = 0
			}
			if dnsCnt > 1 {
				l := len(dnsRows)
				s.bulkInsert("dns", dnsRows[:l], s.createQueryValues(l/rtcValCnt, rtcVal))
				dnsRows = []interface{}{}
				dnsCnt = 0
			}
			if logCnt > 1 {
				l := len(logRows)
				s.bulkInsert("log", logRows[:l], s.createQueryValues(l/rtcValCnt, rtcVal))
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
			query = "INSERT INTO sip_capture_call_" + time.Now().Format("20060102") + values
		case "register":
			query = "INSERT INTO sip_capture_registration_" + time.Now().Format("20060102") + values
		case "rtcp":
			query = "INSERT INTO rtcp_capture_all_" + time.Now().Format("20060102") + values
		case "report":
			query = "INSERT INTO report_capture_all_" + time.Now().Format("20060102") + values
		case "dns":
			query = "INSERT INTO dns_capture_all_" + time.Now().Format("20060102") + values
		case "log":
			query = "INSERT INTO logs_capture_all_" + time.Now().Format("20060102") + values
		}
	} else if config.Setting.DBDriver == "postgres" {
		switch query {
		case "call":
			query = "INSERT INTO sip_capture_call" + values
		case "register":
			query = "INSERT INTO sip_capture_registration" + values
		case "rtcp":
			query = "INSERT INTO rtcp_capture" + values
		case "report":
			query = "INSERT INTO report_capture" + values
		case "dns":
			query = "INSERT INTO dns_capture" + values
		case "log":
			query = "INSERT INTO logs_capture" + values
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
			values += `(?,?,?,?,?,?,?,?,?,?,?,?),`
		} else if config.Setting.DBDriver == "postgres" {
			values += fmt.Sprintf("($%d,$%d,$%d,$%d,$%d,$%d,$%d,$%d,$%d,$%d,$%d,$%d),",
				i*rtcValCnt+1, i*rtcValCnt+2, i*rtcValCnt+3, i*rtcValCnt+4, i*rtcValCnt+5, i*rtcValCnt+6,
				i*rtcValCnt+7, i*rtcValCnt+8, i*rtcValCnt+9, i*rtcValCnt+10, i*rtcValCnt+11, i*rtcValCnt+12)
		}
	}
	values = values[:len(values)-1]
	return values
}
