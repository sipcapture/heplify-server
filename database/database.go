package database

import (
	"fmt"
	"runtime"
	"strings"

	"github.com/negbie/heplify-server/config"
	"github.com/negbie/heplify-server/decoder"
)

type Database struct {
	H    DBHandler
	Chan chan *decoder.HEP
}

type DBHandler interface {
	setup() error
	insert(chan *decoder.HEP)
}

func New(name string) *Database {
	var register = map[string]DBHandler{
		"mysql":    new(MySQL),
		"postgres": new(Postgres),
		"mock":     new(Mock),
	}

	return &Database{
		H: register[name],
	}
}

func (d *Database) Run() error {
	driver := config.Setting.DBDriver
	shema := config.Setting.DBShema
	worker := config.Setting.DBWorker
	if driver != "mysql" && driver != "postgres" && driver != "mock" {
		return fmt.Errorf("invalid DBDriver: %s, please use mysql or postgres", driver)
	}
	if shema != "homer5" && shema != "homer7" && shema != "mock" {
		return fmt.Errorf("invalid DBShema: %s, please use homer5 or homer7", shema)
	}
	if shema == "homer5" && driver != "mysql" {
		return fmt.Errorf("homer5 has only mysql support")
	}
	if shema == "homer7" && driver != "postgres" {
		return fmt.Errorf("homer7 has only postgres support")
	}

	err := d.H.setup()
	if err != nil {
		return err
	}

	if worker > runtime.NumCPU() {
		worker = runtime.NumCPU()
	}

	for i := 0; i < worker; i++ {
		go func() {
			d.H.insert(d.Chan)
		}()
	}
	return nil
}

func (d *Database) End() {
	close(d.Chan)
}

func connectString(dbName string) (string, error) {
	var dsn string
	driver := config.Setting.DBDriver
	addr := strings.Split(config.Setting.DBAddr, ":")
	if len(addr) != 2 {
		return "", fmt.Errorf("wrong database connection format: %v, it should be localhost:3306", config.Setting.DBAddr)
	}
	if (addr[1] == "3306" && driver == "postgres") ||
		addr[1] == "5432" && driver == "mysql" {
		return "", fmt.Errorf("don't use port: %s, for db driver: %s", addr[1], driver)
	}

	if driver == "mysql" {
		if addr[0] == "unix" {
			// user:password@unix(/tmp/mysql.sock)/dbname?loc=Local
			dsn = config.Setting.DBUser + ":" + config.Setting.DBPass +
				"@unix(" + addr[1] + ")/" + dbName +
				"?collation=utf8mb4_unicode_ci&parseTime=true"
		} else {
			// user:password@tcp(localhost:5555)/dbname?tls=skip-verify&autocommit=true
			dsn = config.Setting.DBUser + ":" + config.Setting.DBPass +
				"@tcp(" + addr[0] + ":" + addr[1] + ")/" + dbName +
				"?collation=utf8mb4_unicode_ci&parseTime=true"
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
