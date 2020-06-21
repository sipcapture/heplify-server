package decoder

import (
	"errors"

	"github.com/sipcapture/heplify-server/config"
	"github.com/sipcapture/heplify-server/sipparser"
)

func (h *HEP) parseSIP() error {
	h.SIP = sipparser.ParseMsg(h.Payload, config.Setting.AlegIDs, config.Setting.CustomHeader)

	if h.SIP.Error != nil {
		return h.SIP.Error
	} else if len(h.SIP.CseqMethod) < 3 {
		return errors.New("could not find a valid CSeq in packet")
	} else if len(h.SIP.CallID) < 1 {
		return errors.New("could not find a valid Call-ID in packet")
	}
	if h.SIP.FirstMethod == "" {
		h.SIP.FirstMethod = h.SIP.FirstResp
	}

	switch h.SIP.CseqMethod {
	case "INVITE", "ACK", "BYE", "CANCEL", "UPDATE", "PRACK", "REFER", "INFO":
		h.SIP.Profile = "call"
	case "REGISTER":
		h.SIP.Profile = "registration"
	default:
		h.SIP.Profile = "default"
	}

	if h.CID == "" {
		if h.SIP.XCallID != "" {
			h.CID = h.SIP.XCallID
		} else {
			h.CID = h.SIP.CallID
		}
	}

	h.SID = h.SIP.CallID

	return nil
}
