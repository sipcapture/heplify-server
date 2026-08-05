package database

import (
	"database/sql"
	"time"

	"github.com/negbie/logp"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	dbUp = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "heplify_db_up",
		Help: "Whether the database is reachable (1=up, 0=down)",
	}, []string{"driver"})

	dbInsertErrors = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "heplify_db_insert_errors_total",
		Help: "Total database insert/bulk errors",
	}, []string{"driver"})

	dbInsertSuccess = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "heplify_db_insert_success_total",
		Help: "Total successful database bulk inserts",
	}, []string{"driver"})

	dbLastSuccessUnix = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "heplify_db_last_insert_success_timestamp_seconds",
		Help: "Unix timestamp of the last successful bulk insert",
	}, []string{"driver"})
)

func startDBHealthMonitor(db *sql.DB, driver string) {
	if db == nil {
		return
	}
	dbUp.WithLabelValues(driver).Set(1)
	go func() {
		t := time.NewTicker(10 * time.Second)
		defer t.Stop()
		for range t.C {
			if err := db.Ping(); err != nil {
				dbUp.WithLabelValues(driver).Set(0)
				logp.Err("db health check failed: %v", err)
				continue
			}
			dbUp.WithLabelValues(driver).Set(1)
		}
	}()
}

func recordDBInsertError(driver string) {
	dbInsertErrors.WithLabelValues(driver).Inc()
	dbUp.WithLabelValues(driver).Set(0)
}

func recordDBInsertSuccess(driver string) {
	dbInsertSuccess.WithLabelValues(driver).Inc()
	dbLastSuccessUnix.WithLabelValues(driver).Set(float64(time.Now().Unix()))
	dbUp.WithLabelValues(driver).Set(1)
}
