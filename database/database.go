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
	if config.Setting.Mode == "homer5" {
		name += "Homer5"
	} else if config.Setting.Mode == "homer7" {
		name += "Homer7"
	}
	var register = map[string]DBHandler{
		"mysqlHomer5":    new(SQLHomer5),
		"postgresHomer5": new(SQLHomer5),
		"mysqlHomer7":    new(SQLHomer7),
		"postgresHomer7": new(SQLHomer7),
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
	if config.Setting.Mode != "homer5" && config.Setting.Mode != "homer7" {
		return fmt.Errorf("wrong mode: %s, please use homer5 or homer7", config.Setting.Mode)
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
