package queue

import (
	"fmt"
	"sync"

	"github.com/negbie/heplify-server/config"
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
	var (
		wg  sync.WaitGroup
		err error
	)

	if config.Setting.MQDriver != "nats" && config.Setting.MQDriver != "nsq" {
		return fmt.Errorf("invalid message queue driver: %s, please use nats or nsq", config.Setting.MQDriver)
	}

	err = q.H.setup()
	if err != nil {
		return err
	}

	wg.Add(1)
	go func() {
		defer wg.Done()
		topic := q.Topic
		q.H.add(topic, q.Chan)
	}()
	wg.Wait()
	return nil
}

func (q *Queue) End() {
	close(q.Chan)
}
