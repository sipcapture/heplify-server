package database

import (
	"database/sql"
	"fmt"
	"net/url"
	"strings"
	"time"

	_ "github.com/go-sql-driver/mysql"
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
	dbc     *sql.DB
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

	if config.Setting.DBDriver == "mysql" {
		if s.dbc, err = sql.Open(config.Setting.DBDriver, config.Setting.DBUser+":"+config.Setting.DBPass+"@tcp("+addr[0]+":"+addr[1]+")/"+config.Setting.DBName+"?"+url.QueryEscape("charset=utf8mb4&parseTime=true")); err != nil {
			return err
		}
	} else {
		if s.dbc, err = sql.Open(config.Setting.DBDriver, "host="+addr[0]+"port="+addr[1]+"dbname="+config.Setting.DBName+"user="+config.Setting.DBUser+"password="+config.Setting.DBPass); err != nil {
			return err
		}
	}

	s.dbc.SetMaxOpenConns(40)
	s.dbc.SetMaxIdleConns(20)

	if err = s.dbc.Ping(); err != nil {
		s.dbc.Close()
		return err
	}

	s.sipBulk = config.Setting.DBBulk
	s.rtcBulk = config.Setting.DBBulk / 10

	if s.sipBulk < 1 {
		s.sipBulk = 1
	}

	if s.rtcBulk < 1 {
		s.rtcBulk = 1
	}

	for i := 0; i < s.sipBulk; i++ {
		sipQuery += `(?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?),`
	}
	sipQuery = sipQuery[:len(sipQuery)-1]

	for i := 0; i < s.rtcBulk; i++ {
		rtcQuery += `(?,?,?,?,?,?,?,?,?,?,?,?),`
	}
	rtcQuery = rtcQuery[:len(rtcQuery)-1]

	logp.Info("%s output address: %s, bulk size: %d\n", config.Setting.DBDriver, config.Setting.DBAddr, config.Setting.DBBulk)

	return nil
}

func (s *SQL) insert(topic string, hCh chan *decoder.HEP, ec *uint64) {
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

			if pkt.SIP.Cseq.Method == "REGISTER" {
				regRows = append(regRows, []interface{}{
					ts,
					tsNano,
					short(pkt.SIP.StartLine.Method, 50),
					short(pkt.SIP.StartLine.RespText, 100),
					short(pkt.SIP.StartLine.URI.Raw, 200),
					short(pkt.SIP.StartLine.URI.User, 100),
					short(pkt.SIP.StartLine.URI.Host, 150),
					short(pkt.SIP.From.URI.User, 100),
					short(pkt.SIP.From.URI.Host, 150),
					short(pkt.SIP.From.Tag, 64),
					short(pkt.SIP.To.URI.User, 100),
					short(pkt.SIP.To.URI.Host, 150),
					short(pkt.SIP.To.Tag, 64),
					short(pkt.SIP.PAssertedIdVal, 100),
					short(pkt.SIP.Contact.URI.User, 120),
					short(pkt.SIP.Authorization.Username, 120),
					short(pkt.SIP.CallId, 120),
					pkt.AlegID,
					short(pkt.SIP.Via[0].Via, 256),
					short(pkt.SIP.Via[0].Branch, 80),
					short(pkt.SIP.Cseq.Val, 25),
					short(pkt.SIP.DiversionVal, 256),
					pkt.SIP.Reason.Cause,
					short(pkt.SIP.ContentType, 256),
					short(pkt.SIP.Authorization.Val, 256),
					short(pkt.SIP.UserAgent, 256),
					pkt.SrcIPString,
					pkt.SrcPort,
					pkt.DstIPString,
					pkt.DstPort,
					pkt.SIP.Contact.URI.Host,
					pkt.SIP.Contact.URI.PortInt,
					pkt.Protocol,
					pkt.Version,
					short(pkt.SIP.RTPStatVal, 256),
					pkt.ProtoType,
					pkt.NodeID,
					short(pkt.SIP.CallId, 120),
					short(pkt.SIP.Msg, 3000)}...)

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
					short(pkt.SIP.From.URI.User, 100),
					short(pkt.SIP.From.URI.Host, 150),
					short(pkt.SIP.From.Tag, 64),
					short(pkt.SIP.To.URI.User, 100),
					short(pkt.SIP.To.URI.Host, 150),
					short(pkt.SIP.To.Tag, 64),
					short(pkt.SIP.PAssertedIdVal, 100),
					short(pkt.SIP.Contact.URI.User, 120),
					short(pkt.SIP.Authorization.Username, 120),
					short(pkt.SIP.CallId, 120),
					pkt.AlegID,
					short(pkt.SIP.Via[0].Via, 256),
					short(pkt.SIP.Via[0].Branch, 80),
					short(pkt.SIP.Cseq.Val, 25),
					short(pkt.SIP.DiversionVal, 256),
					pkt.SIP.Reason.Cause,
					short(pkt.SIP.ContentType, 256),
					short(pkt.SIP.Authorization.Val, 256),
					short(pkt.SIP.UserAgent, 256),
					pkt.SrcIPString,
					pkt.SrcPort,
					pkt.DstIPString,
					pkt.DstPort,
					pkt.SIP.Contact.URI.Host,
					pkt.SIP.Contact.URI.PortInt,
					pkt.Protocol,
					pkt.Version,
					short(pkt.SIP.RTPStatVal, 256),
					pkt.ProtoType,
					pkt.NodeID,
					short(pkt.SIP.CallId, 120),
					short(pkt.SIP.Msg, 3000)}...)

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
		query = "INSERT INTO log_capture_all_" + time.Now().Format("20060102") + rtcQuery
	}

	logp.Debug("sql", "%s\n%#v", query, rows)

	_, err := s.dbc.Exec(query, rows...)
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

/*
func short(s string, i int) string {
	runes := []rune(s)
	if len(runes) > i {
		return string(runes[:i])
	}
	return s
}
*/
