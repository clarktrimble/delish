package main

import (
	"context"
	"sync"
	"time"

	"github.com/clarktrimble/delish"
	"github.com/clarktrimble/hondo"

	"github.com/clarktrimble/delish/graceful"
	"github.com/clarktrimble/delish/minroute"

	"github.com/clarktrimble/delish/examples/api/minlog"
	"github.com/clarktrimble/delish/examples/api/service"
)

var (
	version string
	wg      sync.WaitGroup
)

type Config struct {
	Version string          `json:"version" ignored:"true"`
	Server  *delish.Config  `json:"server"`
	Service *service.Config `json:"service"`
}

func main() {

	// using a literal config to avoid dep in example
	// see github.com/clarktrimble/launch for envconfig convenience

	cfg := &Config{
		Version: version,
		Server: &delish.Config{
			Port:    8088,
			Timeout: 10 * time.Second,
		},
		Service: &service.Config{
			Interval: 5 * time.Second,
		},
	}

	// setup logger and initialize graceful

	lgr := &minlog.MinLog{}
	ctx := lgr.WithFields(context.Background(),
		"app_id", "api_demo",
		"run_id", hondo.Rand(7),
	)

	ctx = graceful.Initialize(ctx, &wg, lgr, "config", cfg)

	// setup router

	rtr := minroute.New(ctx, lgr)
	rtr.HandleFunc("GET /config", delish.ObjHandler("config", cfg, lgr))

	// start demo service

	svc := cfg.Service.New(rtr, lgr)
	svc.Start(ctx, &wg)

	// start api server and wait for interrupt

	svr := cfg.Server.NewWithLog(ctx, rtr, lgr)
	svr.Start(ctx, &wg)
	graceful.Wait(ctx)

	// delicious!
}
