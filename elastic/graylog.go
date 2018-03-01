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

func (g *Graylog) send(hCh chan *decoder.HEP) {
	var (
		pkt *decoder.HEP
		ok  bool
	)

	for {
		pkt, ok = <-hCh
		if !ok {
			break
		}

	}
}
