package input

import (
	"bytes"
	"crypto/tls"
	"net"

	"github.com/negbie/heplify-server/config"
	"github.com/negbie/logp"
)

func pipeConn(localConn net.Conn, remoteAddr string) {
	rAddr, err := net.ResolveTCPAddr("tcp", remoteAddr)
	if err != nil {
		logp.Err("%v", err)
	}

	remoteConn, err := tls.Dial("tcp", rAddr.String(), &tls.Config{InsecureSkipVerify: true})
	if err != nil {
		logp.Err("%v", err)
		return
	}
	logp.Info("pipe connection to %s established\n", rAddr.String())

	defer func() {
		logp.Info("close pipe connection")
		err := remoteConn.Close()
		if err != nil {
			logp.Err("%v", err)
		}
	}()

	pipe(localConn, remoteConn)
}

func chanFromConn(conn net.Conn) chan []byte {
	c := make(chan []byte)

	go func() {
		tmpBuf := make([]byte, 2048)

		for {
			n, err := conn.Read(tmpBuf)
			if n > 0 {
				if len(config.Setting.PipeFilter) > 0 {
					for k := range config.Setting.PipeFilter {
						if !bytes.Contains(tmpBuf, []byte(config.Setting.PipeFilter[k])) {
							continue
						}
					}
				}
				buf := make([]byte, n)
				copy(buf, tmpBuf[:n])
				c <- buf
			}
			if err != nil {
				if opErr, ok := err.(*net.OpError); ok && opErr.Timeout() {
					continue
				}
				logp.Err("%v", err)
				c <- nil
				break
			}
		}
	}()

	return c
}

func pipe(conn1 net.Conn, conn2 net.Conn) {
	chan1 := chanFromConn(conn1)
	chan2 := chanFromConn(conn2)
	for {
		select {
		case b1 := <-chan1:
			if b1 == nil {
				return
			}
			conn2.Write(b1)
		case b2 := <-chan2:
			if b2 == nil {
				return
			}
			conn1.Write(b2)
		}
	}
}
