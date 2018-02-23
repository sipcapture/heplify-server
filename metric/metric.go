package metric

import (
	"runtime"

	"github.com/negbie/heplify-server"
)

type Metric struct {
	MH   MetricHandler
	Chan chan *decoder.HEPPacket
}

type MetricHandler interface {
	setup() error
	collect(chan *decoder.HEPPacket)
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
		//wg  sync.WaitGroup
		err error
	)

	err = m.MH.setup()
	if err != nil {
		return err
	}

	for i := 0; i < runtime.NumCPU(); i++ {
		go func() {
			m.MH.collect(m.Chan)
		}()
	}

	return nil
}

func (m *Metric) End() {
	close(m.Chan)
}
