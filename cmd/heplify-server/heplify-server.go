package main

import (
	"os"
	"os/signal"
	"sync"
	"syscall"
)

type protos interface {
	run()
	end()
}

func main() {
	var (
		wg    sync.WaitGroup
		sigCh = make(chan os.Signal, 1)
	)

	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	hep := protos.NewHEP()
	protocols := []protos{hep}

	for _, p := range protocols {
		wg.Add(1)
		go func(p protos) {
			defer wg.Done()
			p.run()
		}(p)
	}

	<-sigCh

	for _, p := range protocols {
		wg.Add(1)
		go func(p protos) {
			defer wg.Done()
			p.end()
		}(p)
	}
	wg.Wait()
}
