package input

import (
	"strings"
	"sync/atomic"

	"github.com/negbie/heplify-server/config"
	"github.com/negbie/heplify-server/decoder"
	"github.com/negbie/logp"
	"github.com/valyala/fasthttp"
)

func (h *HEPInput) serveHTTP() {
	addr := config.Setting.HTTPAddr
	addr = strings.TrimPrefix(addr, "https://")
	addr = strings.TrimPrefix(addr, "http://")
	server := &fasthttp.Server{
		Handler: h.requestHandler,
	}
	done := make(chan bool)
	go func() {
		<-h.quit
		if err := server.Shutdown(); err != nil {
			logp.Err("could not gracefully shutdown HTTP server: %v\n", err)
		}
		close(done)
	}()

	if err := server.ListenAndServe(addr); err != nil {
		logp.Err("could not listen on %s: %v\n", addr, err)
	}
	<-done
}

func (h *HEPInput) requestHandler(ctx *fasthttp.RequestCtx) {
	ctx.Request.AppendBodyString("WEBRTC")
	hepPkt, err := decoder.DecodeHEP(ctx.Request.Body())
	if err != nil {
		atomic.AddUint64(&h.stats.ErrCount, 1)
		return
	}
	if h.useLK {
		select {
		case h.lkCh <- hepPkt:
		default:
			logp.Warn("overflowing loki channel")
		}
	}
	if h.usePM && hepPkt.ProtoType == 1032 {
		select {
		case h.pmCh <- hepPkt:
		default:
			logp.Warn("overflowing metric channel")
		}
	}
}
