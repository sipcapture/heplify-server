package database

import (
	"net/url"
	"strconv"
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
	addr []string
	box  *packr.Box
	step int
}

func NewRotator(b *packr.Box) *Rotator {
	r := &Rotator{}
	r.addr = strings.Split(config.Setting.DBAddr, ":")
	r.box = b

	switch config.Setting.DBPartition {
	case "15m":
		r.step = 15
	case "30m":
		r.step = 30
	case "1h":
		r.step = 60
	default:
		r.step = 1440
	}
	return r
}

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

func (r *Rotator) CreateDataTables(pattern strings.Replacer) (err error) {
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
		r.dbExecPartitionFile(db, r.box.String("pgsql/pardata.sql"), pattern)
		r.dbExecPartitionFile(db, r.box.String("pgsql/inddata.sql"), pattern)
	}
	return nil
}

func (r *Rotator) CreateConfTables(pattern strings.Replacer) (err error) {
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
		r.dbExec(db, "CREATE EXTENSION pgcrypto;")
		r.dbExecFile(db, r.box.String("pgsql/tblconf.sql"), pattern)
		r.dbExecFile(db, r.box.String("pgsql/indconf.sql"), pattern)
		r.dbExecFile(db, r.box.String("pgsql/insconf.sql"), pattern)
	}
	return nil
}

func (r *Rotator) DropTables(pattern strings.Replacer) (err error) {
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
		r.dbExecPartitionFile(db, r.box.String("pgsql/droppar.sql"), pattern)
	}
	return nil
}

func (r *Rotator) dbExecFile(db *dbr.Connection, file string, pattern strings.Replacer) {
	dot, err := dotsql.LoadFromString(pattern.Replace(file))
	if err != nil {
		logp.Err("%s\n\n", err)
	}

	for _, query := range dot.QueryMap() {
		logp.Debug("rotator", "db query:\n%s\n\n", query)
		_, err := db.Exec(query)
		if err != nil {
			logp.Warn("%s\n\n", err)
		}
	}
}

func (r *Rotator) dbExecPartitionFile(db *dbr.Connection, file string, pattern strings.Replacer) {
	dot, err := dotsql.LoadFromString(pattern.Replace(file))
	if err != nil {
		logp.Err("%s\n\n", err)
	}

	for _, query := range dot.QueryMap() {
		if r.step == 1440 {
			logp.Debug("rotator", "db query:\n%s\n\n", query)
			_, err := db.Exec(partDay.Replace(query))
			if err != nil {
				logp.Warn("%s\n\n", err)
			}
		} else if r.step != 1440 {
			r.rotatePartitions(db, query)
		}
	}
}

func (r *Rotator) dbExec(db *dbr.Connection, query string) {
	_, err := db.Exec(query)
	if err != nil {
		logp.Warn("%s\n\n", err)
	}
}

func (r *Rotator) Rotate() (err error) {
	r.createTables()
	initRetry := 0
	initJob := cron.New()
	initJob.AddFunc("@every 30s", func() {
		initRetry++
		r.createTables()
		if initRetry == 2 {
			initJob.Stop()
		}
	})
	initJob.Start()

	createJob := cron.New()
	logp.Info("Start daily create data table job at 03:15:00\n")
	createJob.AddFunc("0 15 03 * * *", func() {
		if err := r.CreateDataTables(replaceNextDay()); err != nil {
			logp.Err("%v", err)
		}
		if err := r.CreateDataTables(replaceTwoDay()); err != nil {
			logp.Err("%v", err)
		}
		logp.Info("Finished create data table job next will run at %v\n", time.Now().Add(time.Hour*24+1))
	})
	createJob.Start()

	if config.Setting.DBDropDays > 0 {
		dropJob := cron.New()
		logp.Info("Start daily drop data table job at 03:45:00\n")
		dropJob.AddFunc("0 45 03 * * *", func() {
			if err := r.DropTables(replaceDropDay()); err != nil {
				logp.Err("%v", err)
			}
			logp.Info("Finished drop data table job next will run at %v\n", time.Now().Add(time.Hour*24+1))
		})
		dropJob.Start()
	}
	return nil
}

func (r *Rotator) createTables() {
	if config.Setting.DBUser == "root" || config.Setting.DBUser == "postgres" {
		if err := r.CreateDatabases(); err != nil {
			logp.Err("%v", err)
		}
	}
	if err := r.CreateConfTables(replaceCurDay()); err != nil {
		logp.Err("%v", err)
	}
	if err := r.CreateDataTables(replaceCurDay()); err != nil {
		logp.Err("%v", err)
	}
	if err := r.CreateDataTables(replaceNextDay()); err != nil {
		logp.Err("%v", err)
	}
}

func (r *Rotator) rotatePartitions(db *dbr.Connection, query string) {
	oldName := "pnr0000"
	newName := "pnr0"
	startTime := new(time.Time)
	oldStart := "StartTime"
	newStart := startTime.Add(time.Hour*time.Duration(0) + time.Minute*time.Duration(0)).Format("15:04")

	endTime := new(time.Time)
	oldEnd := "EndTime"
	newEnd := endTime.Add(time.Hour*time.Duration(0) + time.Minute*time.Duration(r.step) - 1).Format("15:04")

	for i := 0; i < 1440/r.step; i++ {
		if i > 0 {
			newName = "pnr" + strconv.Itoa(i)
			newStart = startTime.Add(time.Hour*time.Duration(0) + time.Minute*time.Duration(i*r.step)).Format("15:04")
			newEnd = endTime.Add(time.Hour*time.Duration(0) + time.Minute*time.Duration(i*r.step+r.step) - 1).Format("15:04")
		}
		query = strings.Replace(query, oldName, newName, -1)
		oldName = newName

		query = strings.Replace(query, oldStart, newStart, -1)
		oldStart = newStart

		query = strings.Replace(query, oldEnd, newEnd, -1)
		oldEnd = newEnd

		logp.Debug("rotator", "db query:\n%s\n\n", query)
		_, err := db.Exec(query)
		if err != nil {
			logp.Warn("%s\n\n", err)
		}
	}
}

func replaceCurDay() strings.Replacer {
	return *strings.NewReplacer(
		"TableDate", time.Now().Format("20060102"),
		"PartitionName", time.Now().Format("20060102"),
		"PartitionDate", time.Now().Format("2006-01-02"),
	)
}

func replaceNextDay() strings.Replacer {
	return *strings.NewReplacer(
		"TableDate", time.Now().Add(time.Hour*24+1).Format("20060102"),
		"PartitionName", time.Now().Add(time.Hour*24+1).Format("20060102"),
		"PartitionDate", time.Now().Add(time.Hour*24+1).Format("2006-01-02"),
	)
}

func replaceTwoDay() strings.Replacer {
	return *strings.NewReplacer(
		"TableDate", time.Now().Add(time.Hour*24+2).Format("20060102"),
		"PartitionName", time.Now().Add(time.Hour*24+2).Format("20060102"),
		"PartitionDate", time.Now().Add(time.Hour*24+2).Format("2006-01-02"),
	)
}

func replaceDropDay() strings.Replacer {
	return *strings.NewReplacer(
		"TableDate", time.Now().Add(time.Hour*24*time.Duration(config.Setting.DBDropDays*-1)).Format("20060102"),
		"PartitionName", time.Now().Add(time.Hour*24*time.Duration(config.Setting.DBDropDays*-1)).Format("20060102"),
		"PartitionDate", time.Now().Add(time.Hour*24*time.Duration(config.Setting.DBDropDays*-1)).Format("2006-01-02"),
	)
}

var partDay = strings.NewReplacer("StartTime", "00:00", "EndTime", "23:59")
