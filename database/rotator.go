package database

import (
	"fmt"
	"net/url"
	"strings"
	"time"

	raven "github.com/getsentry/raven-go"
	"github.com/gobuffalo/packr"
	"github.com/gocraft/dbr"
	"github.com/negbie/dotsql"
	"github.com/negbie/heplify-server/config"
	"github.com/negbie/heplify-server/logp"
	"github.com/robfig/cron"
)

type Rotator struct {
	addr []string
	box  *packr.Box
}

func NewRotator(b *packr.Box) *Rotator {
	return &Rotator{
		addr: strings.Split(config.Setting.DBAddr, ":"),
		box:  b,
	}
}

var curDay = strings.NewReplacer(
	"TableDate", time.Now().Format("20060102"),
	"PartitionName", time.Now().Format("20060102"),
	"PartitionDate", time.Now().Format("2006-01-02"),
)

var nextDay = strings.NewReplacer(
	"TableDate", time.Now().Add(time.Hour*24+1).Format("20060102"),
	"PartitionName", time.Now().Add(time.Hour*24+1).Format("20060102"),
	"PartitionDate", time.Now().Add(time.Hour*24+1).Format("2006-01-02"),
)

var twoDay = strings.NewReplacer(
	"TableDate", time.Now().Add(time.Hour*24+2).Format("20060102"),
	"PartitionName", time.Now().Add(time.Hour*24+2).Format("20060102"),
	"PartitionDate", time.Now().Add(time.Hour*24+2).Format("2006-01-02"),
)

var dropDay = strings.NewReplacer(
	"TableDate", time.Now().Add((time.Hour*24*time.Duration(config.Setting.DBDropDays))*-1).Format("20060102"),
	"PartitionName", time.Now().Add((time.Hour*24*time.Duration(config.Setting.DBDropDays))*-1).Format("20060102"),
	"PartitionDate", time.Now().Add((time.Hour*24*time.Duration(config.Setting.DBDropDays))*-1).Format("2006-01-02"),
)

func (r *Rotator) CreateDatabases() (err error) {
	if config.Setting.DBDriver == "mysql" {
		db, err := dbr.Open(config.Setting.DBDriver, config.Setting.DBUser+":"+config.Setting.DBPass+"@tcp("+r.addr[0]+":"+r.addr[1]+")/?"+url.QueryEscape("charset=utf8mb4&parseTime=true"), nil)
		if err != nil {
			return err
		}
		defer db.Close()
		r.dbExec(db, "CREATE DATABASE IF NOT EXISTS "+config.Setting.DBDataTable+` DEFAULT CHARACTER SET = 'utf8' DEFAULT COLLATE 'utf8_general_ci';`)
		r.dbExec(db, "CREATE DATABASE IF NOT EXISTS "+config.Setting.DBConfTable+` DEFAULT CHARACTER SET = 'utf8' DEFAULT COLLATE 'utf8_general_ci';`)
		r.dbExec(db, `CREATE USER IF NOT EXISTS 'homer_user'@'localhost' IDENTIFIED BY 'homer_password';`)
		r.dbExec(db, `CREATE USER IF NOT EXISTS 'homer_user'@'192.168.0.0/255.255.0.0' IDENTIFIED BY 'homer_password';`)
		r.dbExec(db, "GRANT ALL ON "+config.Setting.DBDataTable+`.* TO 'homer_user'@'localhost';`)
		r.dbExec(db, "GRANT ALL ON "+config.Setting.DBConfTable+`.* TO 'homer_user'@'localhost';`)
		r.dbExec(db, "GRANT ALL ON "+config.Setting.DBDataTable+`.* TO 'homer_user'@'192.168.0.0/255.255.0.0';`)
		r.dbExec(db, "GRANT ALL ON "+config.Setting.DBConfTable+`.* TO 'homer_user'@'192.168.0.0/255.255.0.0';`)

	} else if config.Setting.DBDriver == "postgres" {
		db, err := dbr.Open(config.Setting.DBDriver, " host="+r.addr[0]+" port="+r.addr[1]+" user="+config.Setting.DBUser+" password="+config.Setting.DBPass+" sslmode=disable", nil)
		if err != nil {
			return err
		}
		defer db.Close()
		r.dbExec(db, "CREATE DATABASE "+config.Setting.DBDataTable)
		r.dbExec(db, "CREATE DATABASE "+config.Setting.DBConfTable)
		r.dbExec(db, `CREATE USER homer_user WITH PASSWORD 'homer_password';`)
		r.dbExec(db, "GRANT postgres to homer_user;")
		r.dbExec(db, "GRANT ALL PRIVILEGES ON DATABASE "+config.Setting.DBDataTable+" TO homer_user;")
		r.dbExec(db, "GRANT ALL PRIVILEGES ON DATABASE "+config.Setting.DBConfTable+" TO homer_user;")
		r.dbExec(db, "CREATE TABLESPACE homer OWNER homer_user LOCATION '"+config.Setting.DBPath+"';")
		r.dbExec(db, "GRANT ALL ON TABLESPACE homer TO homer_user;")
		r.dbExec(db, "GRANT ALL PRIVILEGES ON ALL TABLES IN SCHEMA public TO homer_user;")
		r.dbExec(db, "GRANT ALL PRIVILEGES ON ALL SEQUENCES IN SCHEMA public TO homer_user;")
	}
	return nil
}

func (r *Rotator) CreateDataTables(pattern *strings.Replacer) (err error) {
	if config.Setting.DBDriver == "mysql" {
		db, err := dbr.Open(config.Setting.DBDriver, config.Setting.DBUser+":"+config.Setting.DBPass+"@tcp("+r.addr[0]+":"+r.addr[1]+")/"+config.Setting.DBDataTable+"?"+url.QueryEscape("charset=utf8mb4&parseTime=true"), nil)
		if err != nil {
			return err
		}
		defer db.Close()
		r.dbExecFile(db, r.box.String("mysql/tbldata.sql"), pattern)
	} else if config.Setting.DBDriver == "postgres" {
		db, err := dbr.Open(config.Setting.DBDriver, " host="+r.addr[0]+" port="+r.addr[1]+" dbname="+config.Setting.DBDataTable+" user="+config.Setting.DBUser+" password="+config.Setting.DBPass+" sslmode=disable", nil)
		if err != nil {
			return err
		}
		defer db.Close()
		r.dbExecFile(db, r.box.String("pgsql/tbldata.sql"), pattern)
		r.dbExecFile(db, r.box.String("pgsql/pardata.sql"), pattern)
		r.dbExecFile(db, r.box.String("pgsql/altdata.sql"), pattern)
		r.dbExecFile(db, r.box.String("pgsql/inddata.sql"), pattern)
	}
	return nil
}

func (r *Rotator) CreateConfTables(pattern *strings.Replacer) (err error) {
	if config.Setting.DBDriver == "mysql" {
		db, err := dbr.Open(config.Setting.DBDriver, config.Setting.DBUser+":"+config.Setting.DBPass+"@tcp("+r.addr[0]+":"+r.addr[1]+")/"+config.Setting.DBConfTable+"?"+url.QueryEscape("charset=utf8mb4&parseTime=true"), nil)
		if err != nil {
			return err
		}
		defer db.Close()
		r.dbExecFile(db, r.box.String("mysql/tblconf.sql"), pattern)
		r.dbExecFile(db, r.box.String("mysql/insconf.sql"), pattern)
	} else if config.Setting.DBDriver == "postgres" {
		db, err := dbr.Open(config.Setting.DBDriver, " host="+r.addr[0]+" port="+r.addr[1]+" dbname="+config.Setting.DBConfTable+" user="+config.Setting.DBUser+" password="+config.Setting.DBPass+" sslmode=disable", nil)
		if err != nil {
			return err
		}
		defer db.Close()
		r.dbExecFile(db, r.box.String("pgsql/tblconf.sql"), pattern)
		r.dbExecFile(db, r.box.String("pgsql/indconf.sql"), pattern)
		r.dbExecFile(db, r.box.String("pgsql/insconf.sql"), pattern)
	}
	return nil
}

func (r *Rotator) DropTables(pattern *strings.Replacer) (err error) {
	if config.Setting.DBDriver == "mysql" {
		db, err := dbr.Open(config.Setting.DBDriver, config.Setting.DBUser+":"+config.Setting.DBPass+"@tcp("+r.addr[0]+":"+r.addr[1]+")/"+config.Setting.DBDataTable+"?"+url.QueryEscape("charset=utf8mb4&parseTime=true"), nil)
		if err != nil {
			return err
		}
		defer db.Close()
		r.dbExecFile(db, r.box.String("mysql/droptbl.sql"), pattern)
	} else if config.Setting.DBDriver == "postgres" {
		db, err := dbr.Open(config.Setting.DBDriver, " host="+r.addr[0]+" port="+r.addr[1]+" dbname="+config.Setting.DBDataTable+" user="+config.Setting.DBUser+" password="+config.Setting.DBPass+" sslmode=disable", nil)
		if err != nil {
			return err
		}
		defer db.Close()
		r.dbExecFile(db, r.box.String("pgsql/droppar.sql"), pattern)
		r.dbExecFile(db, r.box.String("pgsql/droptbl.sql"), pattern)
	}
	return nil
}

func (r *Rotator) dbExecFile(db *dbr.Connection, file string, pattern *strings.Replacer) {
	dot, err := dotsql.LoadFromString(pattern.Replace(file))
	if err != nil {
		logp.Debug("rotator", "dbExecFile:\n%s\n\n", err)
	}

	for k, v := range dot.QueryMap() {
		logp.Debug("rotator", "queryMap:\n%s\n\n", v)
		if config.Setting.SentryDSN != "" {
			raven.CaptureError(fmt.Errorf("%v", v), nil)
		}
		_, err = dot.Exec(db, k)
		if err != nil {
			logp.Debug("rotator", "dotExec:\n%s\n\n", err)
		}
	}
}

func (r *Rotator) dbExec(db *dbr.Connection, query string) {
	_, err := db.Exec(query)
	if err != nil {
		logp.Debug("rotator", "dbExec:\n%s\n\n", err)
	}
}

func (r *Rotator) Rotate() (err error) {
	r.initTables()
	initRetry := 0
	initJob := cron.New()
	initJob.AddFunc("@every 30s", func() {
		initRetry++
		r.initTables()
		if initRetry == 2 {
			initJob.Stop()
		}

	})
	initJob.Start()

	createJob := cron.New()
	logp.Info("Start daily create data table job at 03:15:00\n")
	createJob.AddFunc("0 15 03 * * *", func() {
		if err := r.CreateDataTables(nextDay); err != nil {
			logp.Err("%v", err)
		}
		if err := r.CreateDataTables(twoDay); err != nil {
			logp.Err("%v", err)
		}
		logp.Info("Finished create data table job next will run at %v\n", time.Now().Add(time.Hour*24+1))
	})
	createJob.Start()

	if config.Setting.DBDropDays > 0 {
		dropJob := cron.New()
		logp.Info("Start daily drop data table job at 03:45:00\n")
		dropJob.AddFunc("0 45 03 * * *", func() {
			if err := r.DropTables(dropDay); err != nil {
				logp.Err("%v", err)
			}
			logp.Info("Finished drop data table job next will run at %v\n", time.Now().Add(time.Hour*24+1))
		})
		dropJob.Start()
	}
	return nil
}

func (r *Rotator) initTables() {
	if config.Setting.DBUser == "root" {
		if err := r.CreateDatabases(); err != nil {
			logp.Err("%v", err)
		}
	}
	if err := r.CreateConfTables(curDay); err != nil {
		logp.Err("%v", err)
	}
	if err := r.CreateDataTables(curDay); err != nil {
		logp.Err("%v", err)
	}
	if err := r.CreateDataTables(nextDay); err != nil {
		logp.Err("%v", err)
	}
}
