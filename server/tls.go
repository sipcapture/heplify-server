package input

import (
	"crypto/tls"
	"net"
	"sync/atomic"
	"time"

	"github.com/negbie/heplify-server/config"
	"github.com/negbie/logp"
)

func (h *HEPInput) serveTLS() {
	ca := NewCertificateAuthority()
	listener, err := tls.Listen("tcp", config.Setting.HEPTLSAddr, &tls.Config{
		GetCertificate: ca.GetCertificate,
	})
	if err != nil {
		logp.Err("%v", err)
		return
	}

	for {
		select {
		case <-h.quit:
			listener.Close()
			return
		default:
		}

		conn, err := listener.Accept()
		if err != nil {
			logp.Err("%v", err)
			continue
		}
		h.wg.Add(1)
		go h.handleTLS(conn)
	}
}

func (h *HEPInput) handleTLS(c net.Conn) {
	defer c.Close()
	defer h.wg.Done()
	for {
		select {
		case <-h.quit:
			return
		default:
		}

		c.SetReadDeadline(time.Now().Add(1e9))
		buf := h.buffer.Get().([]byte)
		n, err := c.Read(buf)
		if err != nil {
			if opErr, ok := err.(*net.OpError); ok && opErr.Timeout() {
				continue
			} else {
				return
			}
		} else if n > maxPktLen {
			logp.Warn("received too big packet with %d bytes", n)
			atomic.AddUint64(&h.stats.ErrCount, 1)
			continue
		}
		h.inCh <- buf[:n]
		atomic.AddUint64(&h.stats.PktCount, 1)
	}
}
