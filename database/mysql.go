package database

import (
	"database/sql"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/negbie/logp"
	"github.com/sipcapture/heplify-server/config"
	"github.com/sipcapture/heplify-server/decoder"
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

type MySQL struct {
	db         *sql.DB
	bulkCnt    int
	dbTimer    time.Duration
	sipBulkVal []byte
	rtcBulkVal []byte
}

func (m *MySQL) setup() error {
	cs, err := ConnectString(config.Setting.DBDataTable)
	if err != nil {
		return err
	}

	if m.db, err = sql.Open(config.Setting.DBDriver, cs); err != nil {
		m.db.Close()
		return err
	}

	if err = m.db.Ping(); err != nil {
		m.db.Close()
		return err
	}

	m.db.SetMaxOpenConns(config.Setting.DBWorker * 4)
	m.db.SetMaxIdleConns(config.Setting.DBWorker)

	m.bulkCnt = config.Setting.DBBulk
	if m.bulkCnt < 1 {
		m.bulkCnt = 1
	}
	m.dbTimer = time.Duration(config.Setting.DBTimer) * time.Second

	m.sipBulkVal = sipQueryVal(m.bulkCnt)
	m.rtcBulkVal = rtcQueryVal(m.bulkCnt)

	logp.Info("%s connection established\n", config.Setting.DBDriver)
	return nil
}

func (m *MySQL) insert(hCh chan *decoder.HEP) {
	var (
		callCnt, regCnt, restCnt, dnsCnt, logCnt, rtcpCnt, reportCnt int

		pkt        *decoder.HEP
		ok         bool
		callRows   = make([]interface{}, 0, m.bulkCnt)
		regRows    = make([]interface{}, 0, m.bulkCnt)
		restRows   = make([]interface{}, 0, m.bulkCnt)
		dnsRows    = make([]interface{}, 0, m.bulkCnt)
		logRows    = make([]interface{}, 0, m.bulkCnt)
		rtcpRows   = make([]interface{}, 0, m.bulkCnt)
		reportRows = make([]interface{}, 0, m.bulkCnt)
		maxWait    = m.dbTimer
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
			short(pkt.SIP.FirstMethod, 50),
			short(pkt.SIP.FirstRespText, 100),
			short(pkt.SIP.URIRaw, 200),
			short(pkt.SIP.URIUser, 100),
			short(pkt.SIP.URIHost, 150),
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
			short(pkt.CID, 120),
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
				if m.db != nil {
					m.db.Close()
				}
				return
			}

			if pkt.ProtoType == 1 && pkt.Payload != "" && pkt.SIP != nil {
				switch pkt.SIP.CseqMethod {
				case "INVITE", "UPDATE", "BYE", "ACK", "PRACK", "REFER", "CANCEL", "INFO":
					callRows = addSIPRow(callRows)
					callCnt++
					if callCnt == m.bulkCnt {
						m.bulkInsert(callQuery, m.sipBulkVal, callRows)
						callRows = []interface{}{}
						callCnt = 0
					}
				case "REGISTER":
					regRows = addSIPRow(regRows)
					regCnt++
					if regCnt == m.bulkCnt {
						m.bulkInsert(registerQuery, m.sipBulkVal, regRows)
						regRows = []interface{}{}
						regCnt = 0
					}
				default:
					restRows = addSIPRow(restRows)
					restCnt++
					if restCnt == m.bulkCnt {
						m.bulkInsert(restQuery, m.sipBulkVal, restRows)
						restRows = []interface{}{}
						restCnt = 0
					}

				}
			} else if pkt.ProtoType > 1 && pkt.Payload != "" && pkt.CID != "" {
				switch pkt.ProtoType {
				case 5:
					rtcpRows = addRTCRow(rtcpRows)
					rtcpCnt++
					if rtcpCnt == m.bulkCnt {
						m.bulkInsert(rtcpQuery, m.rtcBulkVal, rtcpRows)
						rtcpRows = []interface{}{}
						rtcpCnt = 0
					}
				case 53:
					dnsRows = addRTCRow(dnsRows)
					dnsCnt++
					if dnsCnt == m.bulkCnt {
						m.bulkInsert(dnsQuery, m.rtcBulkVal, dnsRows)
						dnsRows = []interface{}{}
						dnsCnt = 0
					}
				case 100:
					logRows = addRTCRow(logRows)
					logCnt++
					if logCnt == m.bulkCnt {
						m.bulkInsert(logQuery, m.rtcBulkVal, logRows)
						logRows = []interface{}{}
						logCnt = 0
					}
				default:
					stop()
					timer.Reset(1e9)
					reportRows = addRTCRow(reportRows)
					reportCnt++
					if reportCnt == m.bulkCnt {
						m.bulkInsert(reportQuery, m.rtcBulkVal, reportRows)
						reportRows = []interface{}{}
						reportCnt = 0
					}
				}
			}
		case <-timer.C:
			timer.Reset(maxWait)
			if callCnt > 0 {
				l := len(callRows)
				m.bulkInsert(callQuery, sipQueryVal(l/sipValCnt), callRows[:l])
				callRows = []interface{}{}
				callCnt = 0
			}
			if regCnt > 0 {
				l := len(regRows)
				m.bulkInsert(registerQuery, sipQueryVal(l/sipValCnt), regRows[:l])
				regRows = []interface{}{}
				regCnt = 0
			}
			if restCnt > 0 {
				l := len(restRows)
				m.bulkInsert(restQuery, sipQueryVal(l/sipValCnt), restRows[:l])
				restRows = []interface{}{}
				restCnt = 0
			}
			if rtcpCnt > 0 {
				l := len(rtcpRows)
				m.bulkInsert(rtcpQuery, rtcQueryVal(l/rtcValCnt), rtcpRows[:l])
				rtcpRows = []interface{}{}
				rtcpCnt = 0
			}
			if reportCnt > 0 {
				l := len(reportRows)
				m.bulkInsert(reportQuery, rtcQueryVal(l/rtcValCnt), reportRows[:l])
				reportRows = []interface{}{}
				reportCnt = 0
			}
			if dnsCnt > 0 {
				l := len(dnsRows)
				m.bulkInsert(dnsQuery, rtcQueryVal(l/rtcValCnt), dnsRows[:l])
				dnsRows = []interface{}{}
				dnsCnt = 0
			}
			if logCnt > 0 {
				l := len(logRows)
				m.bulkInsert(logQuery, rtcQueryVal(l/rtcValCnt), logRows[:l])
				logRows = []interface{}{}
				logCnt = 0
			}
		}
	}
}

func (m *MySQL) bulkInsert(q, v []byte, rows []interface{}) {
	tblDate := time.Now().In(time.UTC).AppendFormat(q, "20060102")
	query := make([]byte, len(tblDate)+len(v))
	tdl := copy(query, tblDate)
	copy(query[tdl:], v)
	_, err := m.db.Exec(string(query), rows...)
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
	out := make([]byte, c*len(sipPlaceholder)+len(sipVal)-1)
	bs := copy(out, sipVal)
	for i := 0; i < c; i++ {
		bp := i * len(sipPlaceholder)
		copy(out[bs+bp:], sipPlaceholder)
	}
	return out
}

func rtcQueryVal(c int) []byte {
	out := make([]byte, c*len(rtcPlaceholder)+len(rtcVal)-1)
	bs := copy(out, rtcVal)
	for i := 0; i < c; i++ {
		bp := i * len(rtcPlaceholder)
		copy(out[bs+bp:], rtcPlaceholder)
	}
	return out
}
