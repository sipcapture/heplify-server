package database

import (
	"database/sql"
	"fmt"
	"net/url"
	"time"

	_ "github.com/go-sql-driver/mysql"
	_ "github.com/lib/pq"
	"github.com/negbie/heplify-server"
	"github.com/negbie/heplify-server/config"
	"github.com/negbie/heplify-server/logp"
)

var (
	callRegStmt = `(
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

	jsonStmt = `(
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
	//dbc        *dbr.Connection
	//dbs        *dbr.Session
	dbc        *sql.DB
	regRows    []interface{}
	regCnt     int
	callRows   []interface{}
	callCnt    int
	dnsRows    []interface{}
	dnsCnt     int
	logRows    []interface{}
	logCnt     int
	rtcpRows   []interface{}
	rtcpCnt    int
	reportRows []interface{}
	reportCnt  int
}

func (s *SQL) setup() error {
	var err error
	if config.Setting.DBDriver == "mysql" {
		if s.dbc, err = sql.Open(config.Setting.DBDriver, config.Setting.DBUser+":"+config.Setting.DBPassword+"@tcp("+config.Setting.DBAddr+":3306)/"+config.Setting.DBName+"?"+url.QueryEscape("charset=utf8mb4,utf8&parseTime=true")); err != nil {
			logp.Err("%v", err)
			return err
		}
	} else {
		if s.dbc, err = sql.Open(config.Setting.DBDriver, "host="+config.Setting.DBAddr+"port=5432"+"dbname="+config.Setting.DBName+"user="+config.Setting.DBUser+"password="+config.Setting.DBPassword); err != nil {
			logp.Err("%v", err)
			return err
		}
	}

	if err = s.dbc.Ping(); err != nil {
		s.dbc.Close()
		logp.Err("%v", err)
		return err
	}

	//s.dbs = s.dbc.NewSession(nil)

	return nil
}

func (s *SQL) insert(topic string, mCh chan *decoder.HEPPacket, ec *uint64) {
	var (
		pkt *decoder.HEPPacket
		ok  bool
	)

	logp.Info("Run MySQL Output, server: %+v, topic: %s\n", config.Setting.DBAddr, topic)

	for {
		pkt, ok = <-mCh
		if pkt.ProtoType == 5 {
			fmt.Println("HEP: ", string(pkt.CorrelationID))
		}
		if !ok {
			break
		}

		//fmt.Println(pkt.SipMsg.Msg)

		if pkt.ProtoType == 1 && pkt.SipMsg != nil {

			if pkt.SipMsg.Cseq.Method == "REGISTER" {
				s.regRows = append(s.regRows, []interface{}{
					pkt.Timestamp.Format("2006-01-02 15:04:05"),
					pkt.Timestamp.UnixNano() / (int64(time.Microsecond) / int64(time.Nanosecond)),
					pkt.SipMsg.StartLine.Method,
					pkt.SipMsg.StartLine.RespText,
					pkt.SipMsg.StartLine.URI.Raw,
					pkt.SipMsg.StartLine.URI.User,
					pkt.SipMsg.StartLine.URI.Host,
					pkt.SipMsg.From.URI.User,
					pkt.SipMsg.From.URI.Host,
					pkt.SipMsg.From.Tag,
					pkt.SipMsg.To.URI.User,
					pkt.SipMsg.To.URI.Host,
					pkt.SipMsg.To.Tag,
					pkt.SipMsg.PAssertedIdVal,
					pkt.SipMsg.Contact.URI.User,
					pkt.SipMsg.Authorization.Val,
					pkt.SipMsg.CallId,
					"", // TODO CallId-Aleg,
					pkt.SipMsg.Via[0].Via,
					pkt.SipMsg.Via[0].Branch,
					pkt.SipMsg.Cseq.Val,
					"", // TODO diversion,
					"", // TODO reason,
					pkt.SipMsg.ContentType,
					pkt.SipMsg.Authorization.Credentials,
					pkt.SipMsg.UserAgent,
					pkt.SrcIP.String(),
					pkt.SrcPort,
					pkt.DstIP.String(),
					pkt.DstPort,
					pkt.SipMsg.Contact.URI.Host,
					0,  // TODO contact_port,
					"", // TODO originator_ip,
					0,  // TODO originator_port,
					pkt.Protocol,
					pkt.Version,
					pkt.SipMsg.RTPStat.Val,
					pkt.ProtoType,
					pkt.NodeID,
					pkt.SipMsg.CallId,
					pkt.SipMsg.Msg}...)

				s.regCnt++
				if s.regCnt >= 300 {
					s.batchInsert("INSERT INTO sip_capture_registration_", s.regRows)
				}
			} else {
				s.callRows = append(s.callRows, []interface{}{
					pkt.Timestamp.Format("2006-01-02 15:04:05"),
					pkt.Timestamp.UnixNano() / (int64(time.Microsecond) / int64(time.Nanosecond)),
					pkt.SipMsg.StartLine.Method,
					pkt.SipMsg.StartLine.RespText,
					pkt.SipMsg.StartLine.URI.Raw,
					pkt.SipMsg.StartLine.URI.User,
					pkt.SipMsg.StartLine.URI.Host,
					pkt.SipMsg.From.URI.User,
					pkt.SipMsg.From.URI.Host,
					pkt.SipMsg.From.Tag,
					pkt.SipMsg.To.URI.User,
					pkt.SipMsg.To.URI.Host,
					pkt.SipMsg.To.Tag,
					pkt.SipMsg.PAssertedIdVal,
					pkt.SipMsg.Contact.URI.User,
					pkt.SipMsg.Authorization.Val,
					pkt.SipMsg.CallId,
					"", // TODO CallId-Aleg,
					pkt.SipMsg.Via[0].Via,
					pkt.SipMsg.Via[0].Branch,
					pkt.SipMsg.Cseq.Val,
					"", // TODO diversion,
					"", // TODO reason,
					pkt.SipMsg.ContentType,
					pkt.SipMsg.Authorization.Credentials,
					pkt.SipMsg.UserAgent,
					pkt.SrcIP.String(),
					pkt.SrcPort,
					pkt.DstIP.String(),
					pkt.DstPort,
					pkt.SipMsg.Contact.URI.Host,
					0,  // TODO contact_port,
					"", // TODO originator_ip,
					0,  // TODO originator_port,
					pkt.Protocol,
					pkt.Version,
					pkt.SipMsg.RTPStat.Val,
					pkt.ProtoType,
					pkt.NodeID,
					pkt.SipMsg.CallId,
					pkt.SipMsg.Msg}...)

				s.callCnt++
				if s.callCnt >= 300 {
					s.batchInsert("INSERT INTO sip_capture_call_", s.callRows)
				}
			}
		} else if pkt.ProtoType >= 2 && pkt.ProtoType <= 200 && pkt.CorrelationID != nil {
			switch pkt.ProtoType {
			case 5:
				s.rtcpRows = append(s.rtcpRows, []interface{}{
					pkt.Timestamp.Format("2006-01-02 15:04:05"),
					pkt.Timestamp.UnixNano() / (int64(time.Microsecond) / int64(time.Nanosecond)),
					string(pkt.CorrelationID),
					pkt.SrcIP.String(), pkt.SrcPort, pkt.DstIP.String(), pkt.DstPort,
					pkt.Protocol, pkt.Version, pkt.ProtoType, pkt.NodeID, string(pkt.Payload)}...)

				s.rtcpCnt++
				if s.rtcpCnt >= 100 {
					s.batchInsert("INSERT INTO rtcp_capture_all_", s.rtcpRows)
				}
			case 38:
				s.reportRows = append(s.reportRows, []interface{}{
					pkt.Timestamp.Format("2006-01-02 15:04:05"),
					pkt.Timestamp.UnixNano() / (int64(time.Microsecond) / int64(time.Nanosecond)),
					string(pkt.CorrelationID),
					pkt.SrcIP.String(), pkt.SrcPort, pkt.DstIP.String(), pkt.DstPort,
					pkt.Protocol, pkt.Version, pkt.ProtoType, pkt.NodeID, string(pkt.Payload)}...)

				s.reportCnt++
				if s.reportCnt >= 100 {
					s.batchInsert("INSERT INTO report_capture_all_", s.reportRows)
				}
			case 53:
				s.dnsRows = append(s.dnsRows, []interface{}{
					pkt.Timestamp.Format("2006-01-02 15:04:05"),
					pkt.Timestamp.UnixNano() / (int64(time.Microsecond) / int64(time.Nanosecond)),
					string(pkt.CorrelationID),
					pkt.SrcIP.String(), pkt.SrcPort, pkt.DstIP.String(), pkt.DstPort,
					pkt.Protocol, pkt.Version, pkt.ProtoType, pkt.NodeID, string(pkt.Payload)}...)

				s.dnsCnt++
				if s.dnsCnt >= 100 {
					s.batchInsert("INSERT INTO dns_capture_all_", s.dnsRows)
				}
			case 100:
				s.logRows = append(s.logRows, []interface{}{
					pkt.Timestamp.Format("2006-01-02 15:04:05"),
					pkt.Timestamp.UnixNano() / (int64(time.Microsecond) / int64(time.Nanosecond)),
					string(pkt.CorrelationID),
					pkt.SrcIP.String(), pkt.SrcPort, pkt.DstIP.String(), pkt.DstPort,
					pkt.Protocol, pkt.Version, pkt.ProtoType, pkt.NodeID, string(pkt.Payload)}...)

				s.logCnt++
				if s.logCnt >= 100 {
					s.batchInsert("INSERT INTO logs_capture_all_", s.logRows)
				}

			}

		}
	}
}

func (s *SQL) batchInsert(query string, rows []interface{}) {
	switch query {
	case "INSERT INTO sip_capture_call_":

		query = query + time.Now().Format("20060102") + callRegStmt
		for i := 1; i < s.callCnt; i++ {
			query += `,(?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?)`
		}
		//fmt.Println(rows)
		//fmt.Println(query)
		fmt.Println("call count", s.callCnt)
		s.callRows = []interface{}{}
		s.callCnt = 0

	case "INSERT INTO sip_capture_registration_":

		query = query + time.Now().Format("20060102") + callRegStmt
		for i := 1; i < s.regCnt; i++ {
			query += `,(?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?)`
		}
		//fmt.Println(rows)
		//fmt.Println(query)
		fmt.Println("reg count", s.regCnt)
		s.regRows = []interface{}{}
		s.regCnt = 0

	case "INSERT INTO rtcp_capture_all_":
		query = query + time.Now().Format("20060102") + jsonStmt
		for i := 1; i < s.rtcpCnt; i++ {
			query += `,(?,?,?,?,?,?,?,?,?,?,?,?)`
		}

		s.rtcpRows = []interface{}{}
		s.rtcpCnt = 0

	case "INSERT INTO report_capture_all_":
		query = query + time.Now().Format("20060102") + jsonStmt
		for i := 1; i < s.reportCnt; i++ {
			query += `,(?,?,?,?,?,?,?,?,?,?,?,?)`
		}

		s.reportRows = []interface{}{}
		s.reportCnt = 0

	case "INSERT INTO dns_capture_all_":
		query = query + time.Now().Format("20060102") + jsonStmt
		for i := 1; i < s.dnsCnt; i++ {
			query += `,(?,?,?,?,?,?,?,?,?,?,?,?)`
		}

		s.dnsRows = []interface{}{}
		s.dnsCnt = 0

	case "INSERT INTO log_capture_all_":
		query = query + time.Now().Format("20060102") + jsonStmt
		for i := 1; i < s.logCnt; i++ {
			query += `,(?,?,?,?,?,?,?,?,?,?,?,?)`
		}

		s.logRows = []interface{}{}
		s.logCnt = 0
	}

	stmt, _ := s.dbc.Prepare(query)
	_, err := stmt.Exec(rows...)
	if err != nil {
		logp.Err("%v", err)
		//*ec++
	}

	/* 	go func() {
		_, err := s.dbs.InsertBySql(query, rows...).Exec()
		if err != nil {
			logp.Err("%v", err)
			//*ec++
		}
	}() */

}
