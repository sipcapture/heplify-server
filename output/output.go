package output

import (
	"sync"
)

type Output struct {
	Fan      Out
	ConfFile string
	ErrCount *uint64

	Topic string
	Chan  chan []byte
}

type Out interface {
	set(string) error
	add(string, chan []byte, *uint64)
}

func New(name string) *Output {
	var register = map[string]Out{
		"nsq": new(NSQ),
	}

	return &Output{
		Fan: register[name],
	}
}

func (o *Output) Run() error {
	var (
		wg  sync.WaitGroup
		err error
	)

	err = o.Fan.set(o.ConfFile)
	if err != nil {
		return err
	}

	wg.Add(1)
	go func() {
		defer wg.Done()
		topic := o.Topic
		o.Fan.add(topic, o.Chan, o.ErrCount)
	}()

	wg.Wait()

	return nil
}

func (o *Output) End() {
	close(o.Chan)
}
