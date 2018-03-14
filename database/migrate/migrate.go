package migrate

import (
	"net/url"
	"strings"
	"time"

	"github.com/gobuffalo/packr"

	"github.com/gocraft/dbr"
	"github.com/negbie/dotsql"
	"github.com/negbie/heplify-server/config"
	"github.com/negbie/heplify-server/logp"
)

var pattern = strings.NewReplacer("TableDate", time.Now().Format("20060102"), "PartitionName", time.Now().Format("20060102"), "PartitionDate", time.Now().Format("2006-01-02"))

func CreateDatabases(addr []string) error {
	if config.Setting.DBDriver == "mysql" {
		db, err := dbr.Open(config.Setting.DBDriver, config.Setting.DBUser+":"+config.Setting.DBPass+"@tcp("+addr[0]+":"+addr[1]+")/?"+url.QueryEscape("charset=utf8mb4&parseTime=true"), nil)
		if err != nil {
			return err
		}
		defer db.Close()
		dbExec(db, "CREATE DATABASE IF NOT EXISTS "+config.Setting.DBData+` DEFAULT CHARACTER SET = 'utf8' DEFAULT COLLATE 'utf8_general_ci';`)
		dbExec(db, "CREATE DATABASE IF NOT EXISTS "+config.Setting.DBConf+` DEFAULT CHARACTER SET = 'utf8' DEFAULT COLLATE 'utf8_general_ci';`)
		dbExec(db, `CREATE USER IF NOT EXISTS 'homer_user'@'localhost' IDENTIFIED BY 'homer_password';`)
		dbExec(db, `CREATE USER IF NOT EXISTS 'homer_user'@'192.168.0.0/255.255.0.0' IDENTIFIED BY 'homer_password';`)
		dbExec(db, "GRANT ALL ON "+config.Setting.DBData+`.* TO 'homer_user'@'localhost';`)
		dbExec(db, "GRANT ALL ON "+config.Setting.DBConf+`.* TO 'homer_user'@'localhost';`)
		dbExec(db, "GRANT ALL ON "+config.Setting.DBData+`.* TO 'homer_user'@'192.168.0.0/255.255.0.0';`)
		dbExec(db, "GRANT ALL ON "+config.Setting.DBConf+`.* TO 'homer_user'@'192.168.0.0/255.255.0.0';`)

	} else if config.Setting.DBDriver == "postgres" {
		db, err := dbr.Open(config.Setting.DBDriver, " host="+addr[0]+" port="+addr[1]+" user="+config.Setting.DBUser+" password="+config.Setting.DBPass+" sslmode=disable", nil)
		if err != nil {
			return err
		}
		defer db.Close()
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
	return nil
}

func CreateDataTables(addr []string, box packr.Box) (err error) {
	if config.Setting.DBDriver == "mysql" {
		db, err := dbr.Open(config.Setting.DBDriver, config.Setting.DBUser+":"+config.Setting.DBPass+"@tcp("+addr[0]+":"+addr[1]+")/"+config.Setting.DBData+"?"+url.QueryEscape("charset=utf8mb4&parseTime=true"), nil)
		if err != nil {
			return err
		}
		defer db.Close()
		dbExecFile(db, box.String("mysql/tbldata.sql"))
	} else if config.Setting.DBDriver == "postgres" {
		db, err := dbr.Open(config.Setting.DBDriver, " host="+addr[0]+" port="+addr[1]+" dbname="+config.Setting.DBData+" user="+config.Setting.DBUser+" password="+config.Setting.DBPass+" sslmode=disable", nil)
		if err != nil {
			return err
		}
		defer db.Close()
		dbExecFile(db, box.String("pgsql/tbldata.sql"))
		dbExecFile(db, box.String("pgsql/pardata.sql"))
		dbExecFile(db, box.String("pgsql/altdata.sql"))
		dbExecFile(db, box.String("pgsql/inddata.sql"))
	}
	return nil
}

func CreateConfTables(addr []string, box packr.Box) (err error) {
	if config.Setting.DBDriver == "mysql" {
		db, err := dbr.Open(config.Setting.DBDriver, config.Setting.DBUser+":"+config.Setting.DBPass+"@tcp("+addr[0]+":"+addr[1]+")/"+config.Setting.DBConf+"?"+url.QueryEscape("charset=utf8mb4&parseTime=true"), nil)
		if err != nil {
			logp.Info("%v", err)
		}
		defer db.Close()
		dbExecFile(db, box.String("mysql/tblconf.sql"))
		dbExecFile(db, box.String("mysql/insconf.sql"))
	} else if config.Setting.DBDriver == "postgres" {
		db, err := dbr.Open(config.Setting.DBDriver, " host="+addr[0]+" port="+addr[1]+" dbname="+config.Setting.DBConf+" user="+config.Setting.DBUser+" password="+config.Setting.DBPass+" sslmode=disable", nil)
		if err != nil {
			logp.Info("%v", err)
		}
		defer db.Close()
		dbExecFile(db, box.String("pgsql/tblconf.sql"))
		dbExecFile(db, box.String("pgsql/indconf.sql"))
		dbExecFile(db, box.String("pgsql/insconf.sql"))
	}
	return nil
}

func DropTables(addr []string, box packr.Box) (err error) {
	if config.Setting.DBDriver == "mysql" {
		db, err := dbr.Open(config.Setting.DBDriver, config.Setting.DBUser+":"+config.Setting.DBPass+"@tcp("+addr[0]+":"+addr[1]+")/"+config.Setting.DBData+"?"+url.QueryEscape("charset=utf8mb4&parseTime=true"), nil)
		if err != nil {
			return err
		}
		defer db.Close()
		dbExecFile(db, box.String("mysql/droptbl.sql"))
	} else if config.Setting.DBDriver == "postgres" {
		db, err := dbr.Open(config.Setting.DBDriver, " host="+addr[0]+" port="+addr[1]+" dbname="+config.Setting.DBData+" user="+config.Setting.DBUser+" password="+config.Setting.DBPass+" sslmode=disable", nil)
		if err != nil {
			return err
		}
		defer db.Close()
		dbExecFile(db, box.String("pgsql/droppar.sql"))
		dbExecFile(db, box.String("pgsql/droptbl.sql"))
	}
	return nil
}

func dbExecFile(db *dbr.Connection, file string) {
	dot, err := dotsql.LoadFromString(pattern.Replace(file))
	if err != nil {
		logp.Err("%v", err)
	}

	for k, v := range dot.QueryMap() {
		logp.Debug("rotate", "%s\n\n", v)
		_, err = dot.Exec(db, k)
		if err != nil {
			logp.Err("%v", err)
		}
	}
}

func dbExec(db *dbr.Connection, query string) {
	_, err := db.Exec(query)
	if err != nil {
		logp.Err("%v", err)
	}
}
