package app

import (
	"net/url"
	"time"

	"github.com/caarlos0/env/v9"
)

var conf = struct {
	Debug            bool          `env:"DEBUG" envDefault:"false"`
	HttpPort         string        `env:"HTTP_PORT" envDefault:"80"`
	HttpCors         bool          `env:"HTTP_CORS" envDefault:"false"`
	TargetUrlStr     string        `env:"TARGET_URL"`
	TargetUrl        *url.URL      `env:"-"`
	TargetHost       string        `env:"TARGET_HOST"`
	TargetTimeoutStr string        `env:"TARGET_TIMEOUT"`
	TargetTimeout    time.Duration `env:"-"`
}{}

func init() {
	var err error

	if err := env.Parse(&conf); err != nil {
		panic(err)
	}

	if conf.TargetUrlStr == "" {
		panic("TARGET_URL is required")
	}

	conf.TargetUrl, err = url.Parse(conf.TargetUrlStr)
	if err != nil {
		panic(err)
	}

	if conf.TargetTimeoutStr != "" {
		conf.TargetTimeout, err = time.ParseDuration(conf.TargetTimeoutStr)
		if err != nil {
			panic(err)
		}
	}
}
