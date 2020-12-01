package api

import (
	"context"
	"net/http"
	"time"

	"github.com/rendau/gms_temp/internal/interfaces"
)

type St struct {
	lg interfaces.Logger

	targetUri string

	server *http.Server
	lChan  chan error
}

func New(lg interfaces.Logger, listen, targetUri string) *St {
	api := &St{
		lg:        lg,
		targetUri: targetUri,
		lChan:     make(chan error, 1),
	}

	api.server = &http.Server{
		Addr:              listen,
		Handler:           api.router(),
		ReadTimeout:       2 * time.Minute,
		ReadHeaderTimeout: 10 * time.Second,
	}

	return api
}

func (a *St) Start() {
	go func() {
		err := a.server.ListenAndServe()
		if err != nil && err != http.ErrServerClosed {
			a.lg.Errorw("Http server closed", err)
			a.lChan <- err
		}
	}()
}

func (a *St) Wait() <-chan error {
	return a.lChan
}

func (a *St) Shutdown(ctx context.Context) error {
	return a.server.Shutdown(ctx)
}
