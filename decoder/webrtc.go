package decoder

import (
	"time"

	"github.com/valyala/fastjson"
)

var pp fastjson.ParserPool

//var event = []byte(`{"emitter":"MyJanusInstance","type":2,"timestamp":1524562112570863,"session_id":8879576811135393,"handle_id":6836738107974166,"opaque_id":"audiobridgetest-Kh3w3TQhZ9wm","event":{"name":"attached","plugin":"janus.plugin.audiobridge","opaque_id":"audiobridgetest-Kh3w3TQhZ9wm"}}`)

func (h *HEP) parseWebRTC(packet []byte) error {
	p := pp.Get()
	defer pp.Put(p)
	v, err := p.ParseBytes(packet)
	if err != nil {
		return err
	}

	h.ProtoType = uint32(v.GetInt("type")) + 1000
	t := v.GetInt64("timestamp") * 1000
	if t != 0 {
		h.Timestamp = time.Unix(0, t)
	} else {
		h.Timestamp = time.Now()
	}
	if s := v.Get("session_id"); s != nil {
		h.SrcIP = s.String()
	}
	if s := v.Get("handle_id"); s != nil {
		h.DstIP = s.String()
	}
	if e := v.Get("event"); e != nil {
		h.Payload = e.String()
	} else {
		h.Payload = string(packet)
	}
	if s := v.Get("opaque_id"); s != nil {
		h.CID = s.String()
	} else {
		h.CID = "webRTCBingo"
	}
	return nil
}
