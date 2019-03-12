package queue

import (
	nats "github.com/nats-io/go-nats"
	"github.com/sipcapture/heplify-server/config"
	"github.com/negbie/logp"
)

// NATS producer struct
type NATS struct {
	connection *nats.Conn
}

func (n *NATS) setup() error {
	var err error

	n.connection, err = nats.Connect("nats://" + config.Setting.MQAddr)
	if err != nil {
		logp.Err("%v", err)
		return err
	}

	return nil
}

func (n *NATS) add(topic string, qCh chan []byte) {
	var (
		msg []byte
		err error
		ok  bool
	)

	logp.Info("Run NATS Output, server: %s, topic: %s\n", config.Setting.MQAddr, topic)

	for {
		msg, ok = <-qCh
		if !ok {
			break
		}

		err = n.connection.Publish(topic, msg)
		if err != nil {
			logp.Err("%v", err)
		}
	}
}
