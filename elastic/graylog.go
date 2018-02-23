package elastic

import (
	"github.com/negbie/heplify-server"
)

type Graylog struct {
}

func (g *Graylog) setup() error {
	var err error
	return err
}

func (g *Graylog) send(mCh chan *decoder.HEPPacket) {
	var (
		pkt *decoder.HEPPacket
		ok  bool
	)

	for {
		pkt, ok = <-mCh
		if !ok {
			break
		}

	}
}
