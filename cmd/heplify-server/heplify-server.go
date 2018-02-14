package main

import (
	"flag"
	"fmt"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/negbie/heplify-server/config"
	"github.com/negbie/heplify-server/input"
	"github.com/negbie/heplify-server/logp"
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

	flag.StringVar(&logging.Level, "l", "info", "Log level [debug, info, warning, error]")
	flag.StringVar(&fileRotator.Path, "p", "./", "Log filepath")
	flag.StringVar(&fileRotator.Name, "n", "heplify-server.log", "Log filename")
	flag.StringVar(&config.Setting.HEPAddr, "hs", "localhost:9060", "HEP server address")
	flag.IntVar(&config.Setting.HEPWorkers, "hw", 100, "HEP workers")
	flag.StringVar(&config.Setting.HEPTopic, "ht", "hep", "HEP topic name")
	flag.StringVar(&config.Setting.NSQName, "qn", "nsq", "output name")
	flag.StringVar(&config.Setting.NSQAddr, "ns", "localhost:4015", "NSQ server address")
	flag.StringVar(&config.Setting.DBDriver, "dr", "mysql", "Database driver [mysql, postgres]")
	flag.StringVar(&config.Setting.DBAddr, "ds", "localhost", "Database server address")
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
