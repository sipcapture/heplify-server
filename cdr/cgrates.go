package cdr

import (
	"encoding/binary"
	"net/rpc"
	"net/rpc/jsonrpc"
	"time"

	"github.com/VictoriaMetrics/fastcache"
	"github.com/negbie/logp"
	"github.com/sipcapture/heplify-server/config"
	"github.com/sipcapture/heplify-server/decoder"
)

const (
	cacheSize = 20 * 1024 * 1024
)

var (
	attributeS = true
	ralS       = true
	chargerS   = true
	store      = true
	export     = true
	thresholdS = true
	statS      = true
)

type CGR struct {
	cache  *fastcache.Cache
	client *rpc.Client
}

type ArgV1ProcessEvent struct {
	CGREvent
	AttributeS *bool // control AttributeS processing
	RALs       *bool // control if we rate the event
	ChargerS   *bool // control ChargerS processing
	Store      *bool // control storing of the CDR
	Export     *bool // control online exports for the CDR
	ThresholdS *bool // control ThresholdS
	StatS      *bool // control sending the CDR to StatS for aggregation
	*ArgDispatcher
}

type CGREvent struct {
	Tenant string
	ID     string
	Time   *time.Time // event time
	Event  map[string]interface{}
}

type ArgDispatcher struct {
	APIKey  *string
	RouteID *string
}

func (c *CGR) setup() error {
	var err error
	c.client, err = jsonrpc.Dial("tcp", "localhost:2012")
	if err != nil {
		logp.Warn("%v", err)
		return err
	}

	c.cache = fastcache.New(cacheSize)
	logp.Info("Run CGRateS Output, server: %s\n", config.Setting.CGRAddr)
	return nil
}

func (c *CGR) send(hCh chan *decoder.HEP) {
	for {
		pkt, ok := <-hCh
		if !ok {
			break
		}

		k := []byte(pkt.CID)
		if pkt.SIP.CseqMethod == "INVITE" {
			if !c.cache.Has(k) {
				tu := uint64(pkt.Timestamp.UnixNano())
				tb := make([]byte, 8)
				binary.BigEndian.PutUint64(tb, tu)
				c.cache.Set(k, tb)
			}
		}

		if pkt.SIP.CseqMethod == "BYE" {
			if buf := c.cache.Get(nil, k); buf != nil {
				d := time.Duration(uint64(pkt.Timestamp.UnixNano()) - binary.BigEndian.Uint64(buf))
				if d < 1e15 {
					args := &ArgV1ProcessEvent{
						AttributeS: &attributeS,
						RALs:       &ralS,
						ChargerS:   &chargerS,
						// Store:      &store,
						// Export:     &export,
						ThresholdS: &thresholdS,
						StatS:      &statS,
						CGREvent: CGREvent{
							Tenant: "cgrates.org",
							Event: map[string]interface{}{
								//"EventName":   "TEST_EVENT",
								//"ToR":         "*voice",
								"OriginID":    "123452",
								"Account":     "1001",
								"Subject":     "1001",
								"Destination": "1003",
								"Category":    "call",
								"Tenant":      "cgrates.org",
								"Source":      "192.168.1.1",
								"RequestType": "*prepaid",
								//"SetupTime":   pkt.Timestamp.Add(-1 * d),
								"AnswerTime": pkt.Timestamp,
								"Usage":      d,
							},
						},
					}

					var reply string
					if err := c.client.Call("CDRsV1.ProcessEvent", args, &reply); err != nil {
						logp.Warn("%v", err)
					}
					logp.Info(reply)
				}
				c.cache.Del(k)
			}
		}
	}
}
