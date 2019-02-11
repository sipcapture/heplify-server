package remotelog

import (
	"sync"

	"github.com/negbie/heplify-server/decoder"
)

type Remotelog struct {
	H    RemoteHandler
	Chan chan *decoder.HEP
}

type RemoteHandler interface {
	setup() error
	start(chan *decoder.HEP)
}

func New(name string) *Remotelog {
	var register = map[string]RemoteHandler{
		"elasticsearch": new(Elasticsearch),
		"loki":          new(Loki),
	}

	return &Remotelog{
		H: register[name],
	}
}

func (r *Remotelog) Run() error {
	var (
		wg  sync.WaitGroup
		err error
	)

	err = r.H.setup()
	if err != nil {
		return err
	}

	wg.Add(1)
	go func() {
		defer wg.Done()
		r.H.start(r.Chan)
	}()
	wg.Wait()
	return nil
}

func (r *Remotelog) End() {
	close(r.Chan)
}
