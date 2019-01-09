package input

import (
	"context"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/negbie/heplify-server/config"
	"github.com/negbie/logp"
)

func (h *HEPInput) serveHTTP() {
	router := http.NewServeMux()
	router.Handle("/", index())

	server := &http.Server{
		Addr:         config.Setting.HTTPAddr,
		Handler:      router,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  15 * time.Second,
	}

	done := make(chan bool)
	go func() {
		<-h.quit
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		server.SetKeepAlivesEnabled(false)
		if err := server.Shutdown(ctx); err != nil {
			logp.Err("could not gracefully shutdown HTTP server: %v\n", err)
		}
		close(done)
	}()

	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		logp.Err("could not listen on %s: %v\n", config.Setting.HTTPAddr, err)
	}
	<-done
}

func index() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
			return
		}
		w.WriteHeader(http.StatusOK)
		body, err := ioutil.ReadAll(r.Body)
		if err != nil {
			logp.Err("%v", err)
		}
		_ = body
	})
}
