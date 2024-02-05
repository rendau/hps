package app

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"net/http/httputil"
	"os"
	"os/signal"
	"time"
)

type App struct {
	httpServer *http.Server

	exitCode int
}

func (a *App) Init() {
	//var err error

	// reverse proxy
	proxy := &httputil.ReverseProxy{
		Rewrite: func(r *httputil.ProxyRequest) {
			r.SetURL(conf.TargetUrl)
			r.Out.Header["X-Forwarded-For"] = r.In.Header["X-Forwarded-For"]
			r.SetXForwarded()
			if conf.TargetHost != "" {
				r.Out.Host = conf.TargetHost
			}
		},
	}

	// handler
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		startTime := time.Now()

		defer func() {
			if r.Context().Err() != nil {
				slog.Warn("request canceled", "method", r.Method, "url", r.URL.String(), "duration", time.Since(startTime).String())
			}
		}()

		if conf.TargetTimeout > 0 {
			ctx, cancel := context.WithTimeout(r.Context(), conf.TargetTimeout)
			defer cancel()
			r = r.WithContext(ctx)
		}

		proxy.ServeHTTP(w, r)
	})

	// server
	a.httpServer = &http.Server{
		Addr:              ":" + conf.HttpPort,
		Handler:           handler,
		ReadHeaderTimeout: 10 * time.Second,
		ReadTimeout:       2 * time.Minute,
		MaxHeaderBytes:    300 * 1024,
	}
}

func (a *App) Start() {
	slog.Info("Starting")

	// http server
	{
		go func() {
			err := a.httpServer.ListenAndServe()
			if err != nil && !errors.Is(err, http.ErrServerClosed) {
				errCheck(err, "http-server stopped")
			}
		}()
		slog.Info("http-server started " + a.httpServer.Addr)
	}
}

func (a *App) Listen() {
	signalCtx, signalCtxCancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer signalCtxCancel()

	// wait signal
	<-signalCtx.Done()
}

func (a *App) Stop() {
	slog.Info("Shutting down...")

	// http server
	{
		ctx, ctxCancel := context.WithTimeout(context.Background(), 15*time.Second)
		defer ctxCancel()

		if err := a.httpServer.Shutdown(ctx); err != nil {
			slog.Error("http-server shutdown error", "error", err)
			a.exitCode = 1
		}
	}
}

func (a *App) Exit() {
	slog.Info("Exit")

	os.Exit(a.exitCode)
}

func errCheck(err error, msg string) {
	if err != nil {
		if msg != "" {
			err = fmt.Errorf("%s: %w", msg, err)
		}
		slog.Error(err.Error())
		os.Exit(1)
	}
}
