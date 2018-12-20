package database

import (
	"database/sql"
	"fmt"
	"net/url"
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
	callQuery     = []byte("INSERT INTO sip_capture_call_")
	registerQuery = []byte("INSERT INTO sip_capture_registration_")
	restQuery     = []byte("INSERT INTO sip_capture_rest_")
	rtcpQuery     = []byte("INSERT INTO rtcp_capture_all_")
	reportQuery   = []byte("INSERT INTO report_capture_all_")
	dnsQuery      = []byte("INSERT INTO dns_capture_all_")
	logQuery      = []byte("INSERT INTO logs_capture_all_")

	sipVal = []byte(`(
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
			) VALUES `)

	rtcVal = []byte(`(
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
			) VALUES `)

	sipPlaceholder = []byte("(?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?),")
	sipValCnt      = 39
	rtcPlaceholder = []byte("(?,?,?,?,?,?,?,?,?,?,?,?),")
	rtcValCnt      = 12
)

type SQLHomer5 struct {
	db         *sql.DB
	bulkCnt    int
	sipBulkVal []byte
	rtcBulkVal []byte
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
	} else {
		return fmt.Errorf("homer5 has only mysql support")
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

	s.sipBulkVal = sipQueryVal(s.bulkCnt)
	s.rtcBulkVal = rtcQueryVal(s.bulkCnt)

	logp.Info("%s connection established\n", config.Setting.DBDriver)
	return nil
}

func (s *SQLHomer5) insert(hCh chan *decoder.HEP) {
	var (
		callCnt, regCnt, restCnt, dnsCnt, logCnt, rtcpCnt, reportCnt int

		pkt        *decoder.HEP
		ok         bool
		callRows   = make([]interface{}, 0, s.bulkCnt)
		regRows    = make([]interface{}, 0, s.bulkCnt)
		restRows   = make([]interface{}, 0, s.bulkCnt)
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

	addSIPRow := func(r []interface{}) []interface{} {
		r = append(r, []interface{}{
			pkt.Timestamp.Format("2006-01-02 15:04:05.999999"),
			pkt.Timestamp.UnixNano() / 1000,
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
			pkt.Timestamp.Format("2006-01-02 15:04:05.999999"),
			pkt.Timestamp.UnixNano() / 1000,
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

			if pkt.ProtoType == 1 && pkt.Payload != "" && pkt.SIP != nil {
				switch pkt.SIP.CseqMethod {
				case "INVITE", "UPDATE", "BYE", "ACK", "PRACK", "REFER", "CANCEL", "INFO":
					callRows = addSIPRow(callRows)
					callCnt++
					if callCnt == s.bulkCnt {
						s.bulkInsert(callQuery, s.sipBulkVal, callRows)
						callRows = []interface{}{}
						callCnt = 0
					}
				case "REGISTER":
					regRows = addSIPRow(regRows)
					regCnt++
					if regCnt == s.bulkCnt {
						s.bulkInsert(registerQuery, s.sipBulkVal, regRows)
						regRows = []interface{}{}
						regCnt = 0
					}
				default:
					restRows = addSIPRow(restRows)
					restCnt++
					if restCnt == s.bulkCnt {
						s.bulkInsert(restQuery, s.sipBulkVal, restRows)
						restRows = []interface{}{}
						restCnt = 0
					}

				}
			} else if pkt.ProtoType > 1 && pkt.Payload != "" && pkt.CID != "" {
				switch pkt.ProtoType {
				case 5:
					rtcpRows = addRTCRow(rtcpRows)
					rtcpCnt++
					if rtcpCnt == s.bulkCnt {
						s.bulkInsert(rtcpQuery, s.rtcBulkVal, rtcpRows)
						rtcpRows = []interface{}{}
						rtcpCnt = 0
					}
				case 53:
					dnsRows = addRTCRow(dnsRows)
					dnsCnt++
					if dnsCnt == s.bulkCnt {
						s.bulkInsert(dnsQuery, s.rtcBulkVal, dnsRows)
						dnsRows = []interface{}{}
						dnsCnt = 0
					}
				case 100:
					logRows = addRTCRow(logRows)
					logCnt++
					if logCnt == s.bulkCnt {
						s.bulkInsert(logQuery, s.rtcBulkVal, logRows)
						logRows = []interface{}{}
						logCnt = 0
					}
				default:
					stop()
					timer.Reset(1e9)
					reportRows = addRTCRow(reportRows)
					reportCnt++
					if reportCnt == s.bulkCnt {
						s.bulkInsert(reportQuery, s.rtcBulkVal, reportRows)
						reportRows = []interface{}{}
						reportCnt = 0
					}
				}
			}
		case <-timer.C:
			timer.Reset(maxWait)
			if callCnt > 0 {
				l := len(callRows)
				s.bulkInsert(callQuery, sipQueryVal(l/sipValCnt), callRows[:l])
				callRows = []interface{}{}
				callCnt = 0
			}
			if regCnt > 0 {
				l := len(regRows)
				s.bulkInsert(registerQuery, sipQueryVal(l/sipValCnt), regRows[:l])
				regRows = []interface{}{}
				regCnt = 0
			}
			if restCnt > 0 {
				l := len(restRows)
				s.bulkInsert(restQuery, sipQueryVal(l/sipValCnt), restRows[:l])
				restRows = []interface{}{}
				restCnt = 0
			}
			if rtcpCnt > 0 {
				l := len(rtcpRows)
				s.bulkInsert(rtcpQuery, rtcQueryVal(l/rtcValCnt), rtcpRows[:l])
				rtcpRows = []interface{}{}
				rtcpCnt = 0
			}
			if reportCnt > 0 {
				l := len(reportRows)
				s.bulkInsert(reportQuery, rtcQueryVal(l/rtcValCnt), reportRows[:l])
				reportRows = []interface{}{}
				reportCnt = 0
			}
			if dnsCnt > 0 {
				l := len(dnsRows)
				s.bulkInsert(dnsQuery, rtcQueryVal(l/rtcValCnt), dnsRows[:l])
				dnsRows = []interface{}{}
				dnsCnt = 0
			}
			if logCnt > 0 {
				l := len(logRows)
				s.bulkInsert(logQuery, rtcQueryVal(l/rtcValCnt), logRows[:l])
				logRows = []interface{}{}
				logCnt = 0
			}
		}
	}
}

func (s *SQLHomer5) bulkInsert(q, v []byte, rows []interface{}) {
	tblDate := time.Now().In(time.UTC).AppendFormat(q, "20060102")
	query := make([]byte, 0, len(tblDate)+len(v))
	query = append(tblDate, v...)
	_, err := s.db.Exec(string(query), rows...)
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

func sipQueryVal(c int) []byte {
	out := make([]byte, 0, c*len(sipPlaceholder)+len(sipVal)-1)
	for i := 0; i < c; i++ {
		out = append(out, sipPlaceholder...)
	}
	out = append(sipVal, out...)
	out = out[:len(out)-1]
	return out
}

func rtcQueryVal(c int) []byte {
	out := make([]byte, 0, c*len(rtcPlaceholder)+len(rtcVal)-1)
	for i := 0; i < c; i++ {
		out = append(out, rtcPlaceholder...)
	}
	out = append(rtcVal, out...)
	out = out[:len(out)-1]
	return out
}
