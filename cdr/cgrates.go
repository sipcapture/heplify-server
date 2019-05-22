package cdr

import (
	"crypto/rand"
	"crypto/sha1"
	"encoding/binary"
	"fmt"
	"io"
	"log"
	"time"

	"github.com/VictoriaMetrics/fastcache"
	"github.com/cgrates/rpcclient"
	"github.com/negbie/logp"
	"github.com/sipcapture/heplify-server/config"
	"github.com/sipcapture/heplify-server/decoder"
)

const (
	cacheSize = 20 * 1024 * 1024
)

var (
	attributeS = false
	ralS       = true
	chargerS   = true
	store      = false
	export     = false
	thresholdS = false
	statS      = true
)

type CGR struct {
	cache  *fastcache.Cache
	client *rpcclient.RpcClient
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
	c.client, err = rpcclient.NewRpcClient("tcp", "localhost:2013", false, "",
		"", "", 3, 3,
		time.Duration(1*time.Second), time.Duration(5*time.Minute), "gob", nil, false)
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
						Store:      &store,
						Export:     &export,
						ThresholdS: &thresholdS,
						StatS:      &statS,
						CGREvent: CGREvent{
							Tenant: "cgrates.org",
							ID:     UUIDSha1Prefix(),
							Time:   &pkt.Timestamp,
							Event: map[string]interface{}{
								"ToR":         "*voice",
								"OriginID":    pkt.CID,
								"Account":     pkt.SIP.FromUser,
								"Subject":     pkt.SIP.FromUser,
								"Destination": pkt.SIP.ToUser,
								"Category":    "call",
								"Source":      pkt.SIP.FromHost,
								"RequestType": "*postpaid",
								"AnswerTime":  pkt.Timestamp.Add(-1 * d),
								"Usage":       d,
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

// helper function for uuid generation
func GenUUID() string {
	b := make([]byte, 16)
	_, err := io.ReadFull(rand.Reader, b)
	if err != nil {
		log.Fatal(err)
	}
	b[6] = (b[6] & 0x0F) | 0x40
	b[8] = (b[8] &^ 0x40) | 0x80
	return fmt.Sprintf("%x-%x-%x-%x-%x", b[:4], b[4:6], b[6:8], b[8:10], b[10:])
}

// UUIDSha1Prefix generates a prefix of the sha1 applied to an UUID
// prefix 8 is chosen since the probability of colision starts being minimal after 7 characters (see git commits)
func UUIDSha1Prefix() string {
	return Sha1(GenUUID())[:7]
}

func Sha1(attrs ...string) string {
	hasher := sha1.New()
	for _, attr := range attrs {
		hasher.Write([]byte(attr))
	}
	return fmt.Sprintf("%x", hasher.Sum(nil))
}
