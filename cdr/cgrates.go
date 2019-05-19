package cdr

import (
	"encoding/binary"
	"fmt"

	"github.com/VictoriaMetrics/fastcache"
	"github.com/negbie/logp"
	"github.com/sipcapture/heplify-server/config"
	"github.com/sipcapture/heplify-server/decoder"
)

const (
	invite    = "INVITE"
	bye       = "BYE"
	cacheSize = 20 * 1024 * 1024
)

type CGR struct {
	cache *fastcache.Cache
}

func (c *CGR) setup() error {
	c.cache = fastcache.New(cacheSize)
	return nil
}

func (c *CGR) send(hCh chan *decoder.HEP) {

	logp.Info("Run CGRateS Output, server: %s\n", config.Setting.CGRAddr)

	for {
		pkt, ok := <-hCh
		if !ok {
			break
		}

		k := []byte(pkt.CID)
		if pkt.SIP.CseqMethod == invite {
			if !c.cache.Has(k) {
				tu := uint64(pkt.Timestamp.UnixNano())
				tb := make([]byte, 8)
				binary.BigEndian.PutUint64(tb, tu)
				c.cache.Set(k, tb)
			}
		}

		if pkt.SIP.CseqMethod == bye {
			if buf := c.cache.Get(nil, k); buf != nil {
				d := uint64(pkt.Timestamp.UnixNano()) - binary.BigEndian.Uint64(buf)
				fmt.Println(pkt.CID, d/1e9)
				c.cache.Del(k)
			}
		}
	}
}
