package decoder

import (
	"errors"

	"github.com/sipcapture/heplify-server/config"
	"github.com/sipcapture/heplify-server/sipparser"
)

func (hep *HEP) parseSIP() error {
	hep.SIP = sipparser.ParseMsg(hep.Payload, config.Setting.AlegIDs, config.Setting.CustomHeader)
	if hep.SIP.Error != nil {
		return hep.SIP.Error
	} else if len(hep.SIP.CseqMethod) < 3 {
		return errors.New("could not find a valid CSeq in packet")
	} else if len(hep.SIP.CallID) < 1 {
		return errors.New("could not find a valid Call-ID in packet")
	}
	if hep.SIP.FirstMethod == "" {
		hep.SIP.FirstMethod = hep.SIP.FirstResp
	}

	if hep.CorrelationID == "" {
		if hep.SIP.XCallID != "" {
			hep.CorrelationID = hep.SIP.XCallID
		} else {
			hep.CorrelationID = hep.SIP.CallID
		}
	}

	hep.SID = hep.SIP.CallID

	return nil
}
