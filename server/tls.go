package input

import (
	"crypto/tls"
	"net"
	"sync"
	"sync/atomic"
	"time"
	"path/filepath"

	"github.com/sipcapture/heplify-server/config"
	"github.com/negbie/cert"
	"github.com/negbie/logp"
)

func (h *HEPInput) serveTLS(addr string) {
	defer close(h.exitTLS)

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

	// get path for certificate/key storage
	cPath := config.Setting.TLSCertFolder
	// load any existing certs, otherwise generate a new one
	ca, err := cert.NewCertificateAuthority( filepath.Join(cPath,  "heplify-server") )
	if err != nil {
		logp.Err("%v", err)
		return
	}

	var wg sync.WaitGroup

	for {
		if atomic.LoadUint32(&h.stopped) == 1 {
			logp.Info("stopping TLS listener on %s", ln.Addr())
			ln.Close()
			wg.Wait()
			return
		}

		if err := ln.SetDeadline(time.Now().Add(1e9)); err != nil {
			logp.Err("%v", err)
			break
		}

		conn, err := ln.Accept()
		if err != nil {
			if opErr, ok := err.(*net.OpError); !ok || !opErr.Timeout() {
				logp.Err("failed to accept TLS connection: %v", err.Error())
			}
			continue
		}
		logp.Info("new TLS connection %s -> %s", conn.RemoteAddr(), conn.LocalAddr())
		wg.Add(1)
		go func() {
			h.handleTLS(tls.Server(conn, &tls.Config{GetCertificate: ca.GetCertificate}))
			wg.Done()
		}()
	}
}

func (h *HEPInput) handleTLS(c net.Conn) {
	defer func() {
		logp.Info("closing TLS connection from %s", c.RemoteAddr())
		err := c.Close()
		if err != nil {
			logp.Err("%v", err)
		}
	}()

	for {
		if atomic.LoadUint32(&h.stopped) == 1 {
			return
		}

		buf := h.buffer.Get().([]byte)
		n, err := c.Read(buf)
		if err != nil {
			logp.Warn("%v from %s", err, c.RemoteAddr())
			return
		} else if n > maxPktLen {
			logp.Warn("received too big packet with %d bytes", n)
			atomic.AddUint64(&h.stats.ErrCount, 1)
			continue
		}
		h.inputCh <- buf[:n]
		atomic.AddUint64(&h.stats.PktCount, 1)
	}
}
