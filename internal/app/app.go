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

const (
	HealthcheckPort = "3003"
)

type App struct {
	httpServer        *http.Server
	healthcheckServer *http.Server

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
		if conf.Debug {
			startTime := time.Now()

			defer func() {
				if r.Context().Err() != nil {
					slog.Warn("request canceled", "method", r.Method, "url", r.URL.String(), "duration", time.Since(startTime).String())
				}
			}()
		}

		if conf.TargetTimeout > 0 {
			ctx, cancel := context.WithTimeout(r.Context(), conf.TargetTimeout)
			defer cancel()
			r = r.WithContext(ctx)
		}

		proxy.ServeHTTP(w, r)
	})

	// http server
	a.httpServer = &http.Server{
		Addr:              ":" + conf.HttpPort,
		Handler:           handler,
		ReadHeaderTimeout: 10 * time.Second,
		ReadTimeout:       2 * time.Minute,
		MaxHeaderBytes:    300 * 1024,
	}

	// healthcheck server
	a.healthcheckServer = &http.Server{
		Addr: ":" + HealthcheckPort,
		Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		}),
		ReadHeaderTimeout: 3 * time.Second,
		ReadTimeout:       3 * time.Minute,
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

	// healthcheck server
	{
		go func() {
			err := a.healthcheckServer.ListenAndServe()
			if err != nil && !errors.Is(err, http.ErrServerClosed) {
				errCheck(err, "healthcheck-server stopped")
			}
		}()
		slog.Info("healthcheck-server started " + a.healthcheckServer.Addr)
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

	// healthcheck server
	{
		ctx, ctxCancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer ctxCancel()

		if err := a.healthcheckServer.Shutdown(ctx); err != nil {
			slog.Error("healthcheck-server shutdown error", "error", err)
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
