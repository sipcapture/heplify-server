package decoder

import (
	"time"

	"github.com/valyala/fastjson"
)

var pp fastjson.ParserPool

func (h *HEP) parseWebRTC(packet []byte) error {
	p := pp.Get()
	defer pp.Put(p)
	v, err := p.ParseBytes(packet)
	if err != nil {
		return err
	}
	h.ProtoType = uint32(v.GetInt("type") * 1000)
	if h.ProtoType < 1000 {
		h.ProtoType = 1000
	}
	h.Timestamp = time.Unix(0, v.GetInt64("timestamp")*1000)
	if e := v.Get("event"); e != nil {
		h.Payload = e.String()
	}
	if s := v.Get("session_id"); s != nil {
		h.CID = s.String()
	} else {
		h.CID = "000000000000"
	}
	return nil
}
