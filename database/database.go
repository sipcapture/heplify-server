package database

import (
	"fmt"
	"runtime"

	"github.com/negbie/heplify-server"
	"github.com/negbie/heplify-server/config"
)

type Database struct {
	DBH  DBHandler
	Chan chan *decoder.HEP
}

type DBHandler interface {
	setup() error
	insert(chan *decoder.HEP)
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

	if config.Setting.DBDriver != "mysql" && config.Setting.DBDriver != "postgres" {
		return fmt.Errorf("wrong database driver: %s, please use mysql or postgres", config.Setting.DBDriver)
	}

	err = d.DBH.setup()
	if err != nil {
		return err
	}

	for i := 0; i < runtime.NumCPU(); i++ {
		go func() {
			d.DBH.insert(d.Chan)
		}()
	}

	return nil
}

func (d *Database) End() {
	close(d.Chan)
}
