package main

import (
	"fmt"
	"os"
	"os/signal"
	"sync"
	"syscall"
	//"net"
	//_ "net/http/pprof"

	"github.com/koding/multiconfig"
	"github.com/negbie/heplify-server/config"
	"github.com/negbie/heplify-server/logp"
	"github.com/negbie/heplify-server/server"
)

const version = "heplify-server 0.30"

type server interface {
	Run()
	End()
}

func init() {
	var logging logp.Logging
	var fileRotator logp.FileRotator

	c := multiconfig.New()
	cfg := new(config.HeplifyServer)
	c.MustLoad(cfg)
	config.Setting = *cfg

	cf := multiconfig.NewWithPath("heplify-server.toml")
	err := cf.Load(cfg)
	if err == nil {
		config.Setting = *cfg
	} else {
		fmt.Println("Could not find heplify-server.toml, use flag defaults")
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

func main() {
	var wg sync.WaitGroup
	var sigCh = make(chan os.Signal, 1)

	//go http.ListenAndServe(":8181", http.DefaultServeMux)

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
