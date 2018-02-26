package decoder

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"net"
	"time"
	"unicode/utf8"

	"github.com/negbie/heplify-server/logp"
	"github.com/negbie/sipparser"
)

// The first 4 bytes are the string "HEP3". The next 2 bytes are the length of the
// whole message (len("HEP3") + length of all the chucks we have. The next bytes
// are all the chuncks created by makeChuncks()
// Bytes: 0 1 2 3 4 5 6 7 8 9 10 11 12 13 14 15 16 17 18 19 20 21 22 23 24 25 26 27 28 29 30 31......
//        +-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
//        | "HEP3"|len|chuncks(0x0001|0x0002|0x0003|0x0004|0x0007|0x0008|0x0009|0x000a|0x000b|......)
//        +-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+

var (
	hepVer   = []byte{0x48, 0x45, 0x50, 0x33} // "HEP3"
	hepLen   = []byte{0x00, 0x00}
	hepLen7  = []byte{0x00, 0x07}
	hepLen8  = []byte{0x00, 0x08}
	hepLen10 = []byte{0x00, 0x0a}
	chunck16 = []byte{0x00, 0x00}
	chunck32 = []byte{0x00, 0x00, 0x00, 0x00}
)

// HEP chuncks
const (
	_                 = iota
	Version           // Chunk 0x0001 IP protocol family (0x02=IPv4, 0x0a=IPv6)
	Protocol          // Chunk 0x0002 IP protocol ID (0x06=TCP, 0x11=UDP)
	IP4SrcIP          // Chunk 0x0003 IPv4 source address
	IP4DstIP          // Chunk 0x0004 IPv4 destination address
	IP6SrcIP          // Chunk 0x0005 IPv6 source address
	IP6DstIP          // Chunk 0x0006 IPv6 destination address
	SrcPort           // Chunk 0x0007 Protocol source port
	DstPort           // Chunk 0x0008 Protocol destination port
	Tsec              // Chunk 0x0009 Unix timestamp, seconds
	Tmsec             // Chunk 0x000a Unix timestamp, microseconds
	ProtoType         // Chunk 0x000b Protocol type (DNS, LOG, RTCP, SIP)
	NodeID            // Chunk 0x000c Capture client ID
	KeepAliveTimer    // Chunk 0x000d Keep alive timer
	NodePW            // Chunk 0x000e Authentication key (plain text / TLS connection)
	Payload           // Chunk 0x000f Captured packet payload
	CompressedPayload // Chunk 0x0010 Captured compressed payload (gzip/inflate)
	CorrelationID     // Chunk 0x0011 Correlation ID
	Vlan              // Chunk 0x0012 VLAN
)

// HEPPacket represents a parsed HEP packet
type HEPPacket struct {
	Version           byte
	Protocol          byte
	SrcIP             string
	DstIP             string
	SrcPort           uint16
	DstPort           uint16
	Tsec              uint32
	Tmsec             uint32
	Timestamp         time.Time
	ProtoType         byte
	NodeID            uint32
	KeepAliveTimer    uint16
	NodePW            string
	Payload           string
	CompressedPayload string
	CorrelationID     string
	Vlan              uint16
	SipMsg            *sipparser.SipMsg
}

// SIPPacket represents a parsed SIP packet
type SIPPacket struct {
	Method          string
	ReplyReason     string
	Ruri            string
	RuriUser        string
	RuriDomain      string
	FromUser        string
	FromDomain      string
	FromTag         string
	ToUser          string
	ToDomain        string
	ToTag           string
	PidUser         string
	ContactUser     string
	AuthUser        string
	CallID          string
	CallIDAleg      string
	Via1            string
	Via1Branch      string
	Cseq            string
	Diversion       string
	Reason          string
	ContentType     string
	Auth            string
	UserAgent       string
	SourceIP        string
	SourcePort      string
	DestinationIP   string
	DestinationPort string
	ContactIP       string
	ContactPort     string
	OriginatorIP    string
	OriginatorPort  string
	Proto           string
	Family          string
	RTPStat         string
	Type            string
	NodeID          string
	CorrelationID   string
	Raw             *sipparser.SipMsg
}

// DecodeHEP returns a parsed HEP message
func DecodeHEP(packet []byte) (*HEPPacket, error) {
	newHepMsg := &HEPPacket{}
	err := newHepMsg.parse(packet)
	if err != nil {
		return nil, err
	}
	return newHepMsg, nil
}

func (h *HEPPacket) parse(packet []byte) error {
	if packet[0] == 0x48 && packet[2] == 0x50 && packet[3] == 0x33 {
		err := h.parseHEP(packet)
		if err != nil {
			return err
		}

		h.Timestamp = time.Unix(int64(h.Tsec), int64(h.Tmsec*1000))

		if h.ProtoType == 1 && len(h.Payload) > 64 {
			err = h.parseSIP()
			if err != nil {
				logp.Err("%v", h.SipMsg.Error)
				logp.Err("%v", h.SipMsg.Msg)
				return err
			}
		}

		logp.Debug("hep", "%#v", h)

		return nil
	}
	return errors.New("Not a valid HEP3 packet")
}

func (h *HEPPacket) parseHEP(packet []byte) error {
	var netSrcIP net.IP
	var netDstIP net.IP
	length := binary.BigEndian.Uint16(packet[4:6])
	currentByte := uint16(6)
	for currentByte < length {
		hepChunk := packet[currentByte:]
		//chunkVendorId := binary.BigEndian.Uint16(hepChunk[:2])
		chunkType := binary.BigEndian.Uint16(hepChunk[2:4])
		chunkLength := binary.BigEndian.Uint16(hepChunk[4:6])
		chunkBody := hepChunk[6:chunkLength]

		switch chunkType {
		case Version:
			h.Version = chunkBody[0]
		case Protocol:
			h.Protocol = chunkBody[0]
		case IP4SrcIP:
			netSrcIP = chunkBody
			h.SrcIP = netSrcIP.String()
		case IP4DstIP:
			netDstIP = chunkBody
			h.DstIP = netDstIP.String()
		case IP6SrcIP:
			netSrcIP = chunkBody
			h.SrcIP = netSrcIP.String()
		case IP6DstIP:
			netDstIP = chunkBody
			h.DstIP = netDstIP.String()
		case SrcPort:
			h.SrcPort = binary.BigEndian.Uint16(chunkBody)
		case DstPort:
			h.DstPort = binary.BigEndian.Uint16(chunkBody)
		case Tsec:
			h.Tsec = binary.BigEndian.Uint32(chunkBody)
		case Tmsec:
			h.Tmsec = binary.BigEndian.Uint32(chunkBody)
		case ProtoType:
			h.ProtoType = chunkBody[0]
		case NodeID:
			h.NodeID = binary.BigEndian.Uint32(chunkBody)
		case KeepAliveTimer:
			h.KeepAliveTimer = binary.BigEndian.Uint16(chunkBody)
		case NodePW:
			h.NodePW = string(chunkBody)
		case Payload:
			h.Payload = string(chunkBody)
			if !utf8.ValidString(h.Payload) {
				v := make([]rune, 0, len(h.Payload))
				for i, r := range h.Payload {
					if r == utf8.RuneError {
						_, size := utf8.DecodeRuneInString(h.Payload[i:])
						if size == 1 {
							continue
						}
					}
					v = append(v, r)
				}
				h.Payload = string(v)
			}
		case CompressedPayload:
			h.CompressedPayload = string(chunkBody)
		case CorrelationID:
			h.CorrelationID = string(chunkBody)
		case Vlan:
			h.Vlan = binary.BigEndian.Uint16(chunkBody)
		default:
		}
		currentByte += chunkLength
	}
	return nil
}

func (h *HEPPacket) parseSIP() error {

	h.SipMsg = sipparser.ParseMsg(h.Payload)

	if h.SipMsg.StartLine == nil {
		h.SipMsg.StartLine = new(sipparser.StartLine)
	}
	if h.SipMsg.StartLine.Method == "" {
		h.SipMsg.StartLine.Method = h.SipMsg.StartLine.Resp
	}
	if h.SipMsg.StartLine.URI == nil {
		h.SipMsg.StartLine.URI = new(sipparser.URI)
	}
	if h.SipMsg.From == nil {
		h.SipMsg.From = new(sipparser.From)
	}
	if h.SipMsg.From.URI == nil {
		h.SipMsg.From.URI = new(sipparser.URI)
	}
	if h.SipMsg.To == nil {
		h.SipMsg.To = new(sipparser.From)
	}
	if h.SipMsg.To.URI == nil {
		h.SipMsg.To.URI = new(sipparser.URI)
	}
	if h.SipMsg.Contact == nil {
		h.SipMsg.Contact = new(sipparser.From)
	}
	if h.SipMsg.Contact.URI == nil {
		h.SipMsg.Contact.URI = new(sipparser.URI)
	}
	if h.SipMsg.Authorization == nil {
		h.SipMsg.Authorization = new(sipparser.Authorization)
	}
	if h.SipMsg.Via == nil {
		h.SipMsg.Via = make([]*sipparser.Via, 1)
		h.SipMsg.Via[0] = new(sipparser.Via)
	}
	if h.SipMsg.Cseq == nil {
		h.SipMsg.Cseq = new(sipparser.Cseq)
	}

	if h.SipMsg.Error != nil {
		return h.SipMsg.Error
	}

	return nil
}

func (h *HEPPacket) String() {
	fmt.Printf("Version: \t %d \n", h.Version)
	fmt.Printf("Protocol: \t %d \n", h.Protocol)
	fmt.Printf("ProtoType: \t %d \n", h.ProtoType)
	fmt.Printf("SrcIP: \t\t %s \n", h.SrcIP)
	fmt.Printf("DstIP: \t\t %s \n", h.DstIP)
	fmt.Printf("SrcPort: \t %d \n", h.SrcPort)
	fmt.Printf("DstPort: \t %d \n", h.DstPort)
	fmt.Printf("Tsec: \t\t %d \n", h.Tsec)
	fmt.Printf("Tmsec: \t\t %d \n", h.Tmsec)
	fmt.Printf("Vlan: \t\t %d \n", h.Vlan)
	fmt.Printf("NodeID: \t %d \n", h.NodeID)
	fmt.Printf("NodePW: \t %s \n", string(h.NodePW))
	fmt.Printf("KeepAliveTimer:  %d \n", h.KeepAliveTimer)
	fmt.Printf("CorrelationID:   %s \n", string(h.CorrelationID))
	fmt.Printf("Payload: \n%s\n", string(h.Payload))
}

// EncodeHEP creates the HEP Packet which
// will be send to wire
func EncodeHEP(h *HEPPacket) []byte {
	buf := new(bytes.Buffer)
	hepMsg := makeChuncks(h, buf)
	binary.BigEndian.PutUint16(hepMsg[4:6], uint16(len(hepMsg)))
	return hepMsg
}

// makeChuncks will construct the respective HEP chunck
func makeChuncks(h *HEPPacket, w *bytes.Buffer) []byte {
	w.Write(hepVer)
	// hepMsg length placeholder. Will be written later
	w.Write(hepLen)

	// Chunk IP protocol family (0x02=IPv4, 0x0a=IPv6)
	w.Write([]byte{0x00, 0x00, 0x00, 0x01})
	w.Write(hepLen7)
	w.WriteByte(h.Version)

	// Chunk IP protocol ID (0x06=TCP, 0x11=UDP)
	w.Write([]byte{0x00, 0x00, 0x00, 0x02})
	w.Write(hepLen7)
	w.WriteByte(h.Protocol)

	if h.Version == 0x02 {
		// Chunk IPv4 source address
		w.Write([]byte{0x00, 0x00, 0x00, 0x03})
		binary.BigEndian.PutUint16(hepLen, 6+uint16(len(h.SrcIP)))
		w.Write(hepLen)
		w.Write([]byte(h.SrcIP))

		// Chunk IPv4 destination address
		w.Write([]byte{0x00, 0x00, 0x00, 0x04})
		binary.BigEndian.PutUint16(hepLen, 6+uint16(len(h.DstIP)))
		w.Write(hepLen)
		w.Write([]byte(h.DstIP))
	} else if h.Version == 0x0a {
		// Chunk IPv6 source address
		w.Write([]byte{0x00, 0x00, 0x00, 0x05})
		binary.BigEndian.PutUint16(hepLen, 6+uint16(len(h.SrcIP)))
		w.Write(hepLen)
		w.Write([]byte(h.SrcIP))

		// Chunk IPv6 destination address
		w.Write([]byte{0x00, 0x00, 0x00, 0x06})
		binary.BigEndian.PutUint16(hepLen, 6+uint16(len(h.DstIP)))
		w.Write(hepLen)
		w.Write([]byte(h.DstIP))
	}

	// Chunk protocol source port
	w.Write([]byte{0x00, 0x00, 0x00, 0x07})
	w.Write(hepLen8)
	binary.BigEndian.PutUint16(chunck16, h.SrcPort)
	w.Write(chunck16)

	// Chunk protocol destination port
	w.Write([]byte{0x00, 0x00, 0x00, 0x08})
	w.Write(hepLen8)
	binary.BigEndian.PutUint16(chunck16, h.DstPort)
	w.Write(chunck16)

	// Chunk unix timestamp, seconds
	w.Write([]byte{0x00, 0x00, 0x00, 0x09})
	w.Write(hepLen10)
	binary.BigEndian.PutUint32(chunck32, h.Tsec)
	w.Write(chunck32)

	// Chunk unix timestamp, microseconds offset
	w.Write([]byte{0x00, 0x00, 0x00, 0x0a})
	w.Write(hepLen10)
	binary.BigEndian.PutUint32(chunck32, h.Tmsec)
	w.Write(chunck32)

	// Chunk protocol type (DNS, LOG, RTCP, SIP)
	w.Write([]byte{0x00, 0x00, 0x00, 0x0b})
	w.Write(hepLen7)
	w.WriteByte(h.ProtoType)

	// Chunk capture agent ID
	w.Write([]byte{0x00, 0x00, 0x00, 0x0c})
	w.Write(hepLen10)
	binary.BigEndian.PutUint32(chunck32, h.NodeID)
	w.Write(chunck32)

	// Chunk keep alive timer
	//w.Write([]byte{0x00, 0x00, 0x00, 0x0d})

	// Chunk authenticate key (plain text / TLS connection)
	w.Write([]byte{0x00, 0x00, 0x00, 0x0e})
	binary.BigEndian.PutUint16(hepLen, 6+uint16(len(h.NodePW)))
	w.Write(hepLen)
	w.Write([]byte(h.NodePW))

	// Chunk captured packet payload
	w.Write([]byte{0x00, 0x00, 0x00, 0x0f})
	binary.BigEndian.PutUint16(hepLen, 6+uint16(len(h.Payload)))
	w.Write(hepLen)
	w.Write([]byte(h.Payload))

	// Chunk captured compressed payload (gzip/inflate)
	//w.Write([]byte{0x00,0x00, 0x00,0x10})

	if h.CorrelationID != "" {
		// Chunk internal correlation id
		w.Write([]byte{0x00, 0x00, 0x00, 0x11})
		binary.BigEndian.PutUint16(hepLen, 6+uint16(len(h.CorrelationID)))
		w.Write(hepLen)
		w.Write([]byte(h.CorrelationID))
	}
	/*
		// Chunk VLAN
		w.Write([]byte{0x00, 0x00, 0x00, 0x12})
		w.Write(hepLen8)
		binary.BigEndian.PutUint16(chunck16, h.Vlan)
		w.Write(chunck16)

		// Chunk MOS only
		w.Write([]byte{0x00, 0x00, 0x00, 0x20})
		w.Write(hepLen8)
		binary.BigEndian.PutUint16(chunck16, h.MOS)
		w.Write(chunck16)
	*/
	return w.Bytes()
}
