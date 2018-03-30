package decoder

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"net"
	"strconv"
	"time"
	"unicode/utf8"

	"github.com/coocood/freecache"
	raven "github.com/getsentry/raven-go"
	"github.com/negbie/heplify-server/config"
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
	dedup    = freecache.NewCache(10 * 1024 * 1024)
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
)

// HEP represents a parsed HEP packet
type HEP struct {
	Version     byte
	Protocol    byte
	SrcIP       net.IP
	DstIP       net.IP
	SrcIPString string
	DstIPString string
	SrcPort     uint16
	DstPort     uint16
	Tsec        uint32
	Tmsec       uint32
	Timestamp   time.Time
	ProtoType   byte
	NodeID      uint32
	NodePW      string
	Payload     string
	CID         string
	Vlan        uint16
	SIP         *sipparser.SipMsg
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
	if packet[0] != 0x48 || packet[1] != 0x45 || packet[2] != 0x50 || packet[3] != 0x33 {
		return errors.New("Not a valid HEP3 packet")
	}

	err := h.parseHEP(packet)
	if h.ProtoType == 0 {
		return nil
	} else if err != nil {
		logp.Warn("%v", err)
		if config.Setting.SentryDSN != "" {
			raven.CaptureError(err, nil)
		}
		return err
	}

	h.Timestamp = time.Unix(int64(h.Tsec), int64(h.Tmsec*1000))

	if h.ProtoType == 1 && len(h.Payload) > 64 {
		err = h.parseSIP()
		if err != nil {
			logp.Warn("%v\n%s\n\n", err, strconv.Quote(h.Payload))
			return err
		}
		h.CID = h.SIP.CallID
	}

	logp.Debug("hep", "%+v\n\n", h)
	return nil
}

func (h *HEP) parseHEP(packet []byte) error {
	length := binary.BigEndian.Uint16(packet[4:6])
	if int(length) != len(packet) {
		return fmt.Errorf("HEP packet length is %d but should be %d", len(packet), length)
	}
	currentByte := uint16(6)

	for currentByte < length {
		hepChunk := packet[currentByte:]
		//chunkVendorId := binary.BigEndian.Uint16(hepChunk[:2])
		chunkType := binary.BigEndian.Uint16(hepChunk[2:4])
		chunkLength := binary.BigEndian.Uint16(hepChunk[4:6])
		if len(hepChunk) < int(chunkLength) {
			return fmt.Errorf("HEP chunk overflow %d > %d", chunkLength, len(hepChunk))
		}
		chunkBody := hepChunk[6:chunkLength]

		switch chunkType {
		case Version:
			h.Version = chunkBody[0]
		case Protocol:
			h.Protocol = chunkBody[0]
		case IP4SrcIP:
			h.SrcIP = chunkBody
			h.SrcIPString = h.SrcIP.String()
		case IP4DstIP:
			h.DstIP = chunkBody
			h.DstIPString = h.DstIP.String()
		case IP6SrcIP:
			h.SrcIP = chunkBody
			h.SrcIPString = h.SrcIP.String()
		case IP6DstIP:
			h.DstIP = chunkBody
			h.DstIPString = h.DstIP.String()
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
		case NodePW:
			h.NodePW = string(chunkBody)
		case Payload:
			h.Payload = h.normPayload(chunkBody)
		case CID:
			h.CID = string(chunkBody)
		case Vlan:
			h.Vlan = binary.BigEndian.Uint16(chunkBody)
		default:
		}
		currentByte += chunkLength
	}
	return nil
}

func (h *HEP) parseSIP() error {
	h.SIP = sipparser.ParseMsg(h.Payload)

	if h.SIP.StartLine == nil {
		h.SIP.StartLine = new(sipparser.StartLine)
	}
	if h.SIP.StartLine.URI == nil {
		h.SIP.StartLine.URI = new(sipparser.URI)
	}
	if h.SIP.StartLine.Method == "" {
		h.SIP.StartLine.Method = h.SIP.StartLine.Resp
	}

	if h.SIP.Error != nil {
		return h.SIP.Error
	} else if len(h.SIP.Cseq.Method) < 3 {
		return errors.New("Could not find a valid CSeq in packet")
	} else if len(h.SIP.CallID) < 3 {
		return errors.New("Could not find a valid Call-ID in packet")
	}

	return nil
}

// EncodeHEP creates the HEP Packet which
// will be send to wire
func EncodeHEP(h *HEP) []byte {
	buf := new(bytes.Buffer)
	hepMsg := makeChuncks(h, buf)
	binary.BigEndian.PutUint16(hepMsg[4:6], uint16(len(hepMsg)))
	return hepMsg
}

// makeChuncks will construct the respective HEP chunck
func makeChuncks(h *HEP, w *bytes.Buffer) []byte {
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

	if h.CID != "" && h.ProtoType != 1 {
		// Chunk internal correlation id
		w.Write([]byte{0x00, 0x00, 0x00, 0x11})
		binary.BigEndian.PutUint16(hepLen, 6+uint16(len(h.CID)))
		w.Write(hepLen)
		w.Write([]byte(h.CID))
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

func (h *HEP) normPayload(pb []byte) string {
	if config.Setting.Dedup {
		_, err := dedup.Get(pb)
		if err == nil {
			h.ProtoType = 0
			return ""
		}
		err = dedup.Set(pb, nil, 1)
		if err != nil {
			logp.Warn("%v", err)
		}
	}
	if !utf8.Valid(pb) {
		v := make([]rune, 0, len(pb))
		for i, r := range pb {
			if rune(r) == utf8.RuneError {
				_, size := utf8.DecodeRune(pb[i:])
				if size == 1 {
					continue
				}
			}
			v = append(v, rune(r))
		}
		return string(v)
	}
	return string(pb)
}
