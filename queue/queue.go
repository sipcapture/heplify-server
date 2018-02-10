package queue

import (
	"sync"

	"github.com/negbie/heplify-server"
)

type MessageQueue struct {
	Queue    QueueHandler
	ErrCount *uint64

	Topic string
	Chan  chan *decoder.HEPPacket
}

type QueueHandler interface {
	setup() error
	add(string, chan *decoder.HEPPacket, *uint64)
}

func New(name string) *MessageQueue {
	var register = map[string]QueueHandler{
		"nsq": new(NSQ),
	}

	return &MessageQueue{
		Queue: register[name],
	}
}

func (m *MessageQueue) Run() error {
	var (
		wg  sync.WaitGroup
		err error
	)

	err = m.Queue.setup()
	if err != nil {
		return err
	}

	wg.Add(1)
	go func() {
		defer wg.Done()
		topic := m.Topic
		m.Queue.add(topic, m.Chan, m.ErrCount)
	}()

	wg.Wait()
	return nil
}

func (m *MessageQueue) End() {
	close(m.Chan)
}
