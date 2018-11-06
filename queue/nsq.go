package queue

import (
	"fmt"
	"time"

	"github.com/sipcapture/heplify-server/config"
	"github.com/sipcapture/logp"
	nsq "github.com/nsqio/go-nsq"
)

// NSQ producer struct
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

	n.producer, err = nsq.NewProducer(config.Setting.MQAddr, cfg)
	if err != nil {
		logp.Err("%v", err)
		return err
	}
	return nil
}

func (n *NSQ) add(topic string, qCh chan []byte) {
	var (
		msg []byte
		err error
		ok  bool
	)

	logp.Info("Run NSQ Output, server: %s, topic: %s\n", config.Setting.MQAddr, topic)

	for {
		msg, ok = <-qCh
		if !ok {
			break
		}

		err = n.producer.Publish(topic, msg)
		if err != nil {
			logp.Err("%v", err)
		}
	}
}
