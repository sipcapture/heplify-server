package main

import (
	"bytes"
	"fmt"
	"html/template"
	"io/ioutil"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"

	"github.com/koding/multiconfig"
	"github.com/negbie/heplify-server/config"
	input "github.com/negbie/heplify-server/server"
	"github.com/negbie/logp"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

type server interface {
	Run()
	End()
}

func init() {
	var err error
	var logging logp.Logging
	var fileRotator logp.FileRotator

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

	logp.DebugSelectorsStr = &config.Setting.LogDbg
	logging.Level = config.Setting.LogLvl
	logp.ToStderr = &config.Setting.LogStd
	fileRotator.Path = "./"
	fileRotator.Name = "heplify-server.log"
	logging.Files = &fileRotator

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

	/* 	autopprof.Capture(autopprof.CPUProfile{
		Duration: 15 * time.Second,
	}) */

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
		for _, srv := range servers {
			wg.Add(1)
			go func(s server) {
				defer wg.Done()
				s.End()
			}(srv)
		}
		wg.Wait()
	}

	if len(config.Setting.ConfigHTTPAddr) > 2 {
		cfgFile := config.Setting.Config
		tmpl := template.Must(template.New("main").Parse(configForm))
		http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
			if r.Method != http.MethodPost {
				tmpl.Execute(w, nil)
				return
			}

			fv := []byte(r.FormValue("config"))
			if len(fv) < 6 || !bytes.Contains(fv, []byte("=")) {
				tmpl.Execute(w, struct {
					Success bool
					Err     error
				}{false, fmt.Errorf("Wrong format or too short! Please see https://github.com/sipcapture/heplify-server/tree/master/example")})
				return
			}

			ioutil.WriteFile(cfgFile, fv, 0644)
			cf := multiconfig.NewWithPath(cfgFile)
			cfg := new(config.HeplifyServer)
			err := cf.Load(cfg)
			if err != nil {
				logp.Warn("Failed config reload from %v. %v", r.RemoteAddr, err)
				tmpl.Execute(w, struct {
					Success bool
					Err     error
				}{false, err})
				return
			}
			logp.Info("Successfull config reloaded from %v", r.RemoteAddr)
			endServer()
			config.Setting = *cfg
			startServer()
			tmpl.Execute(w, struct {
				Success bool
				Err     error
			}{true, nil})

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

var configForm = `
<!DOCTYPE html>
<html>
    <head>
		<title>heplify-server web config</title>
    </head>
    <body>
        <h2>heplify-server.toml</h2>
		{{if .Success}}
			<h4>Success!</h4>
		{{end}}
		{{if not .Success}}
			<h4>{{.Err}}</h4>
		{{end}}
		<form method="POST">
			<label>Config:</label><br />
			<textarea type="text" name="config" style="font-family: Arial;font-size: 10pt;width:80%;height:25vw"></textarea><br />
			<input type="submit" value="Apply config" />
		</form>
    </body>
</html>
`
