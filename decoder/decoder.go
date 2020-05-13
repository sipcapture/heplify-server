package decoder

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	reflect "reflect"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"
	"unsafe"

	"github.com/VictoriaMetrics/fastcache"
	xxhash "github.com/cespare/xxhash/v2"
	"github.com/negbie/logp"
	"github.com/sipcapture/heplify-server/config"
	"github.com/sipcapture/heplify-server/sipparser"
)

// The first 4 bytes are the string "HEP3". The next 2 bytes are the length of the
// whole message ( len("HEP3") ) + length of all the chunks we have. The next bytes
// are all the chunks created by makeChunks()
// Bytes: 0 1 2 3 4 5 6 7 8 9 10 11 12 13 14 15 16 17 18 19 20 21 22 23 24 25 26 27 28 29 30 31......
//        +-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
//        | "HEP3"|len|chunks(0x0001|0x0002|0x0003|0x0004|0x0007|0x0008|0x0009|0x000a|0x000b|......)
//        +-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+

var (
	dedup                 = fastcache.New(32 * 1024 * 1024)
	strBackslashQuote     = []byte(`\"`)
	strBackslashBackslash = []byte(`\\`)
	strBackslashN         = []byte(`\n`)
	strBackslashR         = []byte(`\r`)
	strBackslashT         = []byte(`\t`)
	strBackslashF         = []byte(`\u000c`)
	strBackslashB         = []byte(`\u0008`)
	strBackslashLT        = []byte(`\u003c`)
	strBackslashQ         = []byte(`\u0027`)
	strEmpty              = []byte(``)
)

// HEP chunks
const (
	IPVersion 	= 1  // Chunk 0x0001 IP protocol family (0x02=IPv4, 0x0a=IPv6)
	TransportProto	= 2  // Chunk 0x0002 IP protocol ID (0x06=TCP, 0x11=UDP)
	IP4SourceIP	= 3  // Chunk 0x0003 IPv4 source address
	IP4DestIP	= 4  // Chunk 0x0004 IPv4 destination address
	IP6SourceIP	= 5  // Chunk 0x0005 IPv6 source address
	IP6DestIP	= 6  // Chunk 0x0006 IPv6 destination address
	SourcePort	= 7  // Chunk 0x0007 Protocol source port
	DestPort	= 8  // Chunk 0x0008 Protocol destination port
	TimeSec 	= 9  // Chunk 0x0009 Unix timestamp, seconds
	TimeMicrosec	= 10 // Chunk 0x000a Unix timestamp, microseconds
	AppProto	= 11 // Chunk 0x000b Ingested Protocol type (DNS, LOG, RTCP, SIP)
	NodeID    	= 12 // Chunk 0x000c Capture client ID
	NodePW    	= 14 // Chunk 0x000e Authentication key (plain text / TLS connection)
	Payload   	= 15 // Chunk 0x000f Captured packet payload
	CorrelationID	= 17 // Chunk 0x0011 Correlation ID
	Vlan     	= 18 // Chunk 0x0012 VLAN
	NodeName  	= 19 // Chunk 0x0013 NodeName
)

// HEP represents HEP packet
type HEP struct {
	IPVersion 	uint32 `protobuf:"varint,1,req,name=IPVersion" json:"Version"`
	TransportProto 	uint32 `protobuf:"varint,2,req,name=TransportProto" json:"Protocol"`
	SourceIP 	string `protobuf:"bytes,3,req,name=SourceIP" json:"SrcIP"`
	DestIP 		string `protobuf:"bytes,4,req,name=DestIP" json:"DstIP"`
	SourcePort 	uint32 `protobuf:"varint,5,req,name=SourcePort" json:"SrcPort"`
	DestPort 	uint32 `protobuf:"varint,6,req,name=DestPort" json:"DstPort"`
	TimeSec 	uint32 `protobuf:"varint,7,req,name=TimeSec" json:"Tsec"`
	TimeMicrosec 	uint32 `protobuf:"varint,8,req,name=TimeMicrosec" json:"Tmsec"`
	AppProto 	uint32 `protobuf:"varint,9,req,name=AppProto" json:"ProtoType"`
	NodeID 		uint32 `protobuf:"varint,10,req,name=NodeID" json:"NodeID"`
	NodePW 		string `protobuf:"bytes,11,req,name=NodePW" json:"NodePW"`
	Payload 	string `protobuf:"bytes,12,req,name=Payload" json:"Payload"`
	CorrelationID 	string `protobuf:"bytes,13,req,name=CorrelationID" json:"CID"`
	Vlan 		uint32 `protobuf:"varint,14,req,name=Vlan" json:"Vlan"`
	// The following variables hold values derived from the ingress data, i.e. not found at ingress.
	ProtoAsString 	string
	AssembledTimestamp	time.Time
	SIP         *sipparser.SipMsg
	NodeName    	string
	SID         string
}

// DecodeHEP returns a parsed HEP message
func DecodeHEP(packet []byte) (*HEP, error) {
	hep := &HEP{}
	err := hep.parse(packet)
	if err != nil {
		return nil, err
	}
	return hep, nil
}

func (hep *HEP) parse(packet []byte) error {
	var err error
	if bytes.HasPrefix(packet, []byte{0x48, 0x45, 0x50, 0x33}) {
		err = hep.parseHEP(packet)
		if err != nil {
			logp.Warn("%v", err)
			return err
		}
	} else {
		err = hep.Unmarshal(packet)
		if err != nil {
			logp.Warn("malformed packet with length %d which is neither HEP nor protobuf encapsulated", len(packet))
			return err
		}
	}

	t := time.Now()
	hep.AssembledTimestamp = time.Unix(int64(hep.TimeSec), int64(hep.TimeMicrosec*1000))
	d := t.Sub(hep.AssembledTimestamp)
	if d < 0 || (hep.TimeSec == 0 && hep.TimeMicrosec == 0) {
		logp.Debug("hep", "got timestamp in the future with delta: %d from NodeID %d", d, hep.NodeID)
		hep.AssembledTimestamp = t
	}

	hep.normPayload()
	if hep.AppProto == 0 {
		return nil
	}

	if hep.AppProto == 1 && len(hep.Payload) > 32 {
		err = hep.parseSIP()
		if err != nil {
			logp.Warn("%v\n%q\nNodeID: %d, AppProto: %d, IPVersion: %d, CapturedTransport: %d, length: %d, flow: %s:%d->%s:%d\n\n",
				err, hep.Payload, 
				hep.NodeID, hep.AppProto, hep.IPVersion, hep.TransportProto, len(hep.Payload), hep.SourceIP, hep.SourcePort, hep.DestIP, hep.DestPort)
			return err
		}

		if len(config.Setting.DiscardMethod) > 0 {
			for k := range config.Setting.DiscardMethod {
				if config.Setting.DiscardMethod[k] == hep.SIP.CseqMethod {
					hep.AppProto = 0
					return nil
				}
			}
		}
	}

	if hep.NodeName == "" {
		hep.NodeName = strconv.FormatUint(uint64(hep.NodeID), 10)
	}

	logp.Debug("HEP", "%+v\n\n", hep)
	return nil
}

var fixUTF8 = func(r rune) rune {
	if r == utf8.RuneError || r == '\x00' {
		return -1
	}
	return r
}

func (hep *HEP) normPayload() {
	if config.Setting.Dedup {
		ts := uint64(hep.AssembledTimestamp.UnixNano())
		kh := make([]byte, 8)
		ks := xxhash.Sum64String(hep.Payload)
		binary.BigEndian.PutUint64(kh, ks)

		if buf := dedup.Get(nil, kh); buf != nil {
			i := binary.BigEndian.Uint64(buf)
			d := ts - i
			if i > ts {
				d = i - ts
			}
			if d < 500e6 {
				hep.AppProto = 0
				return
			}
		}

		tb := make([]byte, 8)
		binary.BigEndian.PutUint64(tb, ts)
		dedup.Set(kh, tb)
	}
	if !utf8.ValidString(hep.Payload) {
		hep.Payload = strings.Map(fixUTF8, hep.Payload)
	}
	if config.Setting.DBDriver == "postgres" && strings.Index(hep.Payload, "\x00") > -1 {
		hep.Payload = strings.Map(fixUTF8, hep.Payload)
	}
}

func (hep *HEP) EscapeFields(w io.Writer, tag string) (int, error) {
	switch tag {
	case "callid":
		return WriteJSONString(w, hep.SIP.CallID)
	case "cseq":
		return WriteJSONString(w, hep.SIP.CseqVal)
	case "method":
		return WriteJSONString(w, hep.SIP.FirstMethod)
	case "ruri_user":
		return WriteJSONString(w, hep.SIP.URIUser)
	case "ruri_domain":
		return WriteJSONString(w, hep.SIP.URIHost)
	case "from_user":
		return WriteJSONString(w, hep.SIP.FromUser)
	case "from_domain":
		return WriteJSONString(w, hep.SIP.FromHost)
	case "from_tag":
		return WriteJSONString(w, hep.SIP.FromTag)
	case "to_user":
		return WriteJSONString(w, hep.SIP.ToUser)
	case "to_domain":
		return WriteJSONString(w, hep.SIP.ToHost)
	case "to_tag":
		return WriteJSONString(w, hep.SIP.ToTag)
	case "via":
		return WriteJSONString(w, hep.SIP.ViaOne)
	case "contact_user":
		return WriteJSONString(w, hep.SIP.ContactUser)
	case "contact_domain":
		return WriteJSONString(w, hep.SIP.ContactHost)
	case "user_agent":
		return WriteJSONString(w, hep.SIP.UserAgent)
	case "pid_user":
		return WriteJSONString(w, hep.SIP.PaiUser)
	case "auth_user":
		return WriteJSONString(w, hep.SIP.AuthUser)
	case "server":
		return WriteJSONString(w, hep.SIP.Server)
	case "content_type":
		return WriteJSONString(w, hep.SIP.ContentType)
	case "reason":
		return WriteJSONString(w, hep.SIP.ReasonVal)
	case "diversion":
		return WriteJSONString(w, hep.SIP.DiversionVal)
	case "expires":
		return WriteJSONString(w, hep.SIP.Expires)
	default:
		return w.Write(strEmpty)
	}
}

func WriteJSONString(w io.Writer, s string) (int, error) {
	write := w.Write
	b := stb(s)
	j := 0
	n := len(b)
	if n > 0 {
		// Hint the compiler to remove bounds checks in the loop below.
		_ = b[n-1]
	}
	for i := 0; i < n; i++ {
		switch b[i] {
		case '"':
			write(b[j:i])
			write(strBackslashQuote)
			j = i + 1
		case '\\':
			write(b[j:i])
			write(strBackslashBackslash)
			j = i + 1
		case '\n':
			write(b[j:i])
			write(strBackslashN)
			j = i + 1
		case '\r':
			write(b[j:i])
			write(strBackslashR)
			j = i + 1
		case '\t':
			write(b[j:i])
			write(strBackslashT)
			j = i + 1
		case '\f':
			write(b[j:i])
			write(strBackslashF)
			j = i + 1
		case '\b':
			write(b[j:i])
			write(strBackslashB)
			j = i + 1
		default:
			if b[i] < 32 {
				write(b[j:i])
				fmt.Fprintf(w, "\\u%0.4x", b[i])
				j = i + 1
				continue
			}
		}
	}
	return write(b[j:])
}

func stb(s string) []byte {
	sh := (*reflect.StringHeader)(unsafe.Pointer(&s))
	bh := reflect.SliceHeader{
		Data: sh.Data,
		Len:  sh.Len,
		Cap:  sh.Len,
	}
	return *(*[]byte)(unsafe.Pointer(&bh))
}
