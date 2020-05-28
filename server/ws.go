package input

import (
	"github.com/gobwas/ws"
	"io"
	"net"
	"sync"
	"sync/atomic"
	"time"

	"github.com/negbie/logp"
)

func (h *HEPInput) serveWS(addr string) {
	defer close(h.exitWS)

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

	var wg sync.WaitGroup

	for {
		if atomic.LoadUint32(&h.stopped) == 1 {
			logp.Info("stopping WS listener on %s", ln.Addr())
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
				logp.Err("failed to accept WS connection: %v", err.Error())
			}
			continue
		}
		_, err = ws.Upgrade(conn)
		if err != nil {
			logp.Err("%v", err)
			return
		}

		logp.Info("new WS connection %s -> %s", conn.RemoteAddr(), conn.LocalAddr())
		wg.Add(1)
		go func() {
			h.handleWS(conn)
			wg.Done()
		}()
	}
}

func (h *HEPInput) handleWS(c net.Conn) {
	defer func() {
		logp.Info("closing WS connection from %s", c.RemoteAddr())
		err := c.Close()
		if err != nil {
			logp.Err("%v", err)
		}
	}()
	for {
		header, err := ws.ReadHeader(c)
		if err != nil {
			logp.Err("%v", err)
			return
		}
		buf := h.buffer.Get().([]byte)
		payload := buf[:header.Length]
		_, err = io.ReadFull(c, payload)
		if err != nil {
			logp.Err("%v", err)
			return
		}

		if header.Masked {
			ws.Cipher(payload, header.Mask, 0)
		}

		if header.OpCode == ws.OpClose {
			return
		}
		h.inputCh <- payload
		atomic.AddUint64(&h.stats.PktCount, 1)
	}
}
