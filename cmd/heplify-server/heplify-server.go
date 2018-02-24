package main

import (
	"flag"
	"fmt"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/negbie/heplify-server/config"
	"github.com/negbie/heplify-server/logp"
	"github.com/negbie/heplify-server/server"
)

const version = "heplify-server 0.01"

type server interface {
	Run()
	End()
}

func init() {

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Use %s like: %s [option]\n", version, os.Args[0])
		flag.PrintDefaults()
	}

	var (
		logging     logp.Logging
		fileRotator logp.FileRotator
	)

	flag.StringVar(&logging.Level, "ll", "info", "Log level [debug, info, warning, error]")
	flag.StringVar(&fileRotator.Path, "fp", "./", "Log filepath")
	flag.StringVar(&fileRotator.Name, "fn", "heplify-server.log", "Log filename")
	flag.StringVar(&config.Setting.HEPAddr, "hs", "0.0.0.0:9060", "HEP server address")
	flag.IntVar(&config.Setting.HEPWorkers, "hw", 100, "HEP workers")
	flag.StringVar(&config.Setting.MQName, "qn", "nsq", "Message queue name")
	flag.StringVar(&config.Setting.MQAddr, "qs", "localhost:4015", "Message queue server address")
	flag.StringVar(&config.Setting.MQTopic, "qt", "hep", "Message queue topic name")
	flag.StringVar(&config.Setting.PromAddr, "ps", "0.0.0.0:8888", "Prometheus exposing address")
	flag.StringVar(&config.Setting.PromTarget, "pt", "", "Prometheus collecting targets")
	flag.StringVar(&config.Setting.DBDriver, "dd", "mysql", "Database driver [mysql, postgres]")
	flag.StringVar(&config.Setting.DBAddr, "ds", "localhost:3306", "Database server address")
	flag.IntVar(&config.Setting.DBBulk, "dk", 100, "Number of rows to insert at once")
	flag.StringVar(&config.Setting.DBName, "dn", "homer_data", "DB name")
	flag.StringVar(&config.Setting.DBUser, "du", "test", "DB user")
	flag.StringVar(&config.Setting.DBPassword, "dp", "test", "DB password")

	flag.Parse()

	logging.Files = &fileRotator
	config.Setting.Logging = &logging
}

func main() {
	var (
		wg    sync.WaitGroup
		sigCh = make(chan os.Signal, 1)
	)

	err := logp.Init("heplify-server", config.Setting.Logging)
	if err != nil {
		logp.Critical("%v", err)
		os.Exit(1)
	}

	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	hep := input.NewHEP()
	servers := []server{hep}

	for _, srv := range servers {
		wg.Add(1)
		go func(s server) {
			defer wg.Done()
			s.Run()
		}(srv)
	}

	<-sigCh

	for _, srv := range servers {
		wg.Add(1)
		go func(s server) {
			defer wg.Done()
			s.End()
		}(srv)
	}
	wg.Wait()
}
