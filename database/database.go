package database

import (
	"runtime"

	"github.com/negbie/heplify-server"
	"github.com/negbie/heplify-server/config"
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
		config.Setting.DBDriver: new(SQL),
	}

	return &Database{
		DB: register[name],
	}
}

func (d *Database) Run() error {
	var (
		//wg  sync.WaitGroup
		err error
	)

	err = d.DB.setup()
	if err != nil {
		return err
	}

	for i := 0; i < runtime.NumCPU(); i++ {
		go func() {
			topic := d.Topic
			d.DB.insert(topic, d.Chan, d.ErrCount)
		}()
	}

	return nil
}

func (d *Database) End() {
	close(d.Chan)
}
