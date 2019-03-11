package input

import (
	"crypto/tls"
	"net"
	"sync/atomic"
	"time"

	"github.com/negbie/logp"
)

func (h *HEPInput) serveTLS(addr string) {
	ta, err := net.ResolveTCPAddr("tcp", addr)
	if err != nil {
		logp.Err("%v", err)
		return
	}

	ln, err := net.ListenTCP("tcp", ta)
	if err != nil {
		logp.Err("%v", err)
		return
	}
	defer ln.Close()
	ca := NewCertificateAuthority()

	for {
		ln.SetDeadline(time.Now().Add(1e9))
		conn, err := ln.Accept()
		if err != nil {
			select {
			case <-h.quitTLS:
				logp.Info("stopping TLS listener on %s", ln.Addr())
				h.quitTLS <- true
				return
			default:
			}
			continue
		}
		go h.handleTLS(tls.Server(conn, &tls.Config{GetCertificate: ca.GetCertificate}))
	}
}

func (h *HEPInput) handleTLS(c net.Conn) {
	defer func() {
		logp.Info("closing TLS connection from %s", c.RemoteAddr())
		defer c.Close()
	}()

	for {
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
