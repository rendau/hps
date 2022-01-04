package cmd

import (
	"context"
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/rendau/hps/internal/adapters/logger/zap"
	"github.com/rendau/hps/internal/domain/api"
	"github.com/spf13/viper"
)

func Execute() {
	var err error

	loadConf()

	app := struct {
		lg  *zap.St
		api *api.St
	}{}

	app.lg, err = zap.New(viper.GetString("log_level"), viper.GetBool("debug"), false)
	if err != nil {
		log.Fatal(err)
	}

	app.api = api.New(
		app.lg,
		viper.GetString("http_listen"),
		viper.GetString("target_uri"),
	)

	app.lg.Infow(
		"Starting",
		"http_listen", viper.GetString("http_listen"),
	)

	app.api.Start()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM, syscall.SIGINT)

	var exitCode int

	select {
	case <-stop:
	case <-app.api.Wait():
		exitCode = 1
	}

	app.lg.Infow("Shutting down...")

	ctx, ctxCancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer ctxCancel()

	err = app.api.Shutdown(ctx)
	if err != nil {
		app.lg.Errorw("Fail to shutdown http-api", err)
		exitCode = 1
	}

	os.Exit(exitCode)
}

func loadConf() {
	viper.SetDefault("debug", "false")
	viper.SetDefault("http_listen", ":80")
	viper.SetDefault("log_level", "debug")

	viper.AutomaticEnv()

	// viper.Set("some.url", uriRPadSlash(viper.GetString("some.url")))
	viper.Set("target_uri", strings.TrimSuffix(viper.GetString("target_uri"), "/"))
}

func uriRPadSlash(uri string) string {
	if uri != "" && !strings.HasSuffix(uri, "/") {
		return uri + "/"
	}
	return uri
}
