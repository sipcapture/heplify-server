package database

import (
	"sync"

	"github.com/negbie/heplify-server"
)

type Database struct {
	DB       DBHandler
	ErrCount *uint64

	Topic string
	Chan  chan *decoder.HEPPacket
}

type DBHandler interface {
	setup() error
	insert(string, chan *decoder.HEPPacket, *uint64)
}

func New(name string) *Database {
	var register = map[string]DBHandler{
		"mysql": new(MySQL),
	}

	return &Database{
		DB: register[name],
	}
}

func (d *Database) Run() error {
	var (
		wg  sync.WaitGroup
		err error
	)

	err = d.DB.setup()
	if err != nil {
		return err
	}

	wg.Add(1)
	go func() {
		defer wg.Done()
		topic := d.Topic
		d.DB.insert(topic, d.Chan, d.ErrCount)
	}()

	wg.Wait()
	return nil
}

func (d *Database) End() {
	close(d.Chan)
}
