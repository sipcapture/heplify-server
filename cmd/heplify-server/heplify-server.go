package main

import (
	"fmt"
	"html/template"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"

	_ "net/http/pprof"

	"github.com/negbie/logp"
	"github.com/negbie/multiconfig"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/sipcapture/heplify-server/config"
	input "github.com/sipcapture/heplify-server/server"
)

type server interface {
	Run()
	End()
}

func init() {
	var err error
	var logging logp.Logging

	c := multiconfig.New()
	cfg := new(config.HeplifyServer)
	c.MustLoad(cfg)
	config.Setting = *cfg

	if tomlExists(config.Setting.Config) {
		cf := multiconfig.NewWithPath(config.Setting.Config)
		err := cf.Load(cfg)
		if err == nil {
			config.Setting = *cfg
		} else {
			fmt.Println("Syntax error in toml config file, use flag defaults.", err)
		}
	} else {
		fmt.Println("Could not find toml config file, use flag defaults.", err)
	}

	config.Setting.AlegIDs = config.GenerateRegexMap(config.Setting.AlegIDs)

	logp.DebugSelectorsStr = &config.Setting.LogDbg
	logp.ToStderr = &config.Setting.LogStd
	logging.Level = config.Setting.LogLvl
	if config.Setting.LogSys {
		logging.ToSyslog = &config.Setting.LogSys
	} else {
		var fileRotator logp.FileRotator
		fileRotator.Path = "./"
		fileRotator.Name = "heplify-server.log"
		logging.Files = &fileRotator
	}

	err = logp.Init("heplify-server", &logging)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func tomlExists(f string) bool {
	_, err := os.Stat(f)
	if os.IsNotExist(err) {
		return false
	} else if !strings.Contains(f, ".toml") {
		return false
	}
	return err == nil
}

func main() {
	var servers []server
	var wg sync.WaitGroup
	var sigCh = make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	if config.Setting.Version {
		fmt.Printf("VERSION: %s\r\n", config.Version)
		os.Exit(0)
	}

	startServer := func() {
		hep := input.NewHEPInput()
		servers = []server{hep}
		for _, srv := range servers {
			wg.Add(1)
			go func(s server) {
				defer wg.Done()
				s.Run()
			}(srv)
		}
	}
	endServer := func() {
		logp.Info("stopping heplify-server...")
		for _, srv := range servers {
			wg.Add(1)
			go func(s server) {
				defer wg.Done()
				s.End()
			}(srv)
		}
		wg.Wait()
		logp.Info("heplify-server has been stopped")
	}

	if len(config.Setting.ConfigHTTPAddr) > 2 {
		tmpl := template.Must(template.New("main").Parse(config.WebForm))
		http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
			if r.Method != http.MethodPost {
				tmpl.Execute(w, config.Setting)
				return
			}

			cfg, err := config.WebConfig(r)
			if err != nil {
				logp.Warn("Failed config reload from %v. %v", r.RemoteAddr, err)
				tmpl.Execute(w, config.Setting)
				return
			}
			logp.Info("Successful config reload from %v", r.RemoteAddr)
			endServer()
			config.Setting = *cfg
			logp.SetToSyslog(config.Setting.LogSys, "")
			tmpl.Execute(w, config.Setting)
			startServer()
		})

		go http.ListenAndServe(config.Setting.ConfigHTTPAddr, nil)
	}

	if promAddr := config.Setting.PromAddr; len(promAddr) > 2 {
		go func() {
			http.Handle("/metrics", promhttp.Handler())
			err := http.ListenAndServe(promAddr, nil)
			if err != nil {
				logp.Err("%v", err)
			}
		}()
	}

	startServer()
	<-sigCh
	endServer()
}
