package database

import (
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
	insTable    string
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
			) VALUES( ?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?);`

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
			) VALUES( ?,?,?,?,?,?,?,?,?,?,?,?);`
)

type SQL struct {
	dbo *DB
}

func (s *SQL) setup() error {
	var err error
	if config.Setting.DBDriver == "mysql" {
		if s.dbo, err = Open(config.Setting.DBDriver, config.Setting.DBUser+":"+config.Setting.DBPassword+"@tcp("+config.Setting.DBAddr+":3306)/"+config.Setting.DBName+"?"+url.QueryEscape("charset=utf8mb4,utf8&parseTime=true"), 1000); err != nil {
			logp.Err("%v", err)
			return err
		}
	} else {
		if s.dbo, err = Open(config.Setting.DBDriver, "host="+config.Setting.DBAddr+"port=5432"+"dbname="+config.Setting.DBName+"user="+config.Setting.DBUser+"password="+config.Setting.DBPassword, 100); err != nil {
			logp.Err("%v", err)
			return err
		}
	}

	if err = s.dbo.Ping(); err != nil {
		s.dbo.Close()
		logp.Err("%v", err)
		return err
	}

	return nil
}

func (s *SQL) insert(topic string, mCh chan *decoder.HEPPacket, ec *uint64) {
	var (
		pkt *decoder.HEPPacket
		ok  bool
		err error
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

		if pkt.ProtoType == 1 && pkt.SipMsg != nil {

			if pkt.SipMsg.Cseq.Method == "REGISTER" {
				insTable = "INSERT INTO sip_capture_registration_"
			} else {
				insTable = "INSERT INTO sip_capture_call_"
			}

			err = s.dbo.BatchInsert(insTable+time.Now().Format("20060102")+callRegStmt,
				pkt.Timestamp.Format("2006-01-02 15:04:05"), // match stupid TIMESTAMP format
				pkt.Timestamp.UnixNano()/(int64(time.Microsecond)/int64(time.Nanosecond)),
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
				pkt.SipMsg.Msg,
			)
		} else if pkt.ProtoType >= 2 && pkt.ProtoType <= 200 && pkt.CorrelationID != nil {
			switch pkt.ProtoType {
			case 5:
				insTable = "INSERT INTO rtcp_capture_all_"
			case 34:
				insTable = "INSERT INTO report_capture_all_"
			case 38:
				insTable = "INSERT INTO report_capture_all_"
			case 53:
				insTable = "INSERT INTO dns_capture_all_"
			case 100:
				insTable = "INSERT INTO logs_capture_all_"
			default:
				insTable = "INSERT INTO rtcp_capture_all_"
			}

			err = s.dbo.BatchInsert(insTable+time.Now().Format("20060102")+jsonStmt,
				pkt.Timestamp.Format("2006-01-02 15:04:05"), // match stupid TIMESTAMP format
				pkt.Timestamp.UnixNano()/(int64(time.Microsecond)/int64(time.Nanosecond)),
				string(pkt.CorrelationID),
				pkt.SrcIP.String(),
				pkt.SrcPort,
				pkt.DstIP.String(),
				pkt.DstPort,
				pkt.Protocol,
				pkt.Version,
				pkt.ProtoType,
				pkt.NodeID,
				string(pkt.Payload),
			)
		}
		if err != nil {
			logp.Err("%v", err)
			*ec++
		}
	}
}
