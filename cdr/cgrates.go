package cdr

import (
	"github.com/negbie/logp"
	"github.com/sipcapture/heplify-server/config"
	"github.com/sipcapture/heplify-server/decoder"
)

type CGR struct {
}

func (c *CGR) setup() error {

	return nil
}

func (c *CGR) send(hCh chan *decoder.HEP) {

	logp.Info("Run CGRateS Output, server: %s\n", config.Setting.CGRAddr)

	for {
		msg, ok := <-hCh
		if !ok {
			break
		}
		_ = msg
		_ = ok

	}
}
