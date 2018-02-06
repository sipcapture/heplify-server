package main

import (
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/negbie/heplify-server/input"
)

type server interface {
	Run()
	End()
}

func main() {
	var (
		wg    sync.WaitGroup
		sigCh = make(chan os.Signal, 1)
	)

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
