package queue

import (
	"fmt"

	"github.com/sipcapture/heplify-server/config"
	"github.com/negbie/logp"
)

type Queue struct {
	H     QueueHandler
	Topic string
	Chan  chan []byte
}

type QueueHandler interface {
	setup() error
	add(string, chan []byte)
}

func New(name string) *Queue {
	var register = map[string]QueueHandler{
		"nsq":  new(NSQ),
		"nats": new(NATS),
	}

	return &Queue{
		H: register[name],
	}
}

func (q *Queue) Run() error {

	if config.Setting.MQDriver != "nats" && config.Setting.MQDriver != "nsq" {
		return fmt.Errorf("invalid message queue driver: %s, please use nats or nsq", config.Setting.MQDriver)
	}

	err := q.H.setup()
	if err != nil {
		return err
	}

	go func() {
		topic := q.Topic
		q.H.add(topic, q.Chan)
	}()

	return nil
}

func (q *Queue) End() {
	close(q.Chan)
	logp.Info("close queue channel")
}
