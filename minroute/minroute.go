package minroute

import (
	"context"
	"net/http"
	"strings"

	"github.com/clarktrimble/delish/respond"
	"github.com/pkg/errors"
)

//go:generate moq -out mock_test.go . Logger

// Logger specifies a logging interface.
type Logger interface {
	Info(ctx context.Context, msg string, kv ...any)
	Error(ctx context.Context, msg string, err error, kv ...any)
	WithFields(ctx context.Context, kv ...any) context.Context
}

// MinRoute maps http methods and paths to handlers
type MinRoute struct {
	Ctx    context.Context
	Logger Logger
	Routes map[string]map[string]http.HandlerFunc
}

// New creates a router with an empty route table.
//
// And stashes a copy of context, which is usually a no-no.
// But, breaking the rules here, as it allows for sensisibly contextual
// error logging when something goes wrong setting a route.
func New(ctx context.Context, lgr Logger) (rtr *MinRoute) {

	rtr = &MinRoute{
		Ctx:    ctx,
		Logger: lgr,
		Routes: map[string]map[string]http.HandlerFunc{
			"GET":    map[string]http.HandlerFunc{},
			"PUT":    map[string]http.HandlerFunc{},
			"POST":   map[string]http.HandlerFunc{},
			"DELETE": map[string]http.HandlerFunc{},
		},
	}

	return
}

// ServeHTTP looks up the handler and invokes
func (rtr *MinRoute) ServeHTTP(writer http.ResponseWriter, request *http.Request) {

	routes, ok := rtr.Routes[request.Method]
	if !ok {
		notFound(request.Context(), writer, rtr.Logger)
		return
	}

	if request.URL == nil {
		notFound(request.Context(), writer, rtr.Logger)
		return
	}

	handler, ok := routes[request.URL.Path]
	if !ok {
		notFound(request.Context(), writer, rtr.Logger)
		return
	}

	handler.ServeHTTP(writer, request)
}

// Set associates a method and path with a handler
func (rtr *MinRoute) HandleFunc(pattern string, handler http.HandlerFunc) {

	split := strings.Split(pattern, " ")
	if len(split) != 2 {
		err := errors.Errorf("failed to split pattern: '%s' into method and path", pattern)
		rtr.Logger.Error(rtr.Ctx, "unable to set route", err)
		return
	}
	method := split[0]
	path := split[1]

	_, ok := rtr.Routes[method]
	if !ok {
		err := errors.Errorf("unsupported method from pattern: '%s'", pattern)
		rtr.Logger.Error(rtr.Ctx, "unable to set route", err)
		return
	}

	rtr.Routes[method][path] = handler
}

// unexported

func notFound(ctx context.Context, writer http.ResponseWriter, lgr Logger) {

	rp := &respond.Respond{
		Writer: writer,
		Logger: lgr,
	}

	rp.NotFound(ctx)
}
