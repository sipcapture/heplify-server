package database

import (
	"fmt"
	"net/url"
	"strings"

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
	if config.Setting.DBShema == "homer5" {
		name += "Homer5"
	} else if config.Setting.DBShema == "homer7" {
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
	if config.Setting.DBDriver != "mysql" && config.Setting.DBDriver != "postgres" {
		return fmt.Errorf("Invalid DBDriver: %s, please use mysql or postgres", config.Setting.DBDriver)
	}
	if config.Setting.DBShema != "homer5" && config.Setting.DBShema != "homer7" {
		return fmt.Errorf("Invalid DBShema: %s, please use homer5 or homer7", config.Setting.DBShema)
	}
	if config.Setting.DBShema == "homer5" && config.Setting.DBDriver != "mysql" {
		return fmt.Errorf("homer5 has only mysql support")
	}
	if config.Setting.DBShema == "homer7" && config.Setting.DBDriver != "postgres" {
		return fmt.Errorf("homer7 has only postgres support")
	}

	err := d.DBH.setup()
	if err != nil {
		return err
	}

	for i := 0; i < config.Setting.DBWorker; i++ {
		go func() {
			d.DBH.insert(d.Chan)
		}()
	}
	return nil
}

func (d *Database) End() {
	close(d.Chan)
}

func connectString(dbName string) (string, error) {
	var dsn string
	addr := strings.Split(config.Setting.DBAddr, ":")
	if len(addr) != 2 {
		return "", fmt.Errorf("wrong database connection format: %v, it should be localhost:3306", config.Setting.DBAddr)
	}
	if (addr[1] == "3306" && config.Setting.DBDriver == "postgres") ||
		addr[1] == "5432" && config.Setting.DBDriver == "mysql" {
		return "", fmt.Errorf("don't use port: %s, for db driver: %s", addr[1], config.Setting.DBDriver)
	}

	if config.Setting.DBDriver == "mysql" {
		if addr[0] == "unix" {
			// user:password@unix(/tmp/mysql.sock)/dbname?loc=Local
			dsn = config.Setting.DBUser + ":" + config.Setting.DBPass +
				"@unix(" + addr[1] + ")/" + dbName +
				"?" + url.QueryEscape("charset=utf8mb4&parseTime=true")
		} else {
			// user:password@tcp(localhost:5555)/dbname?tls=skip-verify&autocommit=true
			dsn = config.Setting.DBUser + ":" + config.Setting.DBPass +
				"@tcp(" + addr[0] + ":" + addr[1] + ")/" + dbName +
				"?" + url.QueryEscape("charset=utf8mb4&parseTime=true")
		}
	} else {
		if dbName == "" {
			dbName = "''"
		}
		if addr[0] == "unix" {
			addr[0] = addr[1]
			addr[1] = "''"
		}
		dsn = "sslmode=disable connect_timeout=2" +
			" host=" + addr[0] +
			" port=" + addr[1] +
			" dbname=" + dbName +
			" user=" + config.Setting.DBUser +
			" password=" + config.Setting.DBPass
	}
	return dsn, nil
}
