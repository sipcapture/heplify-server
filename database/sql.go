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

var (
	sipQuery = `(
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

	rtcQuery = `(
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
)

type SQL struct {
	//dbc     *sql.DB
	dbc     *dbr.Connection
	dbs     *dbr.Session
	sipBulk int
	rtcBulk int
}

func (s *SQL) setup() error {
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

	s.sipBulk = config.Setting.DBBulk
	s.rtcBulk = config.Setting.DBBulk / 10

	if s.sipBulk < 1 {
		s.sipBulk = 1
	}

	if s.rtcBulk < 1 {
		s.rtcBulk = 1
	}

	for i := 0; i < s.sipBulk; i++ {
		if config.Setting.DBDriver == "mysql" {
			sipQuery += `(?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?),`
		} else if config.Setting.DBDriver == "postgres" {
			sipQuery += fmt.Sprintf("($%d,$%d,$%d,$%d,$%d,$%d,$%d,$%d,$%d,$%d,$%d,$%d,$%d,$%d,$%d,$%d,$%d,$%d,$%d,$%d,$%d,$%d,$%d,$%d,$%d,$%d,$%d,$%d,$%d,$%d,$%d,$%d,$%d,$%d,$%d,$%d,$%d,$%d,$%d),",
				i*39+1, i*39+2, i*39+3, i*39+4, i*39+5, i*39+6, i*39+7, i*39+8, i*39+9, i*39+10, i*39+11, i*39+12, i*39+13, i*39+14, i*39+15, i*39+16, i*39+17, i*39+18, i*39+19, i*39+20, i*39+21,
				i*39+22, i*39+23, i*39+24, i*39+25, i*39+26, i*39+27, i*39+28, i*39+29, i*39+30, i*39+31, i*39+32, i*39+33, i*39+34, i*39+35, i*39+36, i*39+37, i*39+38, i*39+39)
		}
	}
	sipQuery = sipQuery[:len(sipQuery)-1]

	for i := 0; i < s.rtcBulk; i++ {
		if config.Setting.DBDriver == "mysql" {
			rtcQuery += `(?,?,?,?,?,?,?,?,?,?,?,?),`
		} else if config.Setting.DBDriver == "postgres" {
			rtcQuery += fmt.Sprintf("($%d,$%d,$%d,$%d,$%d,$%d,$%d,$%d,$%d,$%d,$%d,$%d),",
				i*12+1, i*12+2, i*12+3, i*12+4, i*12+5, i*12+6, i*12+7, i*12+8, i*12+9, i*12+10, i*12+11, i*12+12)
		}
	}
	rtcQuery = rtcQuery[:len(rtcQuery)-1]

	logp.Info("%s output address: %s, bulk size: %d\n", config.Setting.DBDriver, config.Setting.DBAddr, config.Setting.DBBulk)

	return nil
}

func (s *SQL) insert(hCh chan *decoder.HEP) {
	var (
		pkt        *decoder.HEP
		ts         string
		tsNano     int64
		ok         bool
		regCnt     int
		callCnt    int
		dnsCnt     int
		logCnt     int
		rtcpCnt    int
		reportCnt  int
		regRows    = make([]interface{}, 0, s.sipBulk)
		callRows   = make([]interface{}, 0, s.sipBulk)
		dnsRows    = make([]interface{}, 0, s.rtcBulk)
		logRows    = make([]interface{}, 0, s.rtcBulk)
		rtcpRows   = make([]interface{}, 0, s.rtcBulk)
		reportRows = make([]interface{}, 0, s.rtcBulk)
	)

	for {
		pkt, ok = <-hCh
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
					short(pkt.SIP.CallId, 120),
					pkt.AlegID,
					short(pkt.SIP.Via[0].Via, 256),
					short(pkt.SIP.Via[0].Branch, 80),
					short(pkt.SIP.CseqVal, 25),
					short(pkt.SIP.DiversionVal, 256),
					pkt.SIP.ReasonVal,
					short(pkt.SIP.ContentType, 256),
					short(pkt.SIP.AuthVal, 256),
					short(pkt.SIP.UserAgent, 256),
					pkt.SrcIPString,
					pkt.SrcPort,
					pkt.DstIPString,
					pkt.DstPort,
					pkt.SIP.ContactHost,
					pkt.SIP.ContactPort,
					pkt.Protocol,
					pkt.Version,
					short(pkt.SIP.RTPStatVal, 256),
					pkt.ProtoType,
					pkt.NodeID,
					short(pkt.SIP.CallId, 120),
					short(pkt.Payload, 3000)}...)

				regCnt++
				if regCnt == s.sipBulk {
					s.bulkInsert("register", regRows)
					regRows = []interface{}{}
					regCnt = 0
				}
			} else {
				callRows = append(callRows, []interface{}{
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
					short(pkt.SIP.CallId, 120),
					pkt.AlegID,
					short(pkt.SIP.Via[0].Via, 256),
					short(pkt.SIP.Via[0].Branch, 80),
					short(pkt.SIP.CseqVal, 25),
					short(pkt.SIP.DiversionVal, 256),
					pkt.SIP.ReasonVal,
					short(pkt.SIP.ContentType, 256),
					short(pkt.SIP.AuthVal, 256),
					short(pkt.SIP.UserAgent, 256),
					pkt.SrcIPString,
					pkt.SrcPort,
					pkt.DstIPString,
					pkt.DstPort,
					pkt.SIP.ContactHost,
					pkt.SIP.ContactPort,
					pkt.Protocol,
					pkt.Version,
					short(pkt.SIP.RTPStatVal, 256),
					pkt.ProtoType,
					pkt.NodeID,
					short(pkt.SIP.CallId, 120),
					short(pkt.Payload, 3000)}...)

				callCnt++
				if callCnt == s.sipBulk {
					s.bulkInsert("call", callRows)
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
				if rtcpCnt == s.rtcBulk {
					s.bulkInsert("rtcp", rtcpRows)
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
				if reportCnt == s.rtcBulk {
					s.bulkInsert("report", reportRows)
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
				if dnsCnt == s.rtcBulk {
					s.bulkInsert("dns", dnsRows)
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
				if logCnt == s.rtcBulk {
					s.bulkInsert("log", logRows)
					logRows = []interface{}{}
					logCnt = 0
				}
			}
		}
	}
}

func (s *SQL) bulkInsert(query string, rows []interface{}) {
	if config.Setting.DBDriver == "mysql" {
		switch query {
		case "call":
			query = "INSERT INTO sip_capture_call_" + time.Now().Format("20060102") + sipQuery
		case "register":
			query = "INSERT INTO sip_capture_registration_" + time.Now().Format("20060102") + sipQuery
		case "rtcp":
			query = "INSERT INTO rtcp_capture_all_" + time.Now().Format("20060102") + rtcQuery
		case "report":
			query = "INSERT INTO report_capture_all_" + time.Now().Format("20060102") + rtcQuery
		case "dns":
			query = "INSERT INTO dns_capture_all_" + time.Now().Format("20060102") + rtcQuery
		case "log":
			query = "INSERT INTO logs_capture" + rtcQuery
		}
	} else if config.Setting.DBDriver == "postgres" {
		switch query {
		case "call":
			query = "INSERT INTO sip_capture_call" + sipQuery
		case "register":
			query = "INSERT INTO sip_capture_registration" + sipQuery
		case "rtcp":
			query = "INSERT INTO rtcp_capture_all" + rtcQuery
		case "report":
			query = "INSERT INTO report_capture_all" + rtcQuery
		case "dns":
			query = "INSERT INTO dns_capture_all" + rtcQuery
		case "log":
			query = "INSERT INTO logs_capture_all" + rtcQuery
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

func short(s string, i int) string {
	if len(s) > i {
		return s[:i]
	}
	return s
}

/*
func short(s string, i int) string {
	runes := []rune(s)
	if len(runes) > i {
		return string(runes[:i])
	}
	return s
}
*/
