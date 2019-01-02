package metric

import (
	"encoding/binary"
	"runtime"
	"time"

	"github.com/negbie/heplify-server"
)

type Metric struct {
	MH   MetricHandler
	Chan chan *decoder.HEP
}

type MetricHandler interface {
	setup() error
	expose(chan *decoder.HEP)
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
	err := m.MH.setup()
	if err != nil {
		return err
	}

	for i := 0; i < runtime.NumCPU(); i++ {
		go func() {
			m.MH.expose(m.Chan)
		}()
	}
	return nil
}

func (m *Metric) End() {
	close(m.Chan)
}

func decTimeByte(b []byte) time.Time {
	i := int64(binary.BigEndian.Uint64(b))
	return time.Unix(0, i)
}

func encTimeByte(t time.Time) []byte {
	buf := make([]byte, 8)
	u := uint64(t.UnixNano())
	binary.BigEndian.PutUint64(buf, u)
	return buf
}
