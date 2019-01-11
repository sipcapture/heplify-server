package decoder

import (
	"errors"

	"github.com/negbie/heplify-server/config"
	"github.com/negbie/sipparser"
)

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
		return errors.New("could not find a valid CSeq in packet")
	} else if len(h.SIP.CallID) < 1 {
		return errors.New("could not find a valid Call-ID in packet")
	}
	h.CID = h.SIP.CallID
	return nil
}
