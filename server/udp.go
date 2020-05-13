package input

import (
	"net"
	"sync/atomic"
	"time"

	"github.com/negbie/logp"
)

func (hepInp *HEPInput) serveUDP(addr string) {
	ua, err := net.ResolveUDPAddr("udp", addr)
	if err != nil {
		logp.Err("%v", err)
		return
	}

	uc, err := net.ListenUDP("udp", ua)
	if err != nil {
		logp.Err("%v", err)
		return
	}

	defer func() {
		logp.Info("stopping UDP listener on %s", uc.LocalAddr())
		uc.Close()
	}()

	for {
		if atomic.LoadUint32(&hepInp.stopped) == 1 {
			return
		}
		uc.SetReadDeadline(time.Now().Add(1e9))
		buf := hepInp.buffer.Get().([]byte)
		n, err := uc.Read(buf)
		if err != nil {
			if opErr, ok := err.(*net.OpError); ok && opErr.Timeout() {
				continue
			} else {
				logp.Err("%v", err)
				return
			}
		} else if n > maxPktLen {
			logp.Warn("received too big packet with %d bytes", n)
			atomic.AddUint64(&hepInp.stats.ErrCount, 1)
			continue
		}
		hepInp.inputCh <- buf[:n]
		atomic.AddUint64(&hepInp.stats.PktCount, 1)
	}
}
