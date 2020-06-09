// Copyright 2011, Shelby Ramsey. All rights reserved.
// Copyright 2018, Eugen Biegler. All rights reserved.
// Use of this code is governed by a BSD license that can be
// found in the LICENSE.txt file.

package sipparser

// Imports from the go standard library
import (
	//"fmt"
	//"fmt"
	"testing"

	"github.com/sipcapture/heplify-server/config"
)

var testInviteMsg = "INVITE sip:15554440000@X.X.X.X:5060;user=phone SIP/2.0\r\nVia: SIP/2.0/UDP X.X.X.X:5060;branch=z9hG4bK34133a599ll241207INV21d7d0684e84a2d2\r\nMax-Forwards: 35\r\nContact: <sip:X.X.X.X:5060>\r\nTo: <sip:15554440000@X.X.X.X;user=phone;noa=national>\r\nFrom: \"Unavailable\"<sip:X.X.X.X;user=phone;noa=national>;tag=21d7d068-co2149-FOOI003\r\nCall-ID: 1393184968_47390262@domain.com\r\nCSeq: 214901 INVITE\r\nAuthorization: Digest username=\"foobaruser124\", realm=\"FOOBAR\", algorithm=MD5, uri=\"sip:foo.bar.com\", nonce=\"4f6d7a1d\", response=\"6a79a5c75572b0f6a18963ae04e971cf\", opaque=\"\"\r\nAllow: INVITE,ACK,CANCEL,BYE,REFER,OPTIONS,NOTIFY,SUBSCRIBE,PRACK,INFO\r\nContent-Type: application/sdp\r\nDate: Thu, 29 Sep 2011 16:54:42 GMT\r\nUser-Agent: FAKE-UA-DATA\r\nP-Asserted-Identity: \"Unavailable\"<sip:Restricted@X.X.X.X:5060>\r\nContent-Length: 322\r\n\r\nv=0\r\no=- 567791720 567791720 IN IP4 X.X.X.X\r\ns=FAKE-DATA\r\nc=IN IP4 X.X.X.X\r\nt=0 0\r\nm=audio 17354 RTP/AVP 0 8 86 18 96\r\na=rtpmap:0 PCMU/8000\r\na=rtpmap:8 PCMA/8000\r\na=rtpmap:86 G726-32/8000\r\na=rtpmap:18 G729/8000\r\na=rtpmap:96 telephone-event/8000\r\na=maxptime:20\r\na=fmtp:18 annexb=yes\r\na=fmtp:96 0-15\r\na=sendrecv\r\n"
var test400ErrorResp = "SIP/2.0 400 Bad Request\r\n"

// should succeed, i.e. not produce an error. 
func TestParse400Err(t *testing.T) {
	s := ParseMsg(test400ErrorResp, nil, nil)
	if s.Error != nil {
		t.Errorf("[TestParse400Err] Error parsing msg. Received: %v", s.Error)
	}
}

func TestHeader(t *testing.T) {
	h := Header{"t", "v"}
	if h.String() != "t: v" {
		t.Errorf("[TestHeader] Error with header.String() method. Unexpected res.")
	}
}

// makes sure that the body is what is expected
func TestBody(t *testing.T) {
	s := ParseMsg("fake\r\nheader\r\n\r\nbody ...\r\n\r\n", nil, nil)
	if s.Body != "body ...\r\n\r\n" {
		t.Errorf("[TestBody] Error getting the right body from the string.")
	}
}

// actual msg testing
func TestParseMsg(t *testing.T) {
	x := []string{"X-MEN", "X-FORCE"}
	y := []string{"X-CUSTOM"}
	m := "SIP/2.0 200 OK\r\nVia: SIP/2.0/UDP 0.0.0.0:5060;branch=z9hG4bK24477ab511325213INV52e94be64e6687e3;received=0.0.0.0\r\nContact: <sip:10003053258853@0.0.0.0:6060>\r\nTo: <sip:10003053258853@0.0.0.0;user=phone;noa=national>;tag=a94c095b773be1dd6e8d668a785a9c843f6f2cc0\r\nFrom: <sip:8173383772@0.0.0.0;user=phone;noa=national>;tag=52e94be6-co2998-INS002\nCall-ID: 111118149-3524331107-398662@barinfo.fooinfous.com\r\nCSeq: 299801 INVITE\r\nDiversion: something\r\nAccept: application/sdp, application/dtmf-relay, text/plain\r\nUser-Agent: FAKE-UA-DATA\r\nRemote-Party-Id: something\r\nServer: something\r\nX-Nonsense-Hdr: nonsense\r\n" +
		"Allow: PRACK, INVITE, BYE, REGISTER, ACK, OPTIONS, CANCEL, SUBSCRIBE, NOTIFY, INFO, REFER, UPDATE\r\nContent-Type: application/sdp\r\nServer: Dialogic-SIP/10.5.3.231 IMGDAL0001 0\r\nSupported: 100rel, path, replaces, timer, tdialog\r\nContent-Length: 239\r\nX-FORCE: Deadpool\r\nX-CUSTOM: Avenger\r\nP-Asserted-Identity: <sip:8884441111@1.1.1.1:5060;user=phone>\r\nP-Asserted-Identity: <sip:8884442222>\r\n" +
		"Contact: something\r\nAuthorization: Digest username=\"foobaruser124\", realm=\"FOOBAR\"\r\nProxy-Authorization: Digest username=\"foobaruser124\", realm=\"FOOBAR\"\r\n" +
		"X-RTP-Stat: CS=0;PS=1433;ES=1525;OS=229280;SP=0/0;SO=0;QS=-;PR=1522;ER=1525;OR=243520;CR=0;SR=0;QR=-;PL=0,0;BL=0;LS=0;RB=0/0;SB=-/-;EN=PCMA,FAX;DE=PCMA;JI=23,2;DL=20,20,21;IP=83.138.49.179:7082,102.183.157.163:25132\r\nX-RTP-Stat-Add: DQ=31;DSS=0;DS=0;PLCS=288;JS=1\r\n\r\nv=0\r\no=Dialogic_SDP 1452654 0 IN IP4 0.0.0.0\r\ns=Dialogic-SIP\r\nc=IN IP4 4.71.122.135\r\nt=0 0\r\nm=audio 11676 RTP/AVP 0 101\r\na=rtpmap:0 PCMU/8000\r\na=rtpmap:101 telephone-event/8000\r\na=fmtp:101 0-15\r\na=silenceSupp:off - - - -\r\na=ptime:20\r\n"
	if got, want := GetSIPHeaderVal("Call-ID:", m), "111118149-3524331107-398662@barinfo.fooinfous.com"; got != want {
		t.Errorf("[TestGetSIPHeaderVal] Error getting Call-ID. Call-ID: should be '%s'. Received: '%s'", want, got)
	}

	s := ParseMsg(m, x, y)
	//fmt.Println(s.StartLine.Val)
	if s.Error != nil {
		t.Errorf("[TestParseMsg] Error parsing msg. Recevied: %v", s.Error)
	}
	if len(s.Body) == 0 {
		t.Error("[TestParseMsg] Error parsing msg. Body should have a length.")
	}

	/* 	if len(s.Headers) == 0 {
		t.Error("[TestParseMsg] Error parsing msg.  Does not appear to be any headers.")
	} */
	/* 	if s.Via == nil || len(s.Via) == 0 {
		t.Error("[TestParseMsg] Error parsing msg.  Does not appear to be any vias parsed.")
		//fmt.Println("msg:", s.Msg)
		//fmt.Println("body:", s.Body)
		//fmt.Println("via:", s.Via)
		//fmt.Println("crlf:", s.crlf)
	} */
	if got, want := s.ContentType, "application/sdp"; got != want {
		t.Errorf("[TestParseMsg] Error parsing msg. Content type should be %s. Received: %s", want, got)
	}
	if s.ContentLength != "239" {
		t.Errorf("[TestParseMsg] Error parsing msg. Content length should be 239. Received: %s", s.ContentLength)
	}
	if s.CallID != "111118149-3524331107-398662@barinfo.fooinfous.com" {
		t.Errorf("[TestParseMsg] Error parsing msg. Call-ID should be 111118149-3524331107-398662@barinfo.fooinfous.com. Received: %s", s.CallID)
	}
	if s.FromTag != "52e94be6-co2998-INS002" {
		t.Errorf("[TestParseMsg] Error parsing msg. From Tag should be 52e94be6-co2998-INS002. Received: %s", s.FromTag)
	}
	if s.ToTag != "a94c095b773be1dd6e8d668a785a9c843f6f2cc0" {
		t.Errorf("[TestParseMsg] Error parsing msg. To Tag should be a94c095b773be1dd6e8d668a785a9c843f6f2cc0. Received: %s", s.ToTag)
	}
	if s.UserAgent != "FAKE-UA-DATA" {
		t.Errorf("[TestParseMsg] Error parsing msg. UserAgent should be FAKE-UA-DATA. Received: %s", s.UserAgent)
	}
	if s.ViaOneBranch != "z9hG4bK24477ab511325213INV52e94be64e6687e3" {
		t.Errorf("[TestParseMsg] Error parsing msg. ViaOneBranch should be z9hG4bK24477ab511325213INV52e94be64e6687e3. Received: %s", s.ViaOneBranch)
	}
	if s.CseqMethod != "INVITE" {
		t.Errorf("[TestParseMsg] Error parsing msg. CseqMethod should be INVITE. Received: %s", s.CseqMethod)
	}

	/* 	if s.ContentLengthInt != 239 {
		t.Errorf("[TestParseMsg] Error parsing msg.  ContentLengthInt should be 239.  Received: %d", s.ContentLengthInt)
	} */
	/* 	if len(s.Supported) != 5 {
		t.Error("[TestParseMsg] Error parsing msg.  s.Support should have length of 5.")
	} */
	if got, want := s.PaiUser, "8884441111"; got != want {
		t.Errorf("[TestParseMsg] Error parsing msg. PaiUser should be %s. Received: %s", want, got)
	}
	if got, want := s.PaiHost, "1.1.1.1"; got != want {
		t.Errorf("[TestParseMsg] Error parsing msg. PaiHost should be %s. Received: %s", want, got)
	}
	if s.RTPStatVal == "" {
		t.Error("[TestParseMsg] Error parsing msg. RTPStatVal shouldn't be empty.")
	}
	if s.DiversionVal == "" {
		t.Error("[TestParseMsg] Error parsing msg. DiversionVal shouldn't be empty.")
	}
	if s.RemotePartyIdVal == "" {
		t.Error("[TestParseMsg] Error parsing msg. RemotePartyIdVal shouldn't be empty.")
	}
	if s.Server == "" {
		t.Error("[TestParseMsg] Error parsing msg. Server shouldn't be empty.")
	}
	if s.AuthVal == "" {
		t.Error("[TestParseMsg] Error parsing msg. AuthVal shouldn't be empty.")
	}
	if s.ContactVal == "" {
		t.Error("[TestParseMsg] Error parsing msg. ContactVal shouldn't be empty.")
	}

	if got, want := s.XCallID, "Deadpool"; got != want {
		t.Errorf("[TestParseMsg] Error parsing msg. XCallID should be %s. Received: %s", want, got)
	}
	if got, want := s.CustomHeader["X-CUSTOM"], "Avenger"; got != want {
		t.Errorf("[TestParseMsg] Error parsing msg. X-CUSTOM should be %s. Received: %s", want, got)
	}

	x = []string{"X-MEN", "P-Charging-Vector,icid-value=\"(.*)\""}
	m = "SIP/2.0 200 OK\r\nVia: SIP/2.0/UDP 0.0.0.0:5060;branch=z9hG4bK24477ab511325213INV52e94be64e6687e3;received=0.0.0.0\r\nContact: <sip:10003053258853@0.0.0.0:6060>\r\nTo: <sip:10003053258853@0.0.0.0;user=phone;noa=national>;tag=a94c095b773be1dd6e8d668a785a9c843f6f2cc0\r\nFrom: <sip:8173383772@0.0.0.0;user=phone;noa=national>;tag=52e94be6-co2998-INS002\nCall-ID: 111118149-3524331107-398662@barinfo.fooinfous.com\r\nCSeq: 299801 INVITE\r\nDiversion: something\r\nAccept: application/sdp, application/dtmf-relay, text/plain\r\nUser-Agent: FAKE-UA-DATA\r\nRemote-Party-Id: something\r\nServer: something\r\nX-Nonsense-Hdr: nonsense\r\n" +
		"Allow: PRACK, INVITE, BYE, REGISTER, ACK, OPTIONS, CANCEL, SUBSCRIBE, NOTIFY, INFO, REFER, UPDATE\r\nContent-Type: application/sdp\r\nServer: Dialogic-SIP/10.5.3.231 IMGDAL0001 0\r\nSupported: 100rel, path, replaces, timer, tdialog\r\nContent-Length: 239\r\nX-FORCE: Deadpool\r\nX-CUSTOM: Avenger\r\nP-Charging-Vector: icid-value=\"somevalue\";orig-ioi=cat.tree.root\r\nP-Asserted-Identity: <sip:8884441111@1.1.1.1:5060;user=phone>\r\nP-Asserted-Identity: <sip:8884442222>\r\n" +
		"Contact: something\r\nAuthorization: Digest username=\"foobaruser124\", realm=\"FOOBAR\"\r\nProxy-Authorization: Digest username=\"foobaruser124\", realm=\"FOOBAR\"\r\n" +
		"X-RTP-Stat: CS=0;PS=1433;ES=1525;OS=229280;SP=0/0;SO=0;QS=-;PR=1522;ER=1525;OR=243520;CR=0;SR=0;QR=-;PL=0,0;BL=0;LS=0;RB=0/0;SB=-/-;EN=PCMA,FAX;DE=PCMA;JI=23,2;DL=20,20,21;IP=83.138.49.179:7082,102.183.157.163:25132\r\nX-RTP-Stat-Add: DQ=31;DSS=0;DS=0;PLCS=288;JS=1\r\n\r\nv=0\r\no=Dialogic_SDP 1452654 0 IN IP4 0.0.0.0\r\ns=Dialogic-SIP\r\nc=IN IP4 4.71.122.135\r\nt=0 0\r\nm=audio 11676 RTP/AVP 0 101\r\na=rtpmap:0 PCMU/8000\r\na=rtpmap:101 telephone-event/8000\r\na=fmtp:101 0-15\r\na=silenceSupp:off - - - -\r\na=ptime:20\r\n"

	x = config.GenerateRegexMap(x)

	s = ParseMsg(m, x, y)
	//fmt.Println(s.StartLine.Val)
	if s.Error != nil {
		t.Errorf("[TestParseMsg] Error parsing msg. Recevied: %v", s.Error)
	}
	if len(s.Body) == 0 {
		t.Error("[TestParseMsg] Error parsing msg. Body should have a length.")
	}
	if got, want := s.XCallID, "somevalue"; got != want {
		t.Errorf("[TestParseMsg] Error parsing msg. XCallID should be %s. Received: %s", want, got)
	}
}

func TestParseInviteMsg(t *testing.T) {
	s := ParseMsg(testInviteMsg, nil, nil)
	if s.Error != nil {
		t.Errorf("[TestParseInviteMsg] Error parsing msg. Recevied: %v", s.Error)
	}
	if len(s.Body) == 0 {
		t.Error("[TestParseInviteMsg] Error parsing msg. Body should have a length.")
	}
}

func TestParseMsgMalformed(t *testing.T) {
	m := "SIP/2.0 200 OK\r\nVia:\r\nTo:\rnContact\r\n   Frommmm     :asf\r\nCall-ID:111118149-3524331107-398662@barinfo.fooinfous.com\r\n\r\nv=0\r\no=Dialogic_SDP 1452654 0 IN IP4 0.0.0.0\r\ns=Dialogic-SIP\r\nc=IN IP4 4.71.122.135\r\nt=0 0\r\nm=audio 11676 RTP/AVP 0 101\r\na=rtpmap:0 PCMU/8000\r\na=rtpmap:101 telephone-event/8000\r\na=fmtp:101 0-15\r\na=silenceSupp:off - - - -\r\na=ptime:20\r\n"
	s := ParseMsg(m, nil, nil)
	if s.Error != nil {
		t.Errorf("[TestParseMsgMalformed] Error parsing msg. Recevied: %v", s.Error)
	}
	if len(s.Body) == 0 {
		t.Error("[TestParseMsgMalformed] Error parsing msg. Body should have a length.")
	}
	if s.CallID != "111118149-3524331107-398662@barinfo.fooinfous.com" {
		t.Errorf("[TestParseMsgMalformed] Error parsing msg. Call-ID should be 111118149-3524331107-398662@barinfo.fooinfous.com. Received: %s", s.CallID)
	}
}

func TestParseMsgCompact(t *testing.T) {
	m := "OPTIONS sip:12329723@sip.test.de SIP/2.0\r\nv:SIP/2.0/UDP 11.23.142.14:5086;rport;branch=z9hG4bKDrrK0aZayHU5m\r\nf:<sip:1288837@sip.test.de>;tag=SBKrNFysdfp5t9BN\r\nt:<sip:422788837@sip.test.de>\r\ni:OvwBVivTMK19kN3Ws51_Dv\r\nCSeq:688171 OPTIONS\r\nAccept:application/vnd.nokia-register-usage\r\ns:REGISTRATION PROBE\r\nl:1\r\n\r\n"
	s := ParseMsg(m, nil, nil)

	if s.Error != nil {
		t.Errorf("[TestParseMsgCompact] Error parsing msg. Recevied: %v", s.Error)
	}
	if got, want := s.CallID, "OvwBVivTMK19kN3Ws51_Dv"; got != want {
		t.Errorf("[TestParseMsgCompact] Error parsing msg. Call-ID should be %s. Received: %s", want, got)
	}

	/* 	if got, want := s.Accept.Val, "application/vnd.nokia-register-usage"; got != want {
		t.Errorf("[TestParseMsgCompact] Error parsing msg.  Content type should be %s.  Received: %s", got, want)
	} */
	if s.ContentLength != "1" {
		t.Errorf("[TestParseMsgCompact] Error parsing msg. Content length should be '1'. Received: %s", s.ContentLength)
	}
	/* 	if s.ContentLengthInt != 1 {
		t.Errorf("[TestParseMsgCompact] Error parsing msg. ContentLengthInt should be 1. Received: %d", s.ContentLengthInt)
	} */
}

// testing the GetCallingParty functionality
func TestGetCallingParty(t *testing.T) {
	rpid := "\"UNKNOWN\" <sip:8885551000@0.0.0.0>;party=calling;screen=yes;privacy=off"
	s := &SipMsg{RemotePartyIdVal: rpid}
	err := s.GetCallingParty(CALLING_PARTY_RPID)
	if err != nil {
		t.Errorf("[TestGetCallingParty] Err with GetCallingParty.  rcvd: %v", err)
	}
	if s.CallingParty == nil {
		t.Error("[TestGetCallingParty] Err calling GetCallingParty.  CallingParty field should not be nil.")
	}
	if s.CallingParty.Name != "UNKNOWN" {
		t.Error("[TestGetCallingParty] Err calling GetCallingParty.  Name should be \"UNKNOWN\".")
	}
	if s.CallingParty.Number != "8885551000" {
		t.Error("[TestGetCallingParty] Err with GetCallingParty. Number should be \"8885551000\".")
	}
	paid := "<sip:8884441000@0.0.0.0:5060;user=phone>"
	s = &SipMsg{PAssertedIdVal: paid}
	err = s.GetCallingParty(CALLING_PARTY_PAID)
	if err != nil {
		t.Errorf("[TestGetCallingParty] Err with GetCallingParty on paid.  rcvd: %v", err)
	}
	if s.CallingParty == nil {
		t.Error("[TestGetCallingParty] Err with GetCallingParty on paid.  No CallingPartyInfo.")
	}
	if s.CallingParty.Name != "" {
		t.Error("[TestGetCallingParty] Err with GetCallingParty on paid.  Name should be \"\".")
	}
	if s.CallingParty.Number != "8884441000" {
		t.Error("[TestGetCallingParty] Err with GetCallingParty on paid.  Number should be \"8884441000\".")
	}
	s = &SipMsg{}
	s.parseFrom("\"5556661000\" <sip:5556661000@0.0.0.0>;tag=ZN21rHN5B7U0K")
	err = s.GetCallingParty("")
	if err != nil {
		t.Errorf("[TestGetCallingParty] Err calling GetCallingParty on default.  rcvd: %v", err)
	}
	if s.CallingParty == nil {
		t.Error("[TestGetCallingParty] Err calling GetCallingParty on default.  No CallingPartyInfo.")
	}
	if s.CallingParty.Name != "5556661000" {
		t.Error("[TestGetCallingParty] Err calling GetCallingParty on default.  Name should be \"5556661000\".")
	}
	if s.CallingParty.Number != "5556661000" {
		t.Error("[TestGetCallingParty] Err calling GetCallingParty on default. Number should be \"5556661000\".")
	}
}

func BenchmarkParseMsg(b *testing.B) {
	for i := 0; i < b.N; i++ {
		ParseMsg(testInviteMsg, nil, nil)
	}
}

func BenchmarkParseMsgParallel(b *testing.B) {
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			ParseMsg(testInviteMsg, nil, nil)
		}
	})
}
