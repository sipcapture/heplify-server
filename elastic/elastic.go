package elastic

import (
	"runtime"

	"github.com/negbie/heplify-server"
)

type Elastic struct {
	EH   ElasticHandler
	Chan chan *decoder.HEP
}

type ElasticHandler interface {
	setup() error
	send(chan *decoder.HEP)
}

func New(name string) *Elastic {
	var register = map[string]ElasticHandler{
		"graylog": new(Graylog),
	}

	return &Elastic{
		EH: register[name],
	}
}

func (e *Elastic) Run() error {
	var (
		//wg  sync.WaitGroup
		err error
	)

	err = e.EH.setup()
	if err != nil {
		return err
	}

	for i := 0; i < runtime.NumCPU(); i++ {
		go func() {
			e.EH.send(e.Chan)
		}()
	}

	return nil
}

func (e *Elastic) End() {
	close(e.Chan)
}
