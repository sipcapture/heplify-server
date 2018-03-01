package database

import (
	"runtime"

	"github.com/negbie/heplify-server"
)

type Database struct {
	DBH      DBHandler
	ErrCount *uint64

	Topic string
	Chan  chan *decoder.HEP
}

type DBHandler interface {
	setup() error
	insert(string, chan *decoder.HEP, *uint64)
}

func New(name string) *Database {
	var register = map[string]DBHandler{
		"mysql":    new(SQL),
		"postgres": new(SQL),
	}

	return &Database{
		DBH: register[name],
	}
}

func (d *Database) Run() error {
	var (
		//wg  sync.WaitGroup
		err error
	)

	err = d.DBH.setup()
	if err != nil {
		return err
	}

	for i := 0; i < runtime.NumCPU(); i++ {
		go func() {
			topic := d.Topic
			d.DBH.insert(topic, d.Chan, d.ErrCount)
		}()
	}

	return nil
}

func (d *Database) End() {
	close(d.Chan)
}
