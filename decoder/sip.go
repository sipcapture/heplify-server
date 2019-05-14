package decoder

import (
	"errors"

	"github.com/aqsyonas/sipparser"
	"github.com/sipcapture/heplify-server/config"
)

func (h *HEP) parseSIP() error {
	h.SIP = sipparser.ParseMsg(h.Payload, config.Setting.AlegIDs, config.Setting.CustomHeaders...)
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
	h.CID = h.SIP.CallID
	return nil
}
