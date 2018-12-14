package remotelog

import (
	"sync"

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
		"elasticsearch": new(Elasticsearch),
		"loki":          new(Loki),
	}

	return &Elastic{
		EH: register[name],
	}
}

func (e *Elastic) Run() error {
	var (
		wg  sync.WaitGroup
		err error
	)

	err = e.EH.setup()
	if err != nil {
		return err
	}

	wg.Add(1)
	go func() {
		defer wg.Done()
		e.EH.send(e.Chan)
	}()
	wg.Wait()
	return nil
}

func (e *Elastic) End() {
	close(e.Chan)
}
