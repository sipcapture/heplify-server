package migrate

import (
	"net/url"
	"time"

	"github.com/gobuffalo/packr"

	"github.com/gocraft/dbr"
	"github.com/negbie/dotsql"
	"github.com/negbie/heplify-server/config"
	"github.com/negbie/heplify-server/logp"
)

func CreateDatabases(addr []string) error {
	var err error
	var db *dbr.Connection

	if config.Setting.DBDriver == "mysql" {
		db, err = dbr.Open(config.Setting.DBDriver, config.Setting.DBUser+":"+config.Setting.DBPass+"@tcp("+addr[0]+":"+addr[1]+")/?"+url.QueryEscape("charset=utf8mb4&parseTime=true"), nil)
		if err != nil {
			return err
		}

		dbExec(db, "CREATE DATABASE IF NOT EXISTS "+config.Setting.DBData+` DEFAULT CHARACTER SET = 'utf8' DEFAULT COLLATE 'utf8_general_ci';`)
		dbExec(db, "CREATE DATABASE IF NOT EXISTS "+config.Setting.DBConf+` DEFAULT CHARACTER SET = 'utf8' DEFAULT COLLATE 'utf8_general_ci';`)
		dbExec(db, `CREATE USER IF NOT EXISTS 'homer_user'@'localhost' IDENTIFIED BY 'homer_password';`)
		dbExec(db, "GRANT ALL ON "+config.Setting.DBData+`.* TO 'homer_user'@'localhost';`)
		dbExec(db, "GRANT ALL ON "+config.Setting.DBConf+`.* TO 'homer_user'@'localhost';`)

	} else if config.Setting.DBDriver == "postgres" {
		db, err = dbr.Open(config.Setting.DBDriver, " host="+addr[0]+" port="+addr[1]+" user="+config.Setting.DBUser+" password="+config.Setting.DBPass+" sslmode=disable", nil)
		if err != nil {
			return err
		}

		dbExec(db, "CREATE DATABASE "+config.Setting.DBData)
		dbExec(db, "CREATE DATABASE "+config.Setting.DBConf)
		dbExec(db, `CREATE USER homer_user WITH PASSWORD 'homer_password';`)
		dbExec(db, "GRANT postgres to homer_user;")
		dbExec(db, "GRANT ALL PRIVILEGES ON DATABASE "+config.Setting.DBData+" TO homer_user;")
		dbExec(db, "GRANT ALL PRIVILEGES ON DATABASE "+config.Setting.DBConf+" TO homer_user;")
		dbExec(db, "CREATE TABLESPACE homer OWNER homer_user LOCATION '"+config.Setting.DBPath+"';")
		dbExec(db, "GRANT ALL ON TABLESPACE homer TO homer_user;")
		dbExec(db, "GRANT ALL PRIVILEGES ON ALL TABLES IN SCHEMA public TO homer_user;")
		dbExec(db, "GRANT ALL PRIVILEGES ON ALL SEQUENCES IN SCHEMA public TO homer_user;")
	}
	defer db.Close()
	return nil
}

func CreateDataTables(addr []string, box packr.Box) error {
	if config.Setting.DBDriver == "mysql" {
		dbExecFile(addr, config.Setting.DBData, box.String("mysql/tbldata.sql"), "20110111")
	} else if config.Setting.DBDriver == "postgres" {
		dbExecFile(addr, config.Setting.DBData, box.String("pgsql/tbldata.sql"), "20110111")
		dbExecFile(addr, config.Setting.DBData, box.String("pgsql/pardata.sql"), "20110111")
		dbExecFile(addr, config.Setting.DBData, box.String("pgsql/altdata.sql"), "20110111")
		dbExecFile(addr, config.Setting.DBData, box.String("pgsql/inddata.sql"), "20110111")
	}
	return nil
}

func CreateConfTables(addr []string, box packr.Box) error {
	if config.Setting.DBDriver == "mysql" {
		dbExecFile(addr, config.Setting.DBConf, box.String("mysql/tblconf.sql"), "20110111")
		dbExecFile(addr, config.Setting.DBConf, box.String("mysql/insconf.sql"), "20110111")
	} else if config.Setting.DBDriver == "postgres" {
		dbExecFile(addr, config.Setting.DBConf, box.String("pgsql/tblconf.sql"), "20110111")
		dbExecFile(addr, config.Setting.DBConf, box.String("pgsql/indconf.sql"), "20110111")
		dbExecFile(addr, config.Setting.DBConf, box.String("pgsql/insconf.sql"), "20110111")
	}
	return nil
}

func dbExecFile(addr []string, table, file, pattern string) {
	var err error
	var db *dbr.Connection

	if config.Setting.DBDriver == "mysql" {
		db, err = dbr.Open(config.Setting.DBDriver, config.Setting.DBUser+":"+config.Setting.DBPass+"@tcp("+addr[0]+":"+addr[1]+")/"+table+"?"+url.QueryEscape("charset=utf8mb4&parseTime=true"), nil)
		if err != nil {
			logp.Info("%v", err)
		}
	} else if config.Setting.DBDriver == "postgres" {
		db, err = dbr.Open(config.Setting.DBDriver, " host="+addr[0]+" port="+addr[1]+" dbname="+table+" user="+config.Setting.DBUser+" password="+config.Setting.DBPass+" sslmode=disable", nil)
		if err != nil {
			logp.Info("%v", err)
		}
	}
	defer db.Close()

	dot, err := dotsql.LoadFromStringReplace(file, pattern, time.Now().Format("20060102"))
	if err != nil {
		logp.Info("%v", err)
	}

	for k := range dot.QueryMap() {
		_, err = dot.Exec(db, k)
		if err != nil {
			logp.Info("%v", err)
		}
	}
	db.Close()
}

func dbExec(db *dbr.Connection, query string) {
	_, err := db.Exec(query)
	if err != nil {
		logp.Info("%v", err)
	}
}
