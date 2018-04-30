package database

import (
	"net/url"
	"strings"
	"time"

	"github.com/go-sql-driver/mysql"
	"github.com/gobuffalo/packr"
	"github.com/gocraft/dbr"
	"github.com/negbie/dotsql"
	"github.com/negbie/heplify-server/config"
	"github.com/negbie/heplify-server/logp"
	"github.com/robfig/cron"
)

const (
	partitionTime      = "{{time}}"
	partitionMinTime   = "{{minTime}}"
	partitionStartTime = "{{startTime}}"
	partitionEndTime   = "{{endTime}}"
)

type Rotator struct {
	addr    []string
	box     *packr.Box
	logStep int
	qosStep int
	sipStep int
}

func NewRotator(b *packr.Box) *Rotator {
	r := &Rotator{}
	r.addr = strings.Split(config.Setting.DBAddr, ":")
	r.box = b
	r.logStep = setStep(config.Setting.DBPartLog)
	r.qosStep = setStep(config.Setting.DBPartQos)
	r.sipStep = setStep(config.Setting.DBPartSip)
	return r
}

func (r *Rotator) CreateDatabases() (err error) {
	for {
		if config.Setting.DBDriver == "mysql" {
			db, err := dbr.Open(config.Setting.DBDriver, config.Setting.DBUser+":"+config.Setting.DBPass+"@tcp("+r.addr[0]+":"+r.addr[1]+")/?"+url.QueryEscape("charset=utf8mb4&parseTime=true"), nil)
			if err = db.Ping(); err != nil {
				db.Close()
				logp.Err("%v", err)
				time.Sleep(5 * time.Second)
			} else {
				r.dbExec(db, "CREATE DATABASE IF NOT EXISTS "+config.Setting.DBDataTable+` DEFAULT CHARACTER SET = 'utf8' DEFAULT COLLATE 'utf8_general_ci';`)
				r.dbExec(db, "CREATE DATABASE IF NOT EXISTS "+config.Setting.DBConfTable+` DEFAULT CHARACTER SET = 'utf8' DEFAULT COLLATE 'utf8_general_ci';`)
				r.dbExec(db, `CREATE USER IF NOT EXISTS 'homer_user'@'%' IDENTIFIED BY 'homer_password';`)
				r.dbExec(db, "GRANT ALL ON "+config.Setting.DBDataTable+`.* TO 'homer_user'@'%';`)
				r.dbExec(db, "GRANT ALL ON "+config.Setting.DBConfTable+`.* TO 'homer_user'@'%';`)
				db.Close()
				break
			}
		} else if config.Setting.DBDriver == "postgres" {
			db, err := dbr.Open(config.Setting.DBDriver, " host="+r.addr[0]+" port="+r.addr[1]+" user="+config.Setting.DBUser+" password="+config.Setting.DBPass+" sslmode=disable", nil)
			if err = db.Ping(); err != nil {
				db.Close()
				logp.Err("%v", err)
				time.Sleep(5 * time.Second)
			} else {
				r.dbExec(db, "CREATE DATABASE "+config.Setting.DBDataTable)
				r.dbExec(db, "CREATE DATABASE "+config.Setting.DBConfTable)
				r.dbExec(db, `CREATE USER homer_user WITH PASSWORD 'homer_password';`)
				r.dbExec(db, "GRANT postgres to homer_user;")
				r.dbExec(db, "GRANT ALL PRIVILEGES ON DATABASE "+config.Setting.DBDataTable+" TO homer_user;")
				r.dbExec(db, "GRANT ALL PRIVILEGES ON DATABASE "+config.Setting.DBConfTable+" TO homer_user;")
				r.dbExec(db, "CREATE TABLESPACE homer OWNER homer_user LOCATION '"+config.Setting.DBTableSpace+"';")
				r.dbExec(db, "GRANT ALL ON TABLESPACE homer TO homer_user;")
				r.dbExec(db, "GRANT ALL PRIVILEGES ON ALL TABLES IN SCHEMA public TO homer_user;")
				r.dbExec(db, "GRANT ALL PRIVILEGES ON ALL SEQUENCES IN SCHEMA public TO homer_user;")
				db.Close()
				break
			}
		}
	}
	return nil
}

func (r *Rotator) CreateDataTables(duration int) (err error) {
	suffix := replaceCreateDay(duration)
	if config.Setting.DBDriver == "mysql" {
		db, err := dbr.Open(config.Setting.DBDriver, config.Setting.DBUser+":"+config.Setting.DBPass+"@tcp("+r.addr[0]+":"+r.addr[1]+")/"+config.Setting.DBDataTable+"?"+url.QueryEscape("charset=utf8mb4&parseTime=true"), nil)
		if err != nil {
			return err
		}
		defer db.Close()
		r.dbExecFileMYSQL(db, r.box.String("mysql/tbldatalog.sql"), suffix, duration, r.logStep)
		r.dbExecFileMYSQL(db, r.box.String("mysql/tbldataqos.sql"), suffix, duration, r.qosStep)
		r.dbExecFileMYSQL(db, r.box.String("mysql/tbldatasip.sql"), suffix, duration, r.sipStep)
		r.dbExecPartitionFile(db, r.box.String("mysql/parlog.sql"), suffix, duration, r.logStep)
		r.dbExecPartitionFile(db, r.box.String("mysql/parqos.sql"), suffix, duration, r.qosStep)
		r.dbExecPartitionFile(db, r.box.String("mysql/parsip.sql"), suffix, duration, r.sipStep)
		r.dbExecFile(db, r.box.String("mysql/parmax.sql"), suffix)
	} else if config.Setting.DBDriver == "postgres" {
		db, err := dbr.Open(config.Setting.DBDriver, " host="+r.addr[0]+" port="+r.addr[1]+" dbname="+config.Setting.DBDataTable+" user="+config.Setting.DBUser+" password="+config.Setting.DBPass+" sslmode=disable", nil)
		if err != nil {
			return err
		}
		defer db.Close()
		r.dbExecFile(db, r.box.String("pgsql/tbldata.sql"), suffix)
		r.dbExecPartitionFile(db, r.box.String("pgsql/parlog.sql"), suffix, duration, r.logStep)
		r.dbExecPartitionFile(db, r.box.String("pgsql/parqos.sql"), suffix, duration, r.qosStep)
		r.dbExecPartitionFile(db, r.box.String("pgsql/parsip.sql"), suffix, duration, r.sipStep)
		r.dbExecPartitionFile(db, r.box.String("pgsql/idxlog.sql"), suffix, duration, r.logStep)
		r.dbExecPartitionFile(db, r.box.String("pgsql/idxqos.sql"), suffix, duration, r.qosStep)
		r.dbExecPartitionFile(db, r.box.String("pgsql/idxsip.sql"), suffix, duration, r.sipStep)
	}
	return nil
}

func (r *Rotator) CreateConfTables(duration int) (err error) {
	suffix := replaceCreateDay(duration)
	if config.Setting.DBDriver == "mysql" {
		db, err := dbr.Open(config.Setting.DBDriver, config.Setting.DBUser+":"+config.Setting.DBPass+"@tcp("+r.addr[0]+":"+r.addr[1]+")/"+config.Setting.DBConfTable+"?"+url.QueryEscape("charset=utf8mb4&parseTime=true"), nil)
		if err != nil {
			return err
		}
		defer db.Close()
		r.dbExecFile(db, r.box.String("mysql/tblconf.sql"), suffix)
		r.dbExecFile(db, r.box.String("mysql/insconf.sql"), suffix)
	} else if config.Setting.DBDriver == "postgres" {
		db, err := dbr.Open(config.Setting.DBDriver, " host="+r.addr[0]+" port="+r.addr[1]+" dbname="+config.Setting.DBConfTable+" user="+config.Setting.DBUser+" password="+config.Setting.DBPass+" sslmode=disable", nil)
		if err != nil {
			return err
		}
		defer db.Close()
		r.dbExec(db, "CREATE EXTENSION pgcrypto;")
		r.dbExecFile(db, r.box.String("pgsql/tblconf.sql"), suffix)
		r.dbExecFile(db, r.box.String("pgsql/idxconf.sql"), suffix)
		r.dbExecFile(db, r.box.String("pgsql/insconf.sql"), suffix)
	}
	return nil
}

func (r *Rotator) DropTables(duration int) (err error) {
	suffix := replaceDropDay(duration)
	if config.Setting.DBDriver == "mysql" {
		db, err := dbr.Open(config.Setting.DBDriver, config.Setting.DBUser+":"+config.Setting.DBPass+"@tcp("+r.addr[0]+":"+r.addr[1]+")/"+config.Setting.DBDataTable+"?"+url.QueryEscape("charset=utf8mb4&parseTime=true"), nil)
		if err != nil {
			return err
		}
		defer db.Close()
		r.dbExecFile(db, r.box.String("mysql/droptbl.sql"), suffix)
	} else if config.Setting.DBDriver == "postgres" {
		db, err := dbr.Open(config.Setting.DBDriver, " host="+r.addr[0]+" port="+r.addr[1]+" dbname="+config.Setting.DBDataTable+" user="+config.Setting.DBUser+" password="+config.Setting.DBPass+" sslmode=disable", nil)
		if err != nil {
			return err
		}
		defer db.Close()
		r.dbExecPartitionFile(db, r.box.String("pgsql/droplog.sql"), suffix, duration, r.logStep)
		r.dbExecPartitionFile(db, r.box.String("pgsql/dropqos.sql"), suffix, duration, r.qosStep)
		r.dbExecPartitionFile(db, r.box.String("pgsql/dropsip.sql"), suffix, duration, r.sipStep)
	}
	return nil
}

func (r *Rotator) dbExecFileMYSQL(db *dbr.Connection, file string, pattern strings.Replacer, d, p int) {
	t := time.Now().Add(time.Hour * time.Duration(24*d))
	tt := time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, t.Location())
	newMinTime := tt.Format("1504")
	newEndTime := tt.Add(time.Minute * time.Duration(p)).Format("2006-01-02 15:04:05")

	dot, err := dotsql.LoadFromString(pattern.Replace(file))
	if err != nil {
		logp.Err("%s\n\n", err)
	}

	for _, query := range dot.QueryMap() {
		query = strings.Replace(query, partitionMinTime, newMinTime, -1)
		query = strings.Replace(query, partitionEndTime, newEndTime, -1)

		logp.Debug("rotator", "db query:\n%s\n\n", query)
		_, err := db.Exec(query)
		checkDBErr(err)
	}
}

func (r *Rotator) dbExecFile(db *dbr.Connection, file string, pattern strings.Replacer) {
	dot, err := dotsql.LoadFromString(pattern.Replace(file))
	if err != nil {
		logp.Err("%s\n\n", err)
	}

	for _, query := range dot.QueryMap() {
		logp.Debug("rotator", "db query:\n%s\n\n", query)
		_, err := db.Exec(query)
		checkDBErr(err)
	}
}

func (r *Rotator) dbExecPartitionFile(db *dbr.Connection, file string, pattern strings.Replacer, d, p int) {
	dot, err := dotsql.LoadFromString(pattern.Replace(file))
	if err != nil {
		logp.Err("%s\n\n", err)
	}

	for _, q := range dot.QueryMap() {
		rotatePartitions(db, q, d, p)
	}
}

func (r *Rotator) dbExec(db *dbr.Connection, query string) {
	_, err := db.Exec(query)
	checkDBErr(err)
}

func (r *Rotator) Rotate() (err error) {
	r.createTables()
	createJob := cron.New()

	logp.Info("Start daily create data table job at 03:15:00\n")
	createJob.AddFunc("0 15 03 * * *", func() {
		if err := r.CreateDataTables(1); err != nil {
			logp.Err("%v", err)
		}
		if err := r.CreateDataTables(2); err != nil {
			logp.Err("%v", err)
		}
		logp.Info("Finished create data table job next will run at %v\n", time.Now().Add(time.Hour*24+1))
	})
	createJob.Start()

	if config.Setting.DBDropDays > 0 {
		dropJob := cron.New()
		logp.Info("Start daily drop data table job at 03:45:00\n")
		dropJob.AddFunc("0 45 03 * * *", func() {
			if err := r.DropTables(config.Setting.DBDropDays); err != nil {
				logp.Err("%v", err)
			}
			logp.Info("Finished drop data table job next will run at %v\n", time.Now().Add(time.Hour*24+1))
		})
		dropJob.Start()
	}
	return nil
}

func (r *Rotator) createTables() {
	if config.Setting.DBUser == "root" || config.Setting.DBUser == "admin" || config.Setting.DBUser == "postgres" {
		if err := r.CreateDatabases(); err != nil {
			logp.Err("%v", err)
		}
	}
	if err := r.CreateConfTables(0); err != nil {
		logp.Err("%v", err)
	}
	if err := r.CreateDataTables(0); err != nil {
		logp.Err("%v", err)
	}
	if err := r.CreateDataTables(1); err != nil {
		logp.Err("%v", err)
	}
	if config.Setting.DBDropOnStart && config.Setting.DBDropDays != 0 {
		if err := r.DropTables(config.Setting.DBDropDays); err != nil {
			logp.Err("%v", err)
		}
	}
}

func rotatePartitions(db *dbr.Connection, query string, d, p int) {
	var newStartTime, newEndTime, newPartTime string
	oriQuery := query

	t := time.Now().Add(time.Hour * time.Duration(24*d))
	startTime := time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, t.Location())
	endTime := time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, t.Location())

	for i := 0; i < 1440/p; i++ {
		query := oriQuery

		newPartTime = startTime.Add(time.Minute * time.Duration(i*p)).Format("1504")
		newStartTime = startTime.Add(time.Minute * time.Duration(i*p)).Format("2006-01-02 15:04:05")
		newEndTime = endTime.Add(time.Minute * time.Duration(i*p+p)).Format("2006-01-02 15:04:05")

		query = strings.Replace(query, partitionTime, newPartTime, -1)
		query = strings.Replace(query, partitionStartTime, newStartTime, -1)
		query = strings.Replace(query, partitionEndTime, newEndTime, -1)

		logp.Debug("rotator", "db query:\n%s\n\n", query)
		_, err := db.Exec(query)
		checkDBErr(err)
	}
}

func replaceCreateDay(d int) strings.Replacer {
	pn := time.Now().Add(time.Hour * time.Duration(24*d)).Format("20060102")
	return *strings.NewReplacer(
		"{{date}}", pn,
	)
}

func replaceDropDay(d int) strings.Replacer {
	pn := time.Now().Add(time.Hour * time.Duration(-24*d)).Format("20060102")
	return *strings.NewReplacer(
		"{{date}}", pn,
	)
}

func setStep(name string) (step int) {
	switch name {
	case "5m":
		step = 5
	case "10m":
		step = 10
	case "15m":
		step = 15
	case "20m":
		step = 20
	case "30m":
		step = 30
	case "45m":
		step = 45
	case "1h":
		step = 60
	case "2h":
		step = 120
	case "6h":
		step = 360
	case "12h":
		step = 720
	case "1d":
		step = 1440
	default:
		logp.Warn("Not allowed rotation step %s please use [1d, 12h, 6h, 2h, 1h, 30m, 20m, 15m, 10m, 5m]", name)
		step = 120
	}
	return
}

func checkDBErr(err error) {
	if err != nil {
		if config.Setting.DBDriver == "mysql" {
			if mErr, ok := err.(*mysql.MySQLError); ok && mErr.Number != 1062 && mErr.Number != 1481 && mErr.Number != 1517 {
				logp.Warn("%s\n\n", err)
			}
		} else if config.Setting.DBDriver == "postgres" {
			logp.Warn("%s\n\n", err)
		}
	}
}
