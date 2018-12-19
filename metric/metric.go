package metric

import (
	"sync"

	"github.com/sipcapture/heplify-server"
)

type Metric struct {
	MH   MetricHandler
	Chan chan *decoder.HEP
}

type MetricHandler interface {
	setup() error
	collect(chan *decoder.HEP)
}

func New(name string) *Metric {
	var register = map[string]MetricHandler{
		"prometheus": new(Prometheus),
	}

	return &Metric{
		MH: register[name],
	}
}

func (m *Metric) Run() error {
	var (
		wg  sync.WaitGroup
		err error
	)

	err = m.MH.setup()
	if err != nil {
		return err
	}

	wg.Add(1)
	go func() {
		defer wg.Done()
		m.MH.collect(m.Chan)
	}()
	wg.Wait()
	return nil

}

func (m *Metric) End() {
	close(m.Chan)
}
