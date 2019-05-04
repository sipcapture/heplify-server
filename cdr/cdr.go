package cdr

import (
	"github.com/negbie/logp"
	"github.com/sipcapture/heplify-server/decoder"
)

type CDR struct {
	H    CDRHandler
	Chan chan *decoder.HEP
}

type CDRHandler interface {
	setup() error
	send(chan *decoder.HEP)
}

func New(name string) *CDR {
	var register = map[string]CDRHandler{
		"cgrates": new(CGR),
	}

	return &CDR{
		H: register[name],
	}
}

func (q *CDR) Run() error {

	err := q.H.setup()
	if err != nil {
		return err
	}

	go func() {
		q.H.send(q.Chan)
	}()

	return nil
}

func (q *CDR) End() {
	close(q.Chan)
	logp.Info("close CDR channel")
}
