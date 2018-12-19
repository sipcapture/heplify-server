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

	"github.com/cespare/xxhash"
	"github.com/coocood/freecache"
	"github.com/sipcapture/heplify-server/config"
	"github.com/negbie/logp"
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

// HEP represents HEP packet
type HEP struct {
	Version   uint32 `protobuf:"varint,1,req,name=Version" json:"Version"`
	Protocol  uint32 `protobuf:"varint,2,req,name=Protocol" json:"Protocol"`
	SrcIP     string `protobuf:"bytes,3,req,name=SrcIP" json:"SrcIP"`
	DstIP     string `protobuf:"bytes,4,req,name=DstIP" json:"DstIP"`
	SrcPort   uint32 `protobuf:"varint,5,req,name=SrcPort" json:"SrcPort"`
	DstPort   uint32 `protobuf:"varint,6,req,name=DstPort" json:"DstPort"`
	Tsec      uint32 `protobuf:"varint,7,req,name=Tsec" json:"Tsec"`
	Tmsec     uint32 `protobuf:"varint,8,req,name=Tmsec" json:"Tmsec"`
	ProtoType uint32 `protobuf:"varint,9,req,name=ProtoType" json:"ProtoType"`
	NodeID    uint32 `protobuf:"varint,10,req,name=NodeID" json:"NodeID"`
	NodePW    string `protobuf:"bytes,11,req,name=NodePW" json:"NodePW"`
	Payload   string `protobuf:"bytes,12,req,name=Payload" json:"Payload"`
	CID       string `protobuf:"bytes,13,req,name=CID" json:"CID"`
	Vlan      uint32 `protobuf:"varint,14,req,name=Vlan" json:"Vlan"`
	NetSrcIP  net.IP
	NetDstIP  net.IP
	Timestamp time.Time
	SIP       *sipparser.SipMsg
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
	if bytes.HasPrefix(packet, hepVer) && len(packet) > 32 {
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

	h.normPayload()
	if h.ProtoType == 0 {
		return nil
	}

	h.Timestamp = time.Unix(int64(h.Tsec), int64(h.Tmsec*1000))
	t := time.Now()
	d := t.Sub(h.Timestamp)
	if d < 0 || (h.Tsec == 0 && h.Tmsec == 0) {
		logp.Debug("heptime", "future packet timestamp: %d, now: %d, delta: %d from nodeID %d",
			h.Timestamp.UnixNano(), t.UnixNano(), d, h.NodeID)
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
					h.Payload = "DISCARD"
				}
			}
		}
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
		if len(hepChunk) < 6 {
			return fmt.Errorf("HEP chunk must be >= 6 byte long but is %d", len(hepChunk))
		}
		//chunkVendorId := binary.BigEndian.Uint16(hepChunk[:2])
		chunkType := binary.BigEndian.Uint16(hepChunk[2:4])
		chunkLength := binary.BigEndian.Uint16(hepChunk[4:6])
		if len(hepChunk) < int(chunkLength) || int(chunkLength) < 6 {
			return fmt.Errorf("HEP chunk with %d byte < chunkLength %d or chunkLength < 6", len(hepChunk), chunkLength)
		}
		chunkBody := hepChunk[6:chunkLength]

		switch chunkType {
		case Version, Protocol, ProtoType:
			if len(chunkBody) != 1 {
				return fmt.Errorf("HEP chunkType %d should be 1 byte long but is %d", chunkType, len(chunkBody))
			}
		case SrcPort, DstPort, Vlan:
			if len(chunkBody) != 2 {
				return fmt.Errorf("HEP chunkType %d should be 2 byte long but is %d", chunkType, len(chunkBody))
			}
		case IP4SrcIP, IP4DstIP, Tsec, Tmsec, NodeID:
			if len(chunkBody) != 4 {
				return fmt.Errorf("HEP chunkType %d should be 4 byte long but is %d", chunkType, len(chunkBody))
			}
		case IP6SrcIP, IP6DstIP:
			if len(chunkBody) != 16 {
				return fmt.Errorf("HEP chunkType %d should be 16 byte long but is %d", chunkType, len(chunkBody))
			}
		}

		switch chunkType {
		case Version:
			h.Version = uint32(chunkBody[0])
		case Protocol:
			h.Protocol = uint32(chunkBody[0])
		case IP4SrcIP:
			h.NetSrcIP = chunkBody
			h.SrcIP = h.NetSrcIP.String()
		case IP4DstIP:
			h.NetDstIP = chunkBody
			h.DstIP = h.NetDstIP.String()
		case IP6SrcIP:
			h.NetSrcIP = chunkBody
			h.SrcIP = h.NetSrcIP.String()
		case IP6DstIP:
			h.NetDstIP = chunkBody
			h.DstIP = h.NetDstIP.String()
		case SrcPort:
			h.SrcPort = uint32(binary.BigEndian.Uint16(chunkBody))
		case DstPort:
			h.DstPort = uint32(binary.BigEndian.Uint16(chunkBody))
		case Tsec:
			h.Tsec = binary.BigEndian.Uint32(chunkBody)
		case Tmsec:
			h.Tmsec = binary.BigEndian.Uint32(chunkBody)
		case ProtoType:
			h.ProtoType = uint32(chunkBody[0])
		case NodeID:
			h.NodeID = binary.BigEndian.Uint32(chunkBody)
		case NodePW:
			h.NodePW = string(chunkBody)
		case Payload:
			h.Payload = string(chunkBody)
		case CID:
			h.CID = string(chunkBody)
		case Vlan:
			h.Vlan = uint32(binary.BigEndian.Uint16(chunkBody))
		default:
		}
		currentByte += chunkLength
	}
	return nil
}

func (h *HEP) parseSIP() error {
	h.SIP = sipparser.ParseMsg(h.Payload, config.Setting.AlegIDs...)

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
	} else if len(h.SIP.CseqMethod) < 3 {
		return errors.New("Could not find a valid CSeq in packet")
	} else if len(h.SIP.CallID) < 1 {
		return errors.New("Could not find a valid Call-ID in packet")
	}
	h.CID = h.SIP.CallID
	return nil
}

func (h *HEP) normPayload() {
	if config.Setting.Dedup {
		hashVal := int64(xxhash.Sum64String(h.SrcIP)) + int64(h.SrcPort) + int64(xxhash.Sum64String(h.Payload))
		_, err := dedup.GetInt(hashVal)
		if err == nil {
			h.ProtoType = 0
			return
		}
		err = dedup.SetInt(hashVal, nil, 1)
		if err != nil {
			logp.Warn("%v", err)
		}
	}
	if h.ProtoType == 1 {
		if !utf8.ValidString(h.Payload) {
			v := make([]rune, 0, len(h.Payload))
			for i, r := range h.Payload {
				if r == utf8.RuneError {
					_, size := utf8.DecodeRuneInString(h.Payload[i:])
					if size == 1 {
						continue
					}
				} else if r == '\x00' {
					continue
				}
				v = append(v, r)
			}
			h.Payload = string(v)
		}
	}
}

func HEPTypeString(pktType uint32) (label string) {
	switch pktType {
	case 1:
		label = "sip"
	case 5:
		label = "rtcp"
	case 34:
		label = "rtpagent"
	case 35:
		label = "rtcpxr"
	case 38:
		label = "horaclifix"
	case 53:
		label = "dns"
	case 100:
		label = "log"
	default:
		label = strconv.Itoa(int(pktType))
	}
	return label
}
