package remotelog

import (
	"github.com/negbie/logp"
	"github.com/sipcapture/heplify-server/decoder"
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
	err := r.H.setup()
	if err != nil {
		return err
	}

	go func() {
		r.H.start(r.Chan)
	}()

	return nil
}

func (r *Remotelog) End() {
	close(r.Chan)
	logp.Info("close remotelog channel")
}
