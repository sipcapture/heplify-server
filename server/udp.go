package input

import (
	"net"
	"sync/atomic"
	"time"

	"github.com/negbie/heplify-server/config"
	"github.com/negbie/logp"
)

func (h *HEPInput) serveUDP() {
	ua, err := net.ResolveUDPAddr("udp", config.Setting.HEPAddr)
	if err != nil {
		logp.Critical("%v", err)
	}

	uc, err := net.ListenUDP("udp", ua)
	if err != nil {
		logp.Critical("%v", err)
	}
	defer uc.Close()
	defer h.wg.Done()
	for {
		select {
		case <-h.quit:
			return
		default:
		}
		uc.SetReadDeadline(time.Now().Add(1e9))
		buf := h.buffer.Get().([]byte)
		n, err := uc.Read(buf)
		if err != nil {
			continue
		} else if n > maxPktLen {
			logp.Warn("received too big packet with %d bytes", n)
			atomic.AddUint64(&h.stats.ErrCount, 1)
			continue
		}
		h.inCh <- buf[:n]
		atomic.AddUint64(&h.stats.PktCount, 1)
	}
}
