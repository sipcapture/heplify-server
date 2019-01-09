package input

import (
	"bufio"
	"encoding/binary"
	"net"
	"sync/atomic"
	"time"

	"github.com/negbie/heplify-server/config"
	"github.com/negbie/logp"
)

func (h *HEPInput) serveTCP() {
	listener, err := net.Listen("tcp", config.Setting.HEPTCPAddr)
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
		logp.Info("new TCP connection %s -> %s", conn.RemoteAddr(), conn.LocalAddr())
		if err != nil {
			logp.Err("%v", err)
			continue
		}
		h.wg.Add(1)
		go h.handleTCP(conn)
	}
}

func (h *HEPInput) handleTCP(c net.Conn) {
	r := bufio.NewReader(c)
	defer c.Close()
	defer h.wg.Done()
	readBytes := func(buffer []byte) (int, error) {
		n := uint(0)
		for n < uint(len(buffer)) {
			nn, err := r.Read(buffer[n:])
			n += uint(nn)
			if err != nil {
				return 0, err
			}
		}
		return int(n), nil
	}
	for {
		select {
		case <-h.quit:
			return
		default:
		}

		c.SetReadDeadline(time.Now().Add(1e9))
		buf := h.buffer.Get().([]byte)
		hb, err := r.Peek(6)
		if err != nil {
			if opErr, ok := err.(*net.OpError); ok && opErr.Timeout() {
				continue
			} else {
				return
			}
		} else {
			size := binary.BigEndian.Uint16(hb[4:6])
			if size > maxPktLen {
				logp.Warn("unexpected packet, did you send TLS into plain TCP input?")
				return
			}
			n, err := readBytes(buf[:size])
			if err != nil || n > maxPktLen {
				logp.Warn("%v, unusal packet size with %d bytes", err, n)
				atomic.AddUint64(&h.stats.ErrCount, 1)
				continue
			}
			h.inCh <- buf[:n]
			atomic.AddUint64(&h.stats.PktCount, 1)
		}
	}
}
