package service

import (
	"context"
	"math/rand"
	"net/http"
	"sync"
	"time"

	"github.com/clarktrimble/delish/respond"
	"github.com/clarktrimble/hondo"
	"github.com/pkg/errors"
)

// Config holds service configurables.
type Config struct {
	Interval time.Duration `json:"interval" desc:"work period" default:"5s"`
}

// New creates a service from Config.
func (cfg *Config) New(rtr router, lgr logger) (svc *service) {

	svc = &service{
		interval: cfg.Interval,
		logger:   lgr,
	}

	rtr.HandleFunc("GET /report", svc.report)

	return
}

// Start starts the service.
func (svc *service) Start(ctx context.Context, wg *sync.WaitGroup) {

	if svc.started {
		err := errors.Errorf("cowardly refusing to start service again")
		svc.logger.Error(ctx, "failed to start worker", err)
		return
	}
	svc.started = true

	ctx = svc.logger.WithFields(ctx, "worker_id", hondo.Rand(7))
	svc.logger.Info(ctx, "worker starting", "name", "service")

	go svc.work(ctx, wg)
}

// unexported

type logger interface {
	Info(ctx context.Context, msg string, kv ...any)
	Error(ctx context.Context, msg string, err error, kv ...any)
	WithFields(ctx context.Context, kv ...interface{}) context.Context
}

type router interface {
	HandleFunc(pattern string, handler http.HandlerFunc)
}

type service struct {
	interval time.Duration
	count    int
	started  bool
	logger   logger
}

func (svc *service) report(writer http.ResponseWriter, request *http.Request) {

	rp := &respond.Respond{
		Writer: writer,
		Logger: svc.logger,
	}

	rp.WriteObjects(request.Context(), map[string]any{"worked": svc.count})
}

func (svc *service) work(ctx context.Context, wg *sync.WaitGroup) {

	wg.Add(1)
	defer wg.Done()

	ticker := time.NewTicker(svc.interval)

	for {
		select {
		case <-ticker.C:
			svc.entangle(ctx)
		case <-ctx.Done():
			svc.logger.Info(ctx, "worker shutting down")
			svc.disentangle()
			svc.logger.Info(ctx, "worker stopped")
			return
		}
	}
}

func (svc *service) entangle(ctx context.Context) {

	if rand.Intn(99) < 9 {
		err := errors.Errorf("oops")
		svc.logger.Error(ctx, "worker canna work", err)
		return
	}

	svc.logger.Info(ctx, "worker gonna work")

	time.Sleep(svc.interval / 3)
	svc.count++
}

func (svc *service) disentangle() {
	time.Sleep(svc.interval / 2)
}
