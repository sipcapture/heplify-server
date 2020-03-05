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
// whole message (len("HEP3") + length of all the chucks we have. The next bytes
// are all the chuncks created by makeChuncks()
// Bytes: 0 1 2 3 4 5 6 7 8 9 10 11 12 13 14 15 16 17 18 19 20 21 22 23 24 25 26 27 28 29 30 31......
//        +-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
//        | "HEP3"|len|chuncks(0x0001|0x0002|0x0003|0x0004|0x0007|0x0008|0x0009|0x000a|0x000b|......)
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

// HEP chuncks
const (
	Version   = 1  // Chunk 0x0001 IP protocol family (0x02=IPv4, 0x0a=IPv6)
	Protocol  = 2  // Chunk 0x0002 IP protocol ID (0x06=TCP, 0x11=UDP)
	IP4SrcIP  = 3  // Chunk 0x0003 IPv4 source address
	IP4DstIP  = 4  // Chunk 0x0004 IPv4 destination address
	IP6SrcIP  = 5  // Chunk 0x0005 IPv6 source address
	IP6DstIP  = 6  // Chunk 0x0006 IPv6 destination address
	SrcPort   = 7  // Chunk 0x0007 Protocol source port
	DstPort   = 8  // Chunk 0x0008 Protocol destination port
	Tsec      = 9  // Chunk 0x0009 Unix timestamp, seconds
	Tmsec     = 10 // Chunk 0x000a Unix timestamp, microseconds
	ProtoType = 11 // Chunk 0x000b Protocol type (DNS, LOG, RTCP, SIP)
	NodeID    = 12 // Chunk 0x000c Capture client ID
	NodePW    = 14 // Chunk 0x000e Authentication key (plain text / TLS connection)
	Payload   = 15 // Chunk 0x000f Captured packet payload
	CID       = 17 // Chunk 0x0011 Correlation ID
	Vlan      = 18 // Chunk 0x0012 VLAN
	NodeName  = 19 // Chunk 0x0013 NodeName
)

// HEP represents HEP packet
type HEP struct {
	Version     uint32 `protobuf:"varint,1,req,name=Version" json:"Version"`
	Protocol    uint32 `protobuf:"varint,2,req,name=Protocol" json:"Protocol"`
	SrcIP       string `protobuf:"bytes,3,req,name=SrcIP" json:"SrcIP"`
	DstIP       string `protobuf:"bytes,4,req,name=DstIP" json:"DstIP"`
	SrcPort     uint32 `protobuf:"varint,5,req,name=SrcPort" json:"SrcPort"`
	DstPort     uint32 `protobuf:"varint,6,req,name=DstPort" json:"DstPort"`
	Tsec        uint32 `protobuf:"varint,7,req,name=Tsec" json:"Tsec"`
	Tmsec       uint32 `protobuf:"varint,8,req,name=Tmsec" json:"Tmsec"`
	ProtoType   uint32 `protobuf:"varint,9,req,name=ProtoType" json:"ProtoType"`
	NodeID      uint32 `protobuf:"varint,10,req,name=NodeID" json:"NodeID"`
	NodePW      string `protobuf:"bytes,11,req,name=NodePW" json:"NodePW"`
	Payload     string `protobuf:"bytes,12,req,name=Payload" json:"Payload"`
	CID         string `protobuf:"bytes,13,req,name=CID" json:"CID"`
	Vlan        uint32 `protobuf:"varint,14,req,name=Vlan" json:"Vlan"`
	ProtoString string
	Timestamp   time.Time
	SIP         *sipparser.SipMsg
	NodeName    string
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

func (h *HEP) parse(packet []byte) error {
	var err error
	if bytes.HasPrefix(packet, []byte{0x48, 0x45, 0x50, 0x33}) {
		err = h.parseHEP(packet)
		if err != nil {
			logp.Warn("%v", err)
			return err
		}
	} else {
		err = h.Unmarshal(packet)
		if err != nil {
			logp.Warn("malformed packet with length %d which is neither hep nor protobuf encapsulated", len(packet))
			return err
		}
	}

	t := time.Now()
	h.normPayload(t)
	if h.ProtoType == 0 {
		return nil
	}

	h.Timestamp = time.Unix(int64(h.Tsec), int64(h.Tmsec*1000))
	d := t.Sub(h.Timestamp)
	if d < 0 || (h.Tsec == 0 && h.Tmsec == 0) {
		logp.Debug("hep", "got timestamp in the future with delta: %d from nodeID %d", d, h.NodeID)
		h.Timestamp = t
	}

	if h.ProtoType == 1 && len(h.Payload) > 32 {
		err = h.parseSIP()
		if err != nil {
			logp.Warn("%v\n%q\nnodeID: %d, protoType: %d, version: %d, protocol: %d, length: %d, flow: %s:%d->%s:%d\n\n",
				err, h.Payload, h.NodeID, h.ProtoType, h.Version, h.Protocol, len(h.Payload), h.SrcIP, h.SrcPort, h.DstIP, h.DstPort)
			return err
		}

		if len(config.Setting.DiscardMethod) > 0 {
			for k := range config.Setting.DiscardMethod {
				if config.Setting.DiscardMethod[k] == h.SIP.CseqMethod {
					h.ProtoType = 0
					return nil
				}
			}
		}
	}

	if h.NodeName == "" {
		h.NodeName = strconv.FormatUint(uint64(h.NodeID), 10)
	}

	logp.Debug("hep", "%+v\n\n", h)
	return nil
}

var fixUTF8 = func(r rune) rune {
	if r == utf8.RuneError || r == '\x00' {
		return -1
	}
	return r
}

func (h *HEP) normPayload(t time.Time) {
	if config.Setting.Dedup {
		ts := uint64(t.UnixNano())
		kh := make([]byte, 8)
		ks := xxhash.Sum64String(h.Payload)
		binary.BigEndian.PutUint64(kh, ks)

		if buf := dedup.Get(nil, kh); buf != nil {
			i := binary.BigEndian.Uint64(buf)
			d := ts - i
			if d < 400e6 || d > 1e18 {
				h.ProtoType = 0
				return
			}
		}

		tb := make([]byte, 8)
		binary.BigEndian.PutUint64(tb, ts)
		dedup.Set(kh, tb)
	}
	if !utf8.ValidString(h.Payload) {
		h.Payload = strings.Map(fixUTF8, h.Payload)
	}
	if config.Setting.DBDriver == "postgres" && strings.Index(h.Payload, "\x00") > -1 {
		h.Payload = strings.Map(fixUTF8, h.Payload)
	}
}

func (h *HEP) EscapeFields(w io.Writer, tag string) (int, error) {
	switch tag {
	case "callid":
		return writeJSONString(w, h.SIP.CallID)
	case "cseq":
		return writeJSONString(w, h.SIP.CseqVal)
	case "method":
		return writeJSONString(w, h.SIP.FirstMethod)
	case "ruri_user":
		return writeJSONString(w, h.SIP.URIUser)
	case "ruri_domain":
		return writeJSONString(w, h.SIP.URIHost)
	case "from_user":
		return writeJSONString(w, h.SIP.FromUser)
	case "from_domain":
		return writeJSONString(w, h.SIP.FromHost)
	case "from_tag":
		return writeJSONString(w, h.SIP.FromTag)
	case "to_user":
		return writeJSONString(w, h.SIP.ToUser)
	case "to_domain":
		return writeJSONString(w, h.SIP.ToHost)
	case "to_tag":
		return writeJSONString(w, h.SIP.ToTag)
	case "via":
		return writeJSONString(w, h.SIP.ViaOne)
	case "contact_user":
		return writeJSONString(w, h.SIP.ContactUser)
	case "contact_domain":
		return writeJSONString(w, h.SIP.ContactHost)
	case "user_agent":
		return writeJSONString(w, h.SIP.UserAgent)
	case "pid_user":
		return writeJSONString(w, h.SIP.PaiUser)
	case "auth_user":
		return writeJSONString(w, h.SIP.AuthUser)
	case "server":
		return writeJSONString(w, h.SIP.Server)
	case "content_type":
		return writeJSONString(w, h.SIP.ContentType)
	case "reason":
		return writeJSONString(w, h.SIP.ReasonVal)
	case "diversion":
		return writeJSONString(w, h.SIP.DiversionVal)
	case "expires":
		return writeJSONString(w, h.SIP.Expires)
	default:
		return w.Write(strEmpty)
	}
}

func writeJSONString(w io.Writer, s string) (int, error) {
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
