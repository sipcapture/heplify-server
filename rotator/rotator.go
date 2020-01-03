package rotator

import (
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/go-sql-driver/mysql"
	"github.com/negbie/logp"
	"github.com/robfig/cron"
	"github.com/sipcapture/heplify-server/config"
	"github.com/sipcapture/heplify-server/database"
)

const (
	partitionDate      = "{{date}}"
	partitionTime      = "{{time}}"
	partitionMinTime   = "{{minTime}}"
	partitionStartTime = "{{startTime}}"
	partitionEndTime   = "{{endTime}}"
	partitionName      = "{{partName}}"
)

type Rotator struct {
	quit             chan bool
	user             string
	dataDB           string
	confDB           string
	driver           string
	rootDBAddr       string
	confDBAddr       string
	dataDBAddr       string
	partLog          int
	partIsup         int
	partQos          int
	partSip          int
	dropDays         int
	dropDaysCall     int
	dropDaysRegister int
	dropDaysDefault  int
	dropOnStart      bool
	createJob        *cron.Cron
	dropJob          *cron.Cron
}

func Setup(quit chan bool) *Rotator {
	r := &Rotator{
		quit:         quit,
		user:         config.Setting.DBUser,
		dataDB:       config.Setting.DBDataTable,
		confDB:       config.Setting.DBConfTable,
		driver:       config.Setting.DBDriver,
		partLog:      setStep(config.Setting.DBPartLog),
		partIsup:     setStep(config.Setting.DBPartIsup),
		partQos:      setStep(config.Setting.DBPartQos),
		partSip:      setStep(config.Setting.DBPartSip),
		dropDays:     config.Setting.DBDropDays,
		dropDaysCall: config.Setting.DBDropDaysCall,
		dropOnStart:  config.Setting.DBDropOnStart,
		createJob:    cron.New(),
		dropJob:      cron.New(),
	}

	r.rootDBAddr, _ = database.ConnectString("")
	r.confDBAddr, _ = database.ConnectString(config.Setting.DBConfTable)
	r.dataDBAddr, _ = database.ConnectString(config.Setting.DBDataTable)
	if r.dropDaysCall == 0 {
		r.dropDaysCall = r.dropDays
	}
	r.dropDaysRegister = config.Setting.DBDropDaysRegister
	if r.dropDaysRegister == 0 {
		r.dropDaysRegister = r.dropDays
	}
	r.dropDaysDefault = config.Setting.DBDropDaysDefault
	if r.dropDaysDefault == 0 {
		r.dropDaysDefault = r.dropDays
	}
	return r
}

func (r *Rotator) CreateDatabases() (err error) {
	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-r.quit:
			r.quit <- true
			return fmt.Errorf("stop database creation")
		case <-ticker.C:
			db, err := sql.Open(r.driver, r.rootDBAddr)
			if err = db.Ping(); err != nil {
				db.Close()
				logp.Err("%v", err)
				break
			}
			if r.driver == "mysql" {
				r.dbExec(db, "CREATE DATABASE IF NOT EXISTS "+r.dataDB+` DEFAULT CHARACTER SET = 'utf8mb4' DEFAULT COLLATE = 'utf8mb4_unicode_ci';`)
				r.dbExec(db, "CREATE DATABASE IF NOT EXISTS "+r.confDB+` DEFAULT CHARACTER SET = 'utf8mb4' DEFAULT COLLATE = 'utf8mb4_unicode_ci';`)
				r.dbExec(db, `CREATE USER IF NOT EXISTS 'homer_user'@'%' IDENTIFIED BY 'homer_password';`)
				r.dbExec(db, "GRANT ALL ON "+r.dataDB+`.* TO 'homer_user'@'%';`)
				r.dbExec(db, "GRANT ALL ON "+r.confDB+`.* TO 'homer_user'@'%';`)
				db.Close()
				return nil
			} else if r.driver == "postgres" {
				r.dbExec(db, "CREATE DATABASE "+r.dataDB)
				r.dbExec(db, "CREATE DATABASE "+r.confDB)
				r.dbExec(db, `CREATE USER homer_user WITH PASSWORD 'homer_password';`)
				r.dbExec(db, "GRANT postgres to homer_user;")
				r.dbExec(db, "GRANT ALL PRIVILEGES ON DATABASE "+r.dataDB+" TO homer_user;")
				r.dbExec(db, "GRANT ALL PRIVILEGES ON DATABASE "+r.confDB+" TO homer_user;")
				r.dbExec(db, "GRANT ALL PRIVILEGES ON ALL TABLES IN SCHEMA public TO homer_user;")
				r.dbExec(db, "GRANT ALL PRIVILEGES ON ALL SEQUENCES IN SCHEMA public TO homer_user;")
				db.Close()
				return nil
			}
		}
	}
}

func replaceDay(d int) *strings.Replacer {
	pd := time.Now().Add(time.Hour * time.Duration(24*d)).Format("20060102")
	return strings.NewReplacer(
		partitionDate, pd,
	)
}

func (r *Rotator) CreateDataTables(duration int) (err error) {
	db, err := sql.Open(r.driver, r.dataDBAddr)
	if err != nil {
		return err
	}
	defer db.Close()

	suffix := replaceDay(duration)
	if r.driver == "mysql" {
		// Set this connection to UTC time and create the partitions with it.
		r.dbExec(db, "SET time_zone = \"+00:00\";")
		if err := r.dbExecFile(db, tbldatalogmaria, suffix, duration, r.partLog); err == nil {
			r.dbExecFileLoop(db, parlogmaria, suffix, duration, r.partLog)
		}
		if err := r.dbExecFile(db, tbldataqosmaria, suffix, duration, r.partQos); err == nil {
			r.dbExecFileLoop(db, parqosmaria, suffix, duration, r.partQos)
		}
		if err := r.dbExecFile(db, tbldatasipmaria, suffix, duration, r.partSip); err == nil {
			r.dbExecFileLoop(db, parsipmaria, suffix, duration, r.partSip)
		}
		//r.dbExecFile(db, parmaxmaria, suffix, 0, 0)
	} else if r.driver == "postgres" {
		// Set this connection to UTC time and create the partitions with it.
		r.dbExec(db, "SET timezone = \"UTC\";")
		r.dbExecFile(db, tbldatapg, suffix, 0, 0)
		r.dbExecFileLoop(db, parlogpg, suffix, duration, r.partLog)
		r.dbExecFileLoop(db, parqospg, suffix, duration, r.partQos)
		r.dbExecFileLoop(db, parisuppg, suffix, duration, r.partIsup)
		r.dbExecFileLoop(db, parsippg, suffix, duration, r.partSip)
		r.dbExecFileLoop(db, idxlogpg, suffix, duration, r.partLog)
		r.dbExecFileLoop(db, idxisuppg, suffix, duration, r.partIsup)
		r.dbExecFileLoop(db, idxqospg, suffix, duration, r.partQos)
		r.dbExecFileLoop(db, idxsippg, suffix, duration, r.partSip)
	}
	return nil
}

func (r *Rotator) CreateConfTables(duration int) (err error) {
	db, err := sql.Open(r.driver, r.confDBAddr)
	if err != nil {
		return err
	}
	defer db.Close()

	suffix := replaceDay(duration)
	if r.driver == "mysql" {
		r.dbExecFile(db, tblconfmaria, suffix, 0, 0)
		r.dbExecFile(db, insconfmaria, suffix, 0, 0)
	} else if r.driver == "postgres" {
		r.dbExecFile(db, idxconfpg, suffix, 0, 0)
		r.dbExecFile(db, tblconfpg, suffix, 0, 0)
		r.dbExecFile(db, insconfpg, suffix, 0, 0)
	}
	return nil
}

func (r *Rotator) DropTables() (err error) {
	logp.Debug("rotator", "start drop tables (%v)\n", time.Now())
	db, err := sql.Open(r.driver, r.dataDBAddr)
	if err != nil {
		return err
	}
	defer db.Close()
	if r.driver == "mysql" {
		r.dbExecDropTables(db, listdroplogmaria, droplogmaria, r.dropDays)
		r.dbExecDropTables(db, listdropreportmaria, dropreportmaria, r.dropDays)
		r.dbExecDropTables(db, listdroprtcpmaria, droprtcpmaria, r.dropDays)
		r.dbExecDropTables(db, listdropcallmaria, dropcallmaria, r.dropDaysCall)
		r.dbExecDropTables(db, listdropregistermaria, dropregistermaria, r.dropDaysRegister)
		r.dbExecDropTables(db, listdropdefaultmaria, dropdefaultmaria, r.dropDaysDefault)
	} else if r.driver == "postgres" {
		r.dbExecDropTables(db, listdroplogpg, droplogpg, r.dropDays)
		r.dbExecDropTables(db, listdropisuppg, dropisuppg, r.dropDays)
		r.dbExecDropTables(db, listdropreportpg, dropreportpg, r.dropDays)
		r.dbExecDropTables(db, listdroprtcppg, droprtcppg, r.dropDays)
		r.dbExecDropTables(db, listdropcallpg, dropcallpg, r.dropDaysCall)
		r.dbExecDropTables(db, listdropregisterpg, dropregisterpg, r.dropDaysRegister)
		r.dbExecDropTables(db, listdropdefaultpg, dropdefaultpg, r.dropDaysDefault)
	}
	logp.Debug("rotator", "finished drop tables (%v)\n", time.Now())
	return nil
}

func (r *Rotator) dbExecDropTables(db * sql.DB, listfile []string, dropfile []string, d int) error {
	t := time.Now().Add(time.Hour * time.Duration(-24*(d-1)))
	t = time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, t.Location())
        partDate := t.Format("20060102")
	partTime := t.Format("1504")
        var rows *sql.Rows
        var lastErr error
	for _, listquery := range listfile {
		listquery = strings.Replace(listquery, partitionDate, partDate, -1)
		listquery = strings.Replace(listquery, partitionTime, partTime, -1)
		rows, lastErr = db.Query(listquery)
		if !checkDBErr(lastErr) {
			for rows.Next() {
				var partName string
				lastErr = rows.Scan(&partName)
				if !checkDBErr(lastErr) {
					for _, dropquery := range dropfile {
						dropquery = strings.Replace(dropquery, partitionName, partName, -1)
		                                logp.Debug("rotator", "db query:\n%s\n\n", dropquery)
						_, lastErr = db.Exec(dropquery)
						if checkDBErr(lastErr) {
							break;
						}
					}
				}
			}
			rows.Close()
		}
	}
	return lastErr
}

func (r *Rotator) dbExec(db *sql.DB, query string) {
	_, err := db.Exec(query)
	checkDBErr(err)
}

func (r *Rotator) dbExecFile(db *sql.DB, file []string, pattern *strings.Replacer, d, p int) error {
	t := time.Now().Add(time.Hour * time.Duration(24*d))
	tt := time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, t.Location())
	newMinTime := tt.Format("1504")
	newEndTime := tt.Add(time.Minute * time.Duration(p)).Format("2006-01-02 15:04:05")

	var lastErr error
	for _, query := range file {
		query = pattern.Replace(query)
		if p != 0 {
			query = strings.Replace(query, partitionMinTime, newMinTime, -1)
			query = strings.Replace(query, partitionEndTime, newEndTime, -1)
		}

		logp.Debug("rotator", "db query:\n%s\n\n", query)
		_, lastErr = db.Exec(query)
		checkDBErr(lastErr)
	}
	return lastErr
}

func (r *Rotator) dbExecFileLoop(db *sql.DB, file []string, pattern *strings.Replacer, d, p int) {
	for _, q := range file {
		q = pattern.Replace(q)
		fileLoop(db, q, d, p)
	}
}

func fileLoop(db *sql.DB, query string, d, p int) {
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

func (r *Rotator) Rotate() {
	r.createTables()
	r.createJob.AddFunc("0 30 03 * * *", func() {
		if err := r.CreateDataTables(1); err != nil {
			logp.Err("%v", err)
		}
		if err := r.CreateDataTables(2); err != nil {
			logp.Err("%v", err)
		}
		logp.Info("finished rotate job next will run at %v\n", time.Now().Add(time.Hour*24+1))
	})
	r.createJob.Start()

	if r.dropDays > 0 {
		r.dropJob.AddFunc("0 45 03 * * *", func() {
			if err := r.DropTables(); err != nil {
				logp.Err("%v", err)
			}
			logp.Info("finished drop job next will run at %v\n", time.Now().Add(time.Hour*24+1))
		})
		r.dropJob.Start()
	}
}

func (r *Rotator) End() {
	r.createJob.Stop()
	r.dropJob.Stop()
}

func (r *Rotator) createTables() {
	if r.user == "root" || r.user == "admin" || r.user == "postgres" {
		if err := r.CreateDatabases(); err != nil {
			logp.Info("%v", err)
			return
		}
	}
	logp.Info("start creating tables (%v)\n", time.Now())
	if err := r.CreateConfTables(0); err != nil {
		logp.Err("%v", err)
	}
	if err := r.CreateDataTables(-1); err != nil {
		logp.Err("%v", err)
	}
	if err := r.CreateDataTables(0); err != nil {
		logp.Err("%v", err)
	}
	if err := r.CreateDataTables(1); err != nil {
		logp.Err("%v", err)
	}
	logp.Info("end creating tables (%v)\n", time.Now())
	if r.dropOnStart && r.dropDays != 0 {
		if err := r.DropTables(); err != nil {
			logp.Err("%v", err)
		}
	}
	if r.dropDays == 0 {
		logp.Warn("don't schedule daily drop job because DBDropDays is 0\n")
		logp.Warn("set DBDropDays greater 0 or old data won't be deleted\n")
	}
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
	case "24h", "1d":
		step = 1440
	default:
		logp.Warn("Not allowed rotation step %s please use [1d, 12h, 6h, 2h, 1h, 30m, 20m, 15m, 10m, 5m]", name)
		step = 120
	}
	return
}

func checkDBErr(err error) bool {
	if err != nil {
		if mErr, ok := err.(*mysql.MySQLError); ok && (mErr.Number == 1050 ||
			mErr.Number == 1062 || mErr.Number == 1481 || mErr.Number == 1517) {
			logp.Debug("rotator", "%s\n\n", err)
		} else {
			logp.Warn("%s\n\n", err)
		}
		return true;
	} else {
		return false;
	}
}
