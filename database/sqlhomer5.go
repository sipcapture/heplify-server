package database

import (
	"database/sql"
	"fmt"
	"math/rand"
	"net/url"
	"runtime"
	"strings"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/gobuffalo/packr"
	_ "github.com/lib/pq"
	"github.com/negbie/heplify-server"
	"github.com/negbie/heplify-server/config"
	"github.com/negbie/logp"
)

var (
	sipVal = `(
			date, 
			micro_ts,
			method, 
			reply_reason, 
			ruri,
			ruri_user, 
			ruri_domain,
			from_user, 
			from_domain, 
			from_tag,
			to_user, 
			to_domain, 
			to_tag,
			pid_user, 
			contact_user, 
			auth_user,
			callid, 
			callid_aleg,
			via_1, 
			via_1_branch,
			cseq, 
			diversion,
			reason, 
			content_type,
			auth, 
			user_agent,
			source_ip, 
			source_port,
			destination_ip, 
			destination_port,
			contact_ip, 
			contact_port,
			proto, 
			family, 
			rtp_stat,
			type, 
			node, 
			correlation_id,
			msg
			) VALUES `
	sipValCnt = 39

	rtcVal = `(
			date,
			micro_ts,
			correlation_id,
			source_ip,
			source_port,
			destination_ip,
			destination_port,
			proto,
			family,
			type,
			node,
			msg
			) VALUES `
	rtcValCnt = 12
)

type SQLHomer5 struct {
	db         *sql.DB
	bulkCnt    int
	sipBulkVal string
	rtcBulkVal string
}

func (s *SQLHomer5) setup() error {
	var err error
	addr := strings.Split(config.Setting.DBAddr, ":")

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
		b := packr.NewBox("./files/homer5")
		r := NewRotator(&b)
		r.Rotate()
	}

	if config.Setting.DBDriver == "mysql" {
		if s.db, err = sql.Open(config.Setting.DBDriver, config.Setting.DBUser+":"+config.Setting.DBPass+"@tcp("+addr[0]+":"+addr[1]+")/"+config.Setting.DBDataTable+"?"+url.QueryEscape("charset=utf8mb4&parseTime=true")); err != nil {
			s.db.Close()
			return err
		}
	} else if config.Setting.DBDriver == "postgres" {
		if s.db, err = sql.Open(config.Setting.DBDriver, "sslmode=disable connect_timeout=2 host="+addr[0]+" port="+addr[1]+" dbname="+config.Setting.DBDataTable+" user="+config.Setting.DBUser+" password="+config.Setting.DBPass); err != nil {
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

	s.sipBulkVal = s.createSipQueryValues(s.bulkCnt, sipVal)
	s.rtcBulkVal = s.createRtcQueryValues(s.bulkCnt, rtcVal)

	logp.Info("%s connection established\n", config.Setting.DBDriver)
	return nil
}

func (s *SQLHomer5) insert(hCh chan *decoder.HEP) {
	var (
		callCnt, regCnt, restCnt, dnsCnt, logCnt, rtcpCnt, reportCnt int

		pkt        *decoder.HEP
		ts         string
		tsNano     int64
		ok         bool
		callRows   = make([]interface{}, 0, s.bulkCnt)
		regRows    = make([]interface{}, 0, s.bulkCnt)
		restRows   = make([]interface{}, 0, s.bulkCnt)
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
	tr := rand.Intn(timer+4-timer+1) + (timer + 1)
	ticker := time.NewTicker(time.Duration(tr) * time.Second)
	if timer == 0 {
		ticker.Stop()
	}

	addSIPRow := func(r []interface{}) []interface{} {
		r = append(r, []interface{}{
			ts,
			tsNano,
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

	addRTCRow := func(r []interface{}) []interface{} {
		r = append(r, []interface{}{
			ts,
			tsNano,
			pkt.CID,
			pkt.SrcIP, pkt.SrcPort, pkt.DstIP, pkt.DstPort,
			pkt.Protocol, pkt.Version, pkt.ProtoType, pkt.NodeID, pkt.Payload}...)
		return r
	}

	for {
		select {
		case pkt, ok = <-hCh:
			if !ok {
				break
			}

			ts = pkt.Timestamp.Format("2006-01-02 15:04:05.999999")
			tsNano = pkt.Timestamp.UnixNano() / 1000

			if pkt.ProtoType == 1 && pkt.Payload != "" && pkt.SIP != nil {
				switch pkt.SIP.CseqMethod {
				case "INVITE", "UPDATE", "BYE", "ACK", "PRACK", "REFER", "CANCEL", "INFO":
					callRows = addSIPRow(callRows)
					callCnt++
					if callCnt == s.bulkCnt {
						s.bulkInsert("call", callRows, s.sipBulkVal)
						callRows = []interface{}{}
						callCnt = 0
					}
				case "REGISTER":
					regRows = addSIPRow(regRows)
					regCnt++
					if regCnt == s.bulkCnt {
						s.bulkInsert("register", regRows, s.sipBulkVal)
						regRows = []interface{}{}
						regCnt = 0
					}
				default:
					restRows = addSIPRow(restRows)
					restCnt++
					if restCnt == s.bulkCnt {
						s.bulkInsert("rest", restRows, s.sipBulkVal)
						restRows = []interface{}{}
						restCnt = 0
					}

				}
			} else if pkt.ProtoType >= 2 && pkt.Payload != "" && pkt.CID != "" {
				switch pkt.ProtoType {
				case 5:
					rtcpRows = addRTCRow(rtcpRows)
					rtcpCnt++
					if rtcpCnt == s.bulkCnt {
						s.bulkInsert("rtcp", rtcpRows, s.rtcBulkVal)
						rtcpRows = []interface{}{}
						rtcpCnt = 0
					}
				case 34, 35, 38:
					reportRows = addRTCRow(reportRows)
					reportCnt++
					if reportCnt == s.bulkCnt {
						s.bulkInsert("report", reportRows, s.rtcBulkVal)
						reportRows = []interface{}{}
						reportCnt = 0
					}
				case 53:
					dnsRows = addRTCRow(dnsRows)
					dnsCnt++
					if dnsCnt == s.bulkCnt {
						s.bulkInsert("dns", dnsRows, s.rtcBulkVal)
						dnsRows = []interface{}{}
						dnsCnt = 0
					}
				case 100:
					logRows = addRTCRow(logRows)
					logCnt++
					if logCnt == s.bulkCnt {
						s.bulkInsert("log", logRows, s.rtcBulkVal)
						logRows = []interface{}{}
						logCnt = 0
					}
				}
			}
		case <-ticker.C:
			if callCnt > 1 {
				l := len(callRows)
				s.bulkInsert("call", callRows[:l], s.createSipQueryValues(l/sipValCnt, sipVal))
				callRows = []interface{}{}
				callCnt = 0
			}
			if regCnt > 1 {
				l := len(regRows)
				s.bulkInsert("register", regRows[:l], s.createSipQueryValues(l/sipValCnt, sipVal))
				regRows = []interface{}{}
				regCnt = 0
			}
			if restCnt > 1 {
				l := len(restRows)
				s.bulkInsert("rest", restRows[:l], s.createSipQueryValues(l/sipValCnt, sipVal))
				restRows = []interface{}{}
				restCnt = 0
			}
			if rtcpCnt > 1 {
				l := len(rtcpRows)
				s.bulkInsert("rtcp", rtcpRows[:l], s.createRtcQueryValues(l/rtcValCnt, rtcVal))
				rtcpRows = []interface{}{}
				rtcpCnt = 0
			}
			if reportCnt > 1 {
				l := len(reportRows)
				s.bulkInsert("report", reportRows[:l], s.createRtcQueryValues(l/rtcValCnt, rtcVal))
				reportRows = []interface{}{}
				reportCnt = 0
			}
			if dnsCnt > 1 {
				l := len(dnsRows)
				s.bulkInsert("dns", dnsRows[:l], s.createRtcQueryValues(l/rtcValCnt, rtcVal))
				dnsRows = []interface{}{}
				dnsCnt = 0
			}
			if logCnt > 1 {
				l := len(logRows)
				s.bulkInsert("log", logRows[:l], s.createRtcQueryValues(l/rtcValCnt, rtcVal))
				logRows = []interface{}{}
				logCnt = 0
			}
		}
	}
}

func (s *SQLHomer5) bulkInsert(query string, rows []interface{}, values string) {
	if config.Setting.DBDriver == "mysql" {
		tableDate := time.Now().UTC().Format("20060102")
		switch query {
		case "call":
			query = "INSERT INTO sip_capture_call_" + tableDate + values
		case "register":
			query = "INSERT INTO sip_capture_registration_" + tableDate + values
		case "rest":
			query = "INSERT INTO sip_capture_rest_" + tableDate + values
		case "rtcp":
			query = "INSERT INTO rtcp_capture_all_" + tableDate + values
		case "report":
			query = "INSERT INTO report_capture_all_" + tableDate + values
		case "dns":
			query = "INSERT INTO dns_capture_all_" + tableDate + values
		case "log":
			query = "INSERT INTO logs_capture_all_" + tableDate + values
		}
	} else if config.Setting.DBDriver == "postgres" {
		switch query {
		case "call":
			query = "INSERT INTO sip_capture_call" + values
		case "register":
			query = "INSERT INTO sip_capture_registration" + values
		case "rest":
			query = "INSERT INTO sip_capture_rest" + values
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

	_, err := s.db.Exec(query, rows...)
	if err != nil {
		logp.Err("%v", err)
	}
}

func short(s string, i int) string {
	if len(s) > i {
		return s[:i]
	}
	return s
}

func (s *SQLHomer5) createSipQueryValues(count int, values string) string {
	for i := 0; i < count; i++ {
		if config.Setting.DBDriver == "mysql" {
			values += `(?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?),`
		} else if config.Setting.DBDriver == "postgres" {
			values += fmt.Sprintf("($%d,$%d,$%d,$%d,$%d,$%d,$%d,$%d,$%d,$%d,$%d,$%d,$%d,$%d,$%d,$%d,$%d,$%d,$%d,$%d,$%d,$%d,$%d,$%d,$%d,$%d,$%d,$%d,$%d,$%d,$%d,$%d,$%d,$%d,$%d,$%d,$%d,$%d,$%d),",
				i*sipValCnt+1, i*sipValCnt+2, i*sipValCnt+3, i*sipValCnt+4, i*sipValCnt+5, i*sipValCnt+6, i*sipValCnt+7, i*sipValCnt+8, i*sipValCnt+9, i*sipValCnt+10, i*sipValCnt+11, i*sipValCnt+12,
				i*sipValCnt+13, i*sipValCnt+14, i*sipValCnt+15, i*sipValCnt+16, i*sipValCnt+17, i*sipValCnt+18, i*sipValCnt+19, i*sipValCnt+20, i*sipValCnt+21,
				i*sipValCnt+22, i*sipValCnt+23, i*sipValCnt+24, i*sipValCnt+25, i*sipValCnt+26, i*sipValCnt+27, i*sipValCnt+28, i*sipValCnt+29, i*sipValCnt+30,
				i*sipValCnt+31, i*sipValCnt+32, i*sipValCnt+33, i*sipValCnt+34, i*sipValCnt+35, i*sipValCnt+36, i*sipValCnt+37, i*sipValCnt+38, i*sipValCnt+39)
		}
	}
	values = values[:len(values)-1]
	return values
}

func (s *SQLHomer5) createRtcQueryValues(count int, values string) string {
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
