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

type SQL struct {
	//dbc     *sql.DB
	dbc        *dbr.Connection
	dbs        *dbr.Session
	sipBulkCnt int
	rtcBulkCnt int
	sipBulkVal string
	rtcBulkVal string
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

	s.sipBulkCnt = config.Setting.DBBulk
	s.rtcBulkCnt = config.Setting.DBBulk / 10

	if s.sipBulkCnt < 1 {
		s.sipBulkCnt = 1
	}
	if s.rtcBulkCnt < 1 {
		s.rtcBulkCnt = 1
	}

	s.sipBulkVal = createSipQueryValues(s.sipBulkCnt, sipVal)
	s.rtcBulkVal = createRtcQueryValues(s.rtcBulkCnt, rtcVal)

	logp.Info("%s output address: %s, bulk size: %d\n", config.Setting.DBDriver, config.Setting.DBAddr, config.Setting.DBBulk)
	return nil
}

func (s *SQL) insert(hCh chan *decoder.HEP) {
	var (
		regCnt, callCnt, dnsCnt, logCnt, rtcpCnt, reportCnt int

		pkt        *decoder.HEP
		ts         string
		tsNano     int64
		ok         bool
		regRows    = make([]interface{}, 0, s.sipBulkCnt)
		callRows   = make([]interface{}, 0, s.sipBulkCnt)
		dnsRows    = make([]interface{}, 0, s.rtcBulkCnt)
		logRows    = make([]interface{}, 0, s.rtcBulkCnt)
		rtcpRows   = make([]interface{}, 0, s.rtcBulkCnt)
		reportRows = make([]interface{}, 0, s.rtcBulkCnt)
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
						short(pkt.SIP.CallID, 120),
						short(pkt.Payload, 3000)}...)

					regCnt++
					if regCnt == s.sipBulkCnt {
						s.bulkInsert("register", regRows, s.sipBulkVal)
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
						short(pkt.SIP.CallID, 120),
						short(pkt.Payload, 3000)}...)

					callCnt++
					if callCnt == s.sipBulkCnt {
						s.bulkInsert("call", callRows, s.sipBulkVal)
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
					if rtcpCnt == s.rtcBulkCnt {
						s.bulkInsert("rtcp", rtcpRows, s.rtcBulkVal)
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
					if reportCnt == s.rtcBulkCnt {
						s.bulkInsert("report", reportRows, s.rtcBulkVal)
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
					if dnsCnt == s.rtcBulkCnt {
						s.bulkInsert("dns", dnsRows, s.rtcBulkVal)
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
					if logCnt == s.rtcBulkCnt {
						s.bulkInsert("log", logRows, s.rtcBulkVal)
						logRows = []interface{}{}
						logCnt = 0
					}
				}
			}
		case <-ticker.C:
			if regCnt > 1 {
				l := len(regRows)
				s.bulkInsert("register", regRows[:l], createSipQueryValues(l/sipValCnt, sipVal))
				regRows = []interface{}{}
				regCnt = 0
			}
			if callCnt > 1 {
				l := len(callRows)
				s.bulkInsert("call", callRows[:l], createSipQueryValues(l/sipValCnt, sipVal))
				callRows = []interface{}{}
				callCnt = 0
			}
			if rtcpCnt > 1 {
				l := len(rtcpRows)
				s.bulkInsert("rtcp", rtcpRows[:l], createRtcQueryValues(l/rtcValCnt, rtcVal))
				rtcpRows = []interface{}{}
				rtcpCnt = 0
			}
			if reportCnt > 1 {
				l := len(reportRows)
				s.bulkInsert("report", reportRows[:l], createRtcQueryValues(l/rtcValCnt, rtcVal))
				reportRows = []interface{}{}
				reportCnt = 0
			}
			if dnsCnt > 1 {
				l := len(dnsRows)
				s.bulkInsert("dns", dnsRows[:l], createRtcQueryValues(l/rtcValCnt, rtcVal))
				dnsRows = []interface{}{}
				dnsCnt = 0
			}
			if logCnt > 1 {
				l := len(logRows)
				s.bulkInsert("log", logRows[:l], createRtcQueryValues(l/rtcValCnt, rtcVal))
				logRows = []interface{}{}
				logCnt = 0
			}
		}
	}
}

func (s *SQL) bulkInsert(query string, rows []interface{}, values string) {
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

func short(s string, i int) string {
	if len(s) > i {
		return s[:i]
	}
	return s
}

func createSipQueryValues(count int, values string) string {
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

func createRtcQueryValues(count int, values string) string {
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
