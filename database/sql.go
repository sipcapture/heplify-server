package database

import (
	"database/sql"
	"net/url"
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
			originator_ip, 
			originator_port,
			proto, 
			family, 
			rtp_stat,
			type, 
			node, 
			correlation_id,
			msg
			) VALUES (?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?)`

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
			) VALUES (?,?,?,?,?,?,?,?,?,?,?,?)`
)

type SQL struct {
	dbc *sql.DB
}

func (s *SQL) setup() error {
	var err error
	if config.Setting.DBDriver == "mysql" {
		if s.dbc, err = sql.Open(config.Setting.DBDriver, config.Setting.DBUser+":"+config.Setting.DBPassword+"@tcp("+config.Setting.DBAddr+":3306)/"+config.Setting.DBName+"?"+url.QueryEscape("charset=utf8mb4&parseTime=true")); err != nil {
			logp.Err("%v", err)
			return err
		}
	} else {
		if s.dbc, err = sql.Open(config.Setting.DBDriver, "host="+config.Setting.DBAddr+"port=5432"+"dbname="+config.Setting.DBName+"user="+config.Setting.DBUser+"password="+config.Setting.DBPassword); err != nil {
			logp.Err("%v", err)
			return err
		}
	}

	s.dbc.SetMaxIdleConns(10)
	s.dbc.SetMaxOpenConns(10)

	if err = s.dbc.Ping(); err != nil {
		s.dbc.Close()
		logp.Err("%v", err)
		return err
	}

	for i := 1; i < 100; i++ {
		sipQuery += `,(?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?)`
	}
	for i := 1; i < 100; i++ {
		rtcQuery += `,(?,?,?,?,?,?,?,?,?,?,?,?)`
	}

	return nil
}

func (s *SQL) insert(topic string, mCh chan *decoder.HEPPacket, ec *uint64) {
	var (
		pkt        *decoder.HEPPacket
		ok         bool
		regCnt     int
		callCnt    int
		dnsCnt     int
		logCnt     int
		rtcpCnt    int
		reportCnt  int
		regRows    = make([]interface{}, 0, 100)
		callRows   = make([]interface{}, 0, 100)
		dnsRows    = make([]interface{}, 0, 100)
		logRows    = make([]interface{}, 0, 100)
		rtcpRows   = make([]interface{}, 0, 100)
		reportRows = make([]interface{}, 0, 100)
	)

	logp.Info("Run MySQL Output, server: %+v, topic: %s\n", config.Setting.DBAddr, topic)

	for {
		pkt, ok = <-mCh
		if !ok {
			break
		}

		if pkt.ProtoType == 1 && pkt.Payload != "" {

			if pkt.SipMsg.Cseq.Method == "REGISTER" {
				regRows = append(regRows, []interface{}{
					pkt.Timestamp.Format("2006-01-02 15:04:05"),
					pkt.Timestamp.UnixNano() / 1000,
					pkt.SipMsg.StartLine.Method,
					pkt.SipMsg.StartLine.RespText,
					pkt.SipMsg.StartLine.URI.Raw,
					pkt.SipMsg.StartLine.URI.User,
					pkt.SipMsg.StartLine.URI.Host,
					pkt.SipMsg.From.URI.User,
					pkt.SipMsg.From.URI.Host,
					short(pkt.SipMsg.From.Tag, 64),
					pkt.SipMsg.To.URI.User,
					pkt.SipMsg.To.URI.Host,
					short(pkt.SipMsg.To.Tag, 64),
					pkt.SipMsg.PAssertedIdVal,
					pkt.SipMsg.Contact.URI.User,
					"", // TODO auth_user,
					pkt.SipMsg.CallId,
					"", // TODO CallId-Aleg,
					pkt.SipMsg.Via[0].Via,
					pkt.SipMsg.Via[0].Branch,
					pkt.SipMsg.Cseq.Val,
					"", // TODO diversion,
					"", // TODO reason,
					pkt.SipMsg.ContentType,
					short(pkt.SipMsg.Authorization.Val, 256),
					pkt.SipMsg.UserAgent,
					pkt.SrcIP,
					pkt.SrcPort,
					pkt.DstIP,
					pkt.DstPort,
					pkt.SipMsg.Contact.URI.Host,
					0,  // TODO contact_port,
					"", // TODO originator_ip,
					0,  // TODO originator_port,
					pkt.Protocol,
					pkt.Version,
					short(pkt.SipMsg.RTPStat.Val, 256),
					pkt.ProtoType,
					pkt.NodeID,
					pkt.SipMsg.CallId,
					short(pkt.Payload, 3000)}...)

				regCnt++
				if regCnt == 100 {
					s.batchInsert("call", regRows)
					regRows = []interface{}{}
					regCnt = 0
				}
			} else {
				callRows = append(callRows, []interface{}{
					pkt.Timestamp.Format("2006-01-02 15:04:05"),
					pkt.Timestamp.UnixNano() / 1000,
					pkt.SipMsg.StartLine.Method,
					pkt.SipMsg.StartLine.RespText,
					pkt.SipMsg.StartLine.URI.Raw,
					pkt.SipMsg.StartLine.URI.User,
					pkt.SipMsg.StartLine.URI.Host,
					pkt.SipMsg.From.URI.User,
					pkt.SipMsg.From.URI.Host,
					short(pkt.SipMsg.From.Tag, 64),
					pkt.SipMsg.To.URI.User,
					pkt.SipMsg.To.URI.Host,
					short(pkt.SipMsg.To.Tag, 64),
					pkt.SipMsg.PAssertedIdVal,
					pkt.SipMsg.Contact.URI.User,
					"", // TODO auth_user,
					pkt.SipMsg.CallId,
					"", // TODO CallId-Aleg,
					pkt.SipMsg.Via[0].Via,
					pkt.SipMsg.Via[0].Branch,
					pkt.SipMsg.Cseq.Val,
					"", // TODO diversion,
					"", // TODO reason,
					pkt.SipMsg.ContentType,
					short(pkt.SipMsg.Authorization.Val, 256),
					pkt.SipMsg.UserAgent,
					pkt.SrcIP,
					pkt.SrcPort,
					pkt.DstIP,
					pkt.DstPort,
					pkt.SipMsg.Contact.URI.Host,
					0,  // TODO contact_port,
					"", // TODO originator_ip,
					0,  // TODO originator_port,
					pkt.Protocol,
					pkt.Version,
					short(pkt.SipMsg.RTPStat.Val, 256),
					pkt.ProtoType,
					pkt.NodeID,
					pkt.SipMsg.CallId,
					short(pkt.Payload, 3000)}...)

				callCnt++
				if callCnt == 100 {
					s.batchInsert("register", callRows)
					callRows = []interface{}{}
					callCnt = 0
				}
			}
		} else if pkt.ProtoType >= 2 && pkt.ProtoType <= 200 && pkt.CorrelationID != "" {
			switch pkt.ProtoType {
			case 5:
				rtcpRows = append(rtcpRows, []interface{}{
					pkt.Timestamp.Format("2006-01-02 15:04:05"),
					pkt.Timestamp.UnixNano() / 1000,
					pkt.CorrelationID,
					pkt.SrcIP, pkt.SrcPort, pkt.DstIP, pkt.DstPort,
					pkt.Protocol, pkt.Version, pkt.ProtoType, pkt.NodeID, pkt.Payload}...)

				rtcpCnt++
				if rtcpCnt == 100 {
					s.batchInsert("rtcp", rtcpRows)
					rtcpRows = []interface{}{}
					rtcpCnt = 0
				}
			case 38:
				reportRows = append(reportRows, []interface{}{
					pkt.Timestamp.Format("2006-01-02 15:04:05"),
					pkt.Timestamp.UnixNano() / 1000,
					pkt.CorrelationID,
					pkt.SrcIP, pkt.SrcPort, pkt.DstIP, pkt.DstPort,
					pkt.Protocol, pkt.Version, pkt.ProtoType, pkt.NodeID, pkt.Payload}...)

				reportCnt++
				if reportCnt == 100 {
					s.batchInsert("report", reportRows)
					reportRows = []interface{}{}
					reportCnt = 0
				}
			case 53:
				dnsRows = append(dnsRows, []interface{}{
					pkt.Timestamp.Format("2006-01-02 15:04:05"),
					pkt.Timestamp.UnixNano() / 1000,
					pkt.CorrelationID,
					pkt.SrcIP, pkt.SrcPort, pkt.DstIP, pkt.DstPort,
					pkt.Protocol, pkt.Version, pkt.ProtoType, pkt.NodeID, pkt.Payload}...)

				dnsCnt++
				if dnsCnt == 100 {
					s.batchInsert("dns", dnsRows)
					dnsRows = []interface{}{}
					dnsCnt = 0
				}
			case 100:
				logRows = append(logRows, []interface{}{
					pkt.Timestamp.Format("2006-01-02 15:04:05"),
					pkt.Timestamp.UnixNano() / 1000,
					pkt.CorrelationID,
					pkt.SrcIP, pkt.SrcPort, pkt.DstIP, pkt.DstPort,
					pkt.Protocol, pkt.Version, pkt.ProtoType, pkt.NodeID, pkt.Payload}...)

				logCnt++
				if logCnt == 100 {
					s.batchInsert("log", logRows)
					logRows = []interface{}{}
					logCnt = 0
				}
			}
		}
	}
}

func (s *SQL) batchInsert(query string, rows []interface{}) {
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

	_, err := s.dbc.Exec(query, rows...)
	if err != nil {
		logp.Err("%v", err)
		//*ec++
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
