// Package delish starts and stops an http server
// coordinating with other services
// with logging and timeouts
// and provides a json responder
package delish

import (
	"context"
	errs "errors"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/pkg/errors"
)

// Config is the server's configuration
type Config struct {
	Host    string        `json:"host" desc:"hostname or ip for which to bind"`
	Port    int           `json:"port" desc:"port on which to listen" required:"true"`
	Timeout time.Duration `json:"timeout" desc:"characteristic timeout" default:"10s"`
}

// Server represents a json api webserver
type Server struct {
	Addr    string
	Handler http.Handler
	Logger  Logger
	Timeout time.Duration
}

// New creates a server from config
func (cfg *Config) New(handler http.Handler, lgr Logger) (srv *Server) {

	srv = &Server{
		Addr:    fmt.Sprintf("%s:%d", cfg.Host, cfg.Port),
		Timeout: cfg.Timeout,
		Handler: handler,
		Logger:  lgr,
	}

	return
}

// Start creates an httpServer, starts it, and waits for context's cancel to be called
func (svr *Server) Start(ctx context.Context, wg *sync.WaitGroup) {

	svr.Logger.Info(ctx, "starting http service")

	httpServer := &http.Server{
		Addr:              svr.Addr,
		ReadHeaderTimeout: 3 * svr.Timeout,
		ReadTimeout:       6 * svr.Timeout,
		WriteTimeout:      9 * svr.Timeout,
		Handler:           svr.Handler,
	}

	go svr.work(ctx, httpServer)
	go svr.wait(ctx, httpServer, wg)
}

// ObjHandler is a convinience method that responds with a marshalled named object
func (svr *Server) ObjHandler(name string, obj any) http.HandlerFunc {

	return func(writer http.ResponseWriter, request *http.Request) {

		rp := &Respond{
			Writer: writer,
			Logger: svr.Logger,
		}

		rp.WriteObjects(request.Context(), map[string]any{name: obj})
	}
}

// unexported

func (svr *Server) work(ctx context.Context, httpServer *http.Server) {

	svr.Logger.Info(ctx, "listening", "address", svr.Addr)

	err := httpServer.ListenAndServe()
	if !errs.Is(err, http.ErrServerClosed) {
		err = errors.Wrapf(err, "failed to listen on: %s", svr.Addr)
		svr.Logger.Error(ctx, "service failed", err)
	}
}

func (svr *Server) wait(ctx context.Context, httpServer *http.Server, wg *sync.WaitGroup) {

	wg.Add(1)
	defer wg.Done()

	<-ctx.Done()
	svr.Logger.Info(ctx, "shutting down http service ..")

	err := httpServer.Shutdown(ctx)
	if err != nil {
		err = errors.Wrapf(err, "failed to shutdown on: %s", svr.Addr)
		svr.Logger.Error(ctx, "shutdown failed", err)
		return
	}
	svr.Logger.Info(ctx, "http service stopped")
}
