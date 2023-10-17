// Package graceful provides for a graceful shutdown among goroutines
// when interrupted by a signal
package graceful

import (
	"context"
	"os"
	"os/signal"
	"sync"
	"syscall"
)

// Graceful is for a graceful shutdown
type Graceful struct {
	WaitGroup *sync.WaitGroup
	Cancel    context.CancelFunc
	Logger    Logger
}

// New creates a graceful
func Initialize(ctx context.Context, wg *sync.WaitGroup, lgr Logger) context.Context {

	ctx, cancel := context.WithCancel(ctx)

	graceful = &Graceful{
		Cancel:    cancel,
		WaitGroup: wg,
		Logger:    lgr,
	}

	return ctx
}

// Wait blocks until interrupted, cancels ctx, waits for group, and exits
func Wait(ctx context.Context) {

	// wait for interrupt

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, stop...)
	<-sigChan

	graceful.Logger.Info(ctx, "shutting down ..")

	// when cancel is called other routines blocking on ctx.Done can proceed with shutdown
	// wait for them to finish via the wait group ... and we're done!

	graceful.Cancel()
	graceful.WaitGroup.Wait()

	graceful.Logger.Info(ctx, "stopped")
}

// unexported

var (
	stop     []os.Signal = []os.Signal{syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT, os.Interrupt}
	graceful *Graceful
)
