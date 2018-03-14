package database

import (
	"github.com/gobuffalo/packr"
	"github.com/negbie/heplify-server/database/migrate"
)

// TODO: CRON
func rotate(addr []string, box packr.Box) (err error) {
	if err := migrate.CreateDataTables(addr, box); err != nil {
		return err
	}

	if err := migrate.DropTables(addr, box); err != nil {
		return err
	}
	return nil
}
