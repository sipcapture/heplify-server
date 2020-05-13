package decoder

import (
	"encoding/binary"
	"fmt"
	"net"
	"strconv"
)

func (hep *HEP) parseHEP(packet []byte) error {
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
		case IPVersion, TransportProto, AppProto:
			if len(chunkBody) != 1 {
				return fmt.Errorf("HEP chunkType %d should be 1 byte long but is %d", chunkType, len(chunkBody))
			}
		case SourcePort, DestPort, Vlan:
			if len(chunkBody) != 2 {
				return fmt.Errorf("HEP chunkType %d should be 2 byte long but is %d", chunkType, len(chunkBody))
			}
		case IP4SourceIP, IP4DestIP, TimeSec, TimeMicrosec, NodeID:
			if len(chunkBody) != 4 {
				return fmt.Errorf("HEP chunkType %d should be 4 byte long but is %d", chunkType, len(chunkBody))
			}
		case IP6SourceIP, IP6DestIP:
			if len(chunkBody) != 16 {
				return fmt.Errorf("HEP chunkType %d should be 16 byte long but is %d", chunkType, len(chunkBody))
			}
		}

		switch chunkType {
		case IPVersion:
			hep.IPVersion = uint32(chunkBody[0])
		case TransportProto:
			hep.TransportProto = uint32(chunkBody[0])
		case IP4SourceIP:
			hep.SourceIP = net.IP(chunkBody).To4().String()
		case IP4DestIP:
			hep.DestIP = net.IP(chunkBody).To4().String()
		case IP6SourceIP:
			hep.SourceIP = net.IP(chunkBody).To16().String()
		case IP6DestIP:
			hep.DestIP = net.IP(chunkBody).To16().String()
		case SourcePort:
			hep.SourcePort = uint32(binary.BigEndian.Uint16(chunkBody))
		case DestPort:
			hep.DestPort = uint32(binary.BigEndian.Uint16(chunkBody))
		case TimeSec:
			hep.TimeSec = binary.BigEndian.Uint32(chunkBody)
		case TimeMicrosec:
			hep.TimeMicrosec = binary.BigEndian.Uint32(chunkBody)
		case AppProto:
			hep.AppProto = uint32(chunkBody[0])
			switch hep.AppProto {
			case 1:
				hep.ProtoAsString = "sip"
			case 4:
				hep.ProtoAsString = "rtp"
			case 5:
				hep.ProtoAsString = "rtcp"
			case 34:
				hep.ProtoAsString = "rtpagent"
			case 35:
				hep.ProtoAsString = "rtcpxr"
			case 38:
				hep.ProtoAsString = "horaclifix"
			case 53:
				hep.ProtoAsString = "dns"
			case 100:
				hep.ProtoAsString = "log"
			default:
				hep.ProtoAsString = strconv.Itoa(int(hep.AppProto))
			}
		case NodeID:
			hep.NodeID = binary.BigEndian.Uint32(chunkBody)
		case NodePW:
			hep.NodePW = string(chunkBody)
		case Payload:
			hep.Payload = string(chunkBody)
		case CorrelationID:
			hep.CorrelationID = string(chunkBody)
		case Vlan:
			hep.Vlan = uint32(binary.BigEndian.Uint16(chunkBody))
		case NodeName:
			hep.NodeName = string(chunkBody)
		default:
		}
		currentByte += chunkLength
	}
	return nil
}
