package main

import (
	"context"
	"sync"
	"time"

	"github.com/clarktrimble/delish"
	"github.com/clarktrimble/hondo"

	"github.com/clarktrimble/delish/graceful"

	"github.com/clarktrimble/delish/examples/api/demosvc"
	"github.com/clarktrimble/delish/examples/api/minlog"
	"github.com/clarktrimble/delish/examples/api/minroute"
)

var (
	version string
	wg      sync.WaitGroup
)

type Config struct {
	Version string         `json:"version" ignored:"true"`
	Server  *delish.Config `json:"server"`
}

// Todo: demo another goroutine

func main() {

	// usually load config with envconfig, but literal for demo

	cfg := &Config{
		Version: version,
		Server: &delish.Config{
			Port:    8088,
			Timeout: 10 * time.Second,
		},
	}

	// create logger and initialize graceful

	lgr := &minlog.MinLog{}
	ctx := lgr.WithFields(context.Background(), "run_id", hondo.Rand(7))

	ctx = graceful.Initialize(ctx, &wg, lgr)

	// create router/handler, and server

	rtr := minroute.New(lgr)
	svr := cfg.Server.NewWithLog(ctx, rtr, lgr)

	// register route directly
	// or via service layer

	rtr.Set("GET", "/config", delish.ObjHandler("config", cfg, lgr))
	demosvc.AddRoute(svr, rtr)

	// delicious!

	svr.Start(ctx, &wg)
	graceful.Wait(ctx)
}
