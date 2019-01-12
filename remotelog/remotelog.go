package remotelog

import (
	"sync"

	"github.com/negbie/heplify-server/decoder"
)

type Remotelog struct {
	EH   ElasticHandler
	Chan chan *decoder.HEP
}

type ElasticHandler interface {
	setup() error
	send(chan *decoder.HEP)
}

func New(name string) *Remotelog {
	var register = map[string]ElasticHandler{
		"elasticsearch": new(Elasticsearch),
		"loki":          new(Loki),
	}

	return &Remotelog{
		EH: register[name],
	}
}

func (r *Remotelog) Run() error {
	var (
		wg  sync.WaitGroup
		err error
	)

	err = r.EH.setup()
	if err != nil {
		return err
	}

	wg.Add(1)
	go func() {
		defer wg.Done()
		r.EH.send(r.Chan)
	}()
	wg.Wait()
	return nil
}

func (r *Remotelog) End() {
	close(r.Chan)
}
