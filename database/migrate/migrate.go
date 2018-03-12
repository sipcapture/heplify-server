package migrate

import (
	"net/url"
	"time"

	"github.com/gocraft/dbr"
	"github.com/negbie/dotsql"
	"github.com/negbie/heplify-server/config"
	"github.com/negbie/heplify-server/logp"
)

func CreateDatabases(addr []string) error {
	//var err error

	if config.Setting.DBDriver == "mysql" {
		db, err := dbr.Open(config.Setting.DBDriver, config.Setting.DBUser+":"+config.Setting.DBPass+"@tcp("+addr[0]+":"+addr[1]+")/?"+url.QueryEscape("charset=utf8mb4&parseTime=true"), nil)
		if err != nil {
			return err
		}
		defer db.Close()

		_, err = db.Exec("CREATE DATABASE IF NOT EXISTS " + config.Setting.DBData + ` DEFAULT CHARACTER SET = 'utf8' DEFAULT COLLATE 'utf8_general_ci';`)
		if err != nil {
			logp.Info("%v", err)
		}
		_, err = db.Exec("CREATE DATABASE IF NOT EXISTS " + config.Setting.DBConf + ` DEFAULT CHARACTER SET = 'utf8' DEFAULT COLLATE 'utf8_general_ci';`)
		if err != nil {
			logp.Info("%v", err)
		}
		_, err = db.Exec(`CREATE USER IF NOT EXISTS 'homer_user'@'localhost' IDENTIFIED BY 'homer_password';`)
		if err != nil {
			logp.Info("%v", err)
		}
		_, err = db.Exec("GRANT ALL ON " + config.Setting.DBData + `.* TO 'homer_user'@'localhost';`)
		if err != nil {
			logp.Info("%v", err)
		}
		_, err = db.Exec("GRANT ALL ON " + config.Setting.DBConf + `.* TO 'homer_user'@'localhost';`)
		if err != nil {
			logp.Info("%v", err)
		}
		db.Close()

	} else if config.Setting.DBDriver == "postgres" {
		db, err := dbr.Open(config.Setting.DBDriver, " host="+addr[0]+" port="+addr[1]+" user="+config.Setting.DBUser+" password="+config.Setting.DBPass+" sslmode=disable", nil)
		if err != nil {
			return err
		}
		defer db.Close()

		_, err = db.Exec("CREATE DATABASE " + config.Setting.DBData)
		if err != nil {
			logp.Info("%v", err)
		}
		_, err = db.Exec("CREATE DATABASE " + config.Setting.DBConf)
		if err != nil {
			logp.Info("%v", err)
		}
		_, err = db.Exec(`CREATE USER homer_user WITH PASSWORD 'homer_password';`)
		if err != nil {
			logp.Info("%v", err)
		}
		_, err = db.Exec("GRANT postgres to homer_user;")
		if err != nil {
			logp.Info("%v", err)
		}
		_, err = db.Exec("GRANT ALL PRIVILEGES ON DATABASE " + config.Setting.DBData + " TO homer_user;")
		if err != nil {
			logp.Info("%v", err)
		}
		_, err = db.Exec("GRANT ALL PRIVILEGES ON DATABASE " + config.Setting.DBConf + " TO homer_user;")
		if err != nil {
			logp.Info("%v", err)
		}
		_, err = db.Exec("CREATE TABLESPACE homer OWNER homer_user LOCATION '" + config.Setting.DBPath + "';")
		if err != nil {
			logp.Info("%v", err)
		}
		_, err = db.Exec("GRANT ALL ON TABLESPACE homer TO homer_user;")
		if err != nil {
			logp.Info("%v", err)
		}
		_, err = db.Exec("GRANT ALL PRIVILEGES ON ALL TABLES IN SCHEMA public TO homer_user;")
		if err != nil {
			logp.Info("%v", err)
		}
		_, err = db.Exec("GRANT ALL PRIVILEGES ON ALL SEQUENCES IN SCHEMA public TO homer_user;")
		if err != nil {
			logp.Info("%v", err)
		}

		db.Close()
	}

	return nil
}

func CreateDataTables(addr []string) error {
	if config.Setting.DBDriver == "mysql" {
		db, err := dbr.Open(config.Setting.DBDriver, config.Setting.DBUser+":"+config.Setting.DBPass+"@tcp("+addr[0]+":"+addr[1]+")/"+config.Setting.DBData+"?"+url.QueryEscape("charset=utf8mb4&parseTime=true"), nil)
		if err != nil {
			return err
		}
		defer db.Close()

		dot, err := dotsql.LoadFromFileReplace("../../database/migrate/mysql/tbldata.sql", "20110111", time.Now().Format("20060102"))
		if err != nil {
			return err
		}

		for k := range dot.QueryMap() {
			_, err = dot.Exec(db, k)
			if err != nil {
				return err
			}
		}

		db.Close()
	} else if config.Setting.DBDriver == "postgres" {
		db, err := dbr.Open(config.Setting.DBDriver, " host="+addr[0]+" port="+addr[1]+" dbname="+config.Setting.DBData+" user="+config.Setting.DBUser+" password="+config.Setting.DBPass+" sslmode=disable", nil)
		if err != nil {
			return err
		}
		defer db.Close()

		dot, err := dotsql.LoadFromFileReplace("../../database/migrate/pgsql/tbldata.sql", "20110111", time.Now().Format("20060102"))
		if err != nil {
			return err
		}

		for k := range dot.QueryMap() {
			_, err = dot.Exec(db, k)
			if err != nil {
				return err
			}
		}

		db.Close()
	}
	return nil
}

func CreateConfTables(addr []string) error {
	if config.Setting.DBDriver == "mysql" {
		db, err := dbr.Open(config.Setting.DBDriver, config.Setting.DBUser+":"+config.Setting.DBPass+"@tcp("+addr[0]+":"+addr[1]+")/"+config.Setting.DBConf+"?"+url.QueryEscape("charset=utf8mb4&parseTime=true"), nil)
		if err != nil {
			return err
		}
		defer db.Close()

		dot, err := dotsql.LoadFromFileReplace("../../database/migrate/mysql/tblconf.sql", "20110111", time.Now().Format("20060102"))
		if err != nil {
			return err
		}

		for k := range dot.QueryMap() {
			_, err = dot.Exec(db, k)
			if err != nil {
				return err
			}
		}

		dot, err = dotsql.LoadFromFileReplace("../../database/migrate/mysql/insconf.sql", "20110111", time.Now().Format("20060102"))
		if err != nil {
			return err
		}

		for k := range dot.QueryMap() {
			_, err = dot.Exec(db, k)
			if err != nil {
				logp.Info("%v", err)
			}
		}

		db.Close()
	} else if config.Setting.DBDriver == "postgres" {
		db, err := dbr.Open(config.Setting.DBDriver, " host="+addr[0]+" port="+addr[1]+" dbname="+config.Setting.DBConf+" user="+config.Setting.DBUser+" password="+config.Setting.DBPass+" sslmode=disable", nil)
		if err != nil {
			return err
		}
		defer db.Close()

		dot, err := dotsql.LoadFromFileReplace("../../database/migrate/pgsql/tblconf.sql", "20110111", time.Now().Format("20060102"))
		if err != nil {
			return err
		}

		for k := range dot.QueryMap() {
			_, err = dot.Exec(db, k)
			if err != nil {
				return err
			}
		}

		dot, err = dotsql.LoadFromFileReplace("../../database/migrate/pgsql/insconf.sql", "20110111", time.Now().Format("20060102"))
		if err != nil {
			return err
		}

		for k := range dot.QueryMap() {
			_, err = dot.Exec(db, k)
			if err != nil {
				logp.Info("%v", err)
			}
		}

		db.Close()
	}
	return nil
}
