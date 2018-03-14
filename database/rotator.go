package database

import (
	"net/url"
	"strings"
	"time"

	"github.com/gobuffalo/packr"
	"github.com/gocraft/dbr"
	"github.com/negbie/dotsql"
	"github.com/negbie/heplify-server/config"
	"github.com/negbie/heplify-server/logp"
	"github.com/robfig/cron"
)

type Rotator struct {
	addr    []string
	box     *packr.Box
	dbm     *dbr.Connection
	pattern *strings.Replacer
}

func NewRotator(b *packr.Box) *Rotator {
	return &Rotator{
		addr: strings.Split(config.Setting.DBAddr, ":"),
		box:  b,
		pattern: strings.NewReplacer(
			"TableDate", time.Now().Format("20060102"),
			"PartitionName", time.Now().Format("20060102"),
			"PartitionDate", time.Now().Format("2006-01-02"),
		),
	}
}

func (r *Rotator) CreateDatabases() (err error) {
	if config.Setting.DBDriver == "mysql" {
		r.dbm, err = dbr.Open(config.Setting.DBDriver, config.Setting.DBUser+":"+config.Setting.DBPass+"@tcp("+r.addr[0]+":"+r.addr[1]+")/?"+url.QueryEscape("charset=utf8mb4&parseTime=true"), nil)
		if err != nil {
			return err
		}
		defer r.dbm.Close()
		r.dbExec("CREATE DATABASE IF NOT EXISTS " + config.Setting.DBData + ` DEFAULT CHARACTER SET = 'utf8' DEFAULT COLLATE 'utf8_general_ci';`)
		r.dbExec("CREATE DATABASE IF NOT EXISTS " + config.Setting.DBConf + ` DEFAULT CHARACTER SET = 'utf8' DEFAULT COLLATE 'utf8_general_ci';`)
		r.dbExec(`CREATE USER IF NOT EXISTS 'homer_user'@'localhost' IDENTIFIED BY 'homer_password';`)
		r.dbExec(`CREATE USER IF NOT EXISTS 'homer_user'@'192.168.0.0/255.255.0.0' IDENTIFIED BY 'homer_password';`)
		r.dbExec("GRANT ALL ON " + config.Setting.DBData + `.* TO 'homer_user'@'localhost';`)
		r.dbExec("GRANT ALL ON " + config.Setting.DBConf + `.* TO 'homer_user'@'localhost';`)
		r.dbExec("GRANT ALL ON " + config.Setting.DBData + `.* TO 'homer_user'@'192.168.0.0/255.255.0.0';`)
		r.dbExec("GRANT ALL ON " + config.Setting.DBConf + `.* TO 'homer_user'@'192.168.0.0/255.255.0.0';`)

	} else if config.Setting.DBDriver == "postgres" {
		r.dbm, err = dbr.Open(config.Setting.DBDriver, " host="+r.addr[0]+" port="+r.addr[1]+" user="+config.Setting.DBUser+" password="+config.Setting.DBPass+" sslmode=disable", nil)
		if err != nil {
			return err
		}
		defer r.dbm.Close()
		r.dbExec("CREATE DATABASE " + config.Setting.DBData)
		r.dbExec("CREATE DATABASE " + config.Setting.DBConf)
		r.dbExec(`CREATE USER homer_user WITH PASSWORD 'homer_password';`)
		r.dbExec("GRANT postgres to homer_user;")
		r.dbExec("GRANT ALL PRIVILEGES ON DATABASE " + config.Setting.DBData + " TO homer_user;")
		r.dbExec("GRANT ALL PRIVILEGES ON DATABASE " + config.Setting.DBConf + " TO homer_user;")
		r.dbExec("CREATE TABLESPACE homer OWNER homer_user LOCATION '" + config.Setting.DBPath + "';")
		r.dbExec("GRANT ALL ON TABLESPACE homer TO homer_user;")
		r.dbExec("GRANT ALL PRIVILEGES ON ALL TABLES IN SCHEMA public TO homer_user;")
		r.dbExec("GRANT ALL PRIVILEGES ON ALL SEQUENCES IN SCHEMA public TO homer_user;")
	}
	return nil
}

func (r *Rotator) CreateDataTables() (err error) {
	if config.Setting.DBDriver == "mysql" {
		r.dbm, err = dbr.Open(config.Setting.DBDriver, config.Setting.DBUser+":"+config.Setting.DBPass+"@tcp("+r.addr[0]+":"+r.addr[1]+")/"+config.Setting.DBData+"?"+url.QueryEscape("charset=utf8mb4&parseTime=true"), nil)
		if err != nil {
			return err
		}
		defer r.dbm.Close()
		r.dbExecFile(r.box.String("mysql/tbldata.sql"))
	} else if config.Setting.DBDriver == "postgres" {
		r.dbm, err = dbr.Open(config.Setting.DBDriver, " host="+r.addr[0]+" port="+r.addr[1]+" dbname="+config.Setting.DBData+" user="+config.Setting.DBUser+" password="+config.Setting.DBPass+" sslmode=disable", nil)
		if err != nil {
			return err
		}
		defer r.dbm.Close()
		r.dbExecFile(r.box.String("pgsql/tbldata.sql"))
		r.dbExecFile(r.box.String("pgsql/pardata.sql"))
		r.dbExecFile(r.box.String("pgsql/altdata.sql"))
		r.dbExecFile(r.box.String("pgsql/inddata.sql"))
	}
	return nil
}

func (r *Rotator) CreateConfTables() (err error) {
	if config.Setting.DBDriver == "mysql" {
		r.dbm, err = dbr.Open(config.Setting.DBDriver, config.Setting.DBUser+":"+config.Setting.DBPass+"@tcp("+r.addr[0]+":"+r.addr[1]+")/"+config.Setting.DBConf+"?"+url.QueryEscape("charset=utf8mb4&parseTime=true"), nil)
		if err != nil {
			logp.Info("%v", err)
		}
		defer r.dbm.Close()
		r.dbExecFile(r.box.String("mysql/tblconf.sql"))
		r.dbExecFile(r.box.String("mysql/insconf.sql"))
	} else if config.Setting.DBDriver == "postgres" {
		r.dbm, err = dbr.Open(config.Setting.DBDriver, " host="+r.addr[0]+" port="+r.addr[1]+" dbname="+config.Setting.DBConf+" user="+config.Setting.DBUser+" password="+config.Setting.DBPass+" sslmode=disable", nil)
		if err != nil {
			logp.Info("%v", err)
		}
		defer r.dbm.Close()
		r.dbExecFile(r.box.String("pgsql/tblconf.sql"))
		r.dbExecFile(r.box.String("pgsql/indconf.sql"))
		r.dbExecFile(r.box.String("pgsql/insconf.sql"))
	}
	return nil
}

func (r *Rotator) DropTables() (err error) {
	if config.Setting.DBDriver == "mysql" {
		r.dbm, err = dbr.Open(config.Setting.DBDriver, config.Setting.DBUser+":"+config.Setting.DBPass+"@tcp("+r.addr[0]+":"+r.addr[1]+")/"+config.Setting.DBData+"?"+url.QueryEscape("charset=utf8mb4&parseTime=true"), nil)
		if err != nil {
			return err
		}
		defer r.dbm.Close()
		r.dbExecFile(r.box.String("mysql/droptbl.sql"))
	} else if config.Setting.DBDriver == "postgres" {
		r.dbm, err = dbr.Open(config.Setting.DBDriver, " host="+r.addr[0]+" port="+r.addr[1]+" dbname="+config.Setting.DBData+" user="+config.Setting.DBUser+" password="+config.Setting.DBPass+" sslmode=disable", nil)
		if err != nil {
			return err
		}
		defer r.dbm.Close()
		r.dbExecFile(r.box.String("pgsql/droppar.sql"))
		r.dbExecFile(r.box.String("pgsql/droptbl.sql"))
	}
	return nil
}

func (r *Rotator) dbExecFile(file string) {
	dot, err := dotsql.LoadFromString(r.pattern.Replace(file))
	if err != nil {
		logp.Err("%v", err)
	}

	for k, v := range dot.QueryMap() {
		logp.Debug("rotator", "%s\n\n", v)
		_, err = dot.Exec(r.dbm, k)
		if err != nil {
			logp.Err("%v", err)
		}
	}
}

func (r *Rotator) dbExec(query string) {
	_, err := r.dbm.Exec(query)
	if err != nil {
		logp.Err("%v", err)
	}
}

func (r *Rotator) Rotate() (err error) {
	retention := time.Hour * 24 * time.Duration(config.Setting.DBDrop)

	r.pattern = strings.NewReplacer(
		"TableDate", time.Now().Add(time.Hour*24+1).Format("20060102"),
		"PartitionName", time.Now().Add(time.Hour*24+1).Format("20060102"),
		"PartitionDate", time.Now().Add(time.Hour*24+1).Format("2006-01-02"),
	)

	if err := r.CreateDataTables(); err != nil {
		logp.Err("%v", err)
	}

	c := cron.New()
	logp.Info("Start daily create data table job at 03:15:00\n\n")
	c.AddFunc("0 15 03 * * *", func() {
		r.pattern = strings.NewReplacer(
			"TableDate", time.Now().Add(time.Hour*24+1).Format("20060102"),
			"PartitionName", time.Now().Add(time.Hour*24+1).Format("20060102"),
			"PartitionDate", time.Now().Add(time.Hour*24+1).Format("2006-01-02"),
		)

		if err := r.CreateDataTables(); err != nil {
			logp.Err("%v", err)
		}
		logp.Info("Finished create data table job next will run at %v\n\n", time.Now().Add(time.Hour*24+1))
	})

	if config.Setting.DBDrop > 0 {
		logp.Info("Start daily drop data table job at 03:45:00\n\n")
		c.AddFunc("0 45 03 * * *", func() {
			r.pattern = strings.NewReplacer(
				"TableDate", time.Now().Add(retention*-1).Format("20060102"),
				"PartitionName", time.Now().Add(retention*-1).Format("20060102"),
				"PartitionDate", time.Now().Add(retention*-1).Format("2006-01-02"),
			)

			if err := r.DropTables(); err != nil {
				logp.Err("%v", err)
			}
			logp.Info("Finished drop data table job next will run at %v\n\n", time.Now().Add(time.Hour*24+1))
		})
	}
	c.Start()
	return nil
}
