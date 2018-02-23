package queue

import (
	"fmt"
	"time"

	"github.com/negbie/heplify-server"
	"github.com/negbie/heplify-server/config"
	"github.com/negbie/heplify-server/logp"
	nsq "github.com/nsqio/go-nsq"
)

type NSQ struct {
	producer *nsq.Producer
}

func (n *NSQ) setup() error {
	var (
		err error
		cfg = nsq.NewConfig()
	)

	cfg.UserAgent = fmt.Sprintf("heplify-server-nsq-%s", nsq.VERSION)
	cfg.DialTimeout = time.Millisecond * time.Duration(2000)

	n.producer, err = nsq.NewProducer(config.Setting.NSQAddr, cfg)
	if err != nil {
		logp.Err("%v", err)
		return err
	}
	return nil
}

func (n *NSQ) add(topic string, mCh chan *decoder.HEPPacket, ec *uint64) {
	var (
		msg *decoder.HEPPacket
		err error
		ok  bool
	)

	logp.Info("Run NSQ Output, server: %+v, topic: %s\n", config.Setting.NSQAddr, topic)

	for {
		msg, ok = <-mCh
		if !ok {
			break
		}

		err = n.producer.Publish(topic, []byte(msg.Payload))
		if err != nil {
			logp.Err("%v", err)
			*ec++
		}
	}
}
