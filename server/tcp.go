package input

import (
	"bufio"
	"encoding/binary"
	"net"
	"sync"
	"sync/atomic"
	"time"

	"github.com/negbie/logp"
)

func (h *HEPInput) serveTCP(addr string) {
	defer close(h.exitTCP)

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
			logp.Info("stopping TCP listener on %s", ln.Addr())
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
				logp.Err("failed to accept TCP connection: %v", err.Error())
			}
			continue
		}
		logp.Info("new TCP connection %s -> %s", conn.RemoteAddr(), conn.LocalAddr())
		wg.Add(1)
		go func() {
			h.handleTCP(conn)
			wg.Done()
		}()
	}
}

func (h *HEPInput) handleTCP(c net.Conn) {
	defer func() {
		logp.Info("closing TCP connection from %s", c.RemoteAddr())
		err := c.Close()
		if err != nil {
			logp.Err("%v", err)
		}
	}()

	r := bufio.NewReader(c)
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
		if atomic.LoadUint32(&h.stopped) == 1 {
			return
		}

		hb, err := r.Peek(6)
		if err != nil {
			logp.Warn("%v from %s", err, c.RemoteAddr())
			return
		} else {
			size := binary.BigEndian.Uint16(hb[4:6])
			if size > maxPktLen {
				logp.Warn("wrong or too big HEP packet size with %d bytes", size)
				//r.Reset(c)
				//continue
				return
			}
			buf := h.buffer.Get().([]byte)
			n, err := readBytes(buf[:size])
			if err != nil || n > maxPktLen {
				logp.Warn("%v, unusal packet size with %d bytes", err, n)
				atomic.AddUint64(&h.stats.ErrCount, 1)
				continue
			}
			h.inputCh <- buf[:n]
			atomic.AddUint64(&h.stats.PktCount, 1)
		}
	}
}
