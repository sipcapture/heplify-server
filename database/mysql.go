package database

import (
	"net/url"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/negbie/heplify-server"
	"github.com/negbie/heplify-server/config"
	"github.com/negbie/heplify-server/logp"
	fastsql "github.com/rmulley/go-fast-sql"
)

var (
	method      string
	respText    string
	uriRaw      string
	uriUser     string
	uriHost     string
	fromUser    string
	fromHost    string
	fromTag     string
	toUser      string
	toHost      string
	toTag       string
	contactUser string
	contactHost string
	contactPort string
)

var (
	insertStatement = `INSERT INTO %s_%s (
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
)

type MySQL struct {
	dbh *fastsql.DB
}

func (m *MySQL) setup() error {
	var err error
	if m.dbh, err = fastsql.Open("mysql", config.Setting.DBUser+":"+config.Setting.DBPassword+"@tcp("+config.Setting.DBAddr+")/"+config.Setting.DBName+"?"+url.QueryEscape("parseTime=true"), 1000); err != nil {
		logp.Err("%v", err)
		return err
	}
	defer m.dbh.Close()
	return nil
}

func (m *MySQL) insert(topic string, mCh chan *decoder.HEPPacket, ec *uint64) {
	var (
		pkt *decoder.HEPPacket
		ok  bool
	)

	logp.Info("Run MySQL Output, server: %+v, topic: %s\n", config.Setting.DBAddr, topic)

	for {
		pkt, ok = <-mCh
		if !ok {
			break
		}

		m.validate(pkt)

		err := m.dbh.BatchInsert(insertStatement,
			time.Now().UTC(),
			time.Now().UnixNano()/(int64(time.Microsecond)/int64(time.Nanosecond)),
			method,
			respText,
			uriRaw,
			uriUser,
			uriHost,
			fromUser,
			fromHost,
			fromTag,
			toUser,
			toHost,
			toTag,
			contactHost,
			contactPort,
			contactUser,
			pkt.SipMsg.CallId,
			"", // TODO CallId-Aleg,
			pkt.SipMsg.Via[0],
			pkt.SipMsg.Via[0].Branch,
			pkt.SipMsg.Cseq.Val,
			"", // TODO diversion,
			"", // TODO reason,
			pkt.SipMsg.ContentType,
			"", // TODO auth,
			pkt.SipMsg.UserAgent,
			pkt.SrcIP.String(),
			pkt.SrcPort,
			pkt.DstIP.String(),
			pkt.DstPort,
			"", // TODO contact_ip,
			"", // TODO contact_port,
			"", // TODO originator_ip,
			"", // TODO originator_port,
			pkt.Protocol,
			pkt.Version,
			"", // TODO rtp_stat,
			pkt.ProtoType,
			pkt.NodeID,
			pkt.SipMsg.CallId,
			pkt.SipMsg.Msg,
		)
		if err != nil {
			logp.Err("%v", err)
			*ec++
		}
	}
}

func (m *MySQL) validate(pkt *decoder.HEPPacket) {
	if pkt.SipMsg.StartLine.URI != nil {
		method = pkt.SipMsg.StartLine.Method
		respText = pkt.SipMsg.StartLine.RespText
		uriRaw = pkt.SipMsg.StartLine.URI.Raw
		uriUser = pkt.SipMsg.StartLine.URI.User
		uriHost = pkt.SipMsg.StartLine.URI.Host
	} else {
		method = ""
		respText = ""
		uriRaw = ""
		uriUser = ""
		uriHost = ""
	}
	if pkt.SipMsg.From != nil {
		fromUser = pkt.SipMsg.From.URI.User
		fromHost = pkt.SipMsg.From.URI.Host
		fromTag = pkt.SipMsg.From.Tag
	} else {
		fromUser = ""
		fromHost = ""
		fromTag = ""
	}
	if pkt.SipMsg.To != nil {
		toUser = pkt.SipMsg.To.URI.User
		toHost = pkt.SipMsg.To.URI.Host
		toTag = pkt.SipMsg.To.Tag
	} else {
		toUser = ""
		toHost = ""
		toTag = ""
	}
	if pkt.SipMsg.Contact != nil {
		contactUser = pkt.SipMsg.Contact.URI.User
		contactHost = pkt.SipMsg.Contact.URI.Host
		contactPort = pkt.SipMsg.Contact.URI.Port
	} else {
		contactUser = ""
		contactHost = ""
		contactPort = ""
	}
}
