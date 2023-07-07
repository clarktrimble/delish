package graceful

// Note: sneaking into package here for test of "singleton"

import (
	"context"
	"os"
	"sync"
	"syscall"
	"testing"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/clarktrimble/delish/test/mock"
)

func TestMid(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Graceful Suite")
}

var _ = Describe("Graceful", func() {
	var (
		ctx context.Context
		lgr *mock.Logger
		wg  sync.WaitGroup
		to  time.Duration
	)

	BeforeEach(func() {
		ctx = context.Background()
		lgr = mock.NewLogger()
		to = 10 * time.Second
	})

	Describe("initializing the package", func() {

		JustBeforeEach(func() {
			ctx = Initialize(ctx, &wg, to, lgr)
		})

		When("all is well", func() {

			It("populates the object", func() {
				Expect(graceful.WaitGroup).ToNot(BeNil())
				Expect(graceful.Cancel).ToNot(BeNil())
				Expect(graceful.Logger).ToNot(BeNil())
			})

		})
	})

	Describe("waiting for an interrupt", func() {

		JustBeforeEach(func() {
			go func() {
				proc, err := os.FindProcess(os.Getpid())
				Expect(err).ToNot(HaveOccurred())

				// interrupt after pausing for Wait
				time.Sleep(99 * time.Millisecond)
				_ = proc.Signal(syscall.SIGQUIT)
			}()

			Wait(ctx)
		})

		When("all is well", func() {
			BeforeEach(func() {
				ctxCancel := Initialize(context.Background(), &wg, to, lgr)
				go testSvc{}.Serve(ctxCancel, &wg, lgr)
			})

			It("starts, blocks, cancels, waits, and stops", func() {
				Expect(lgr.Logged).To(HaveLen(5))
				Expect(lgr.Logged[0]["msg"]).To(Equal("starting testSvc"))
				Expect(lgr.Logged[1]["msg"]).To(Equal("shutting down .."))
				Expect(lgr.Logged[2]["msg"]).To(Equal("shutting down testSvc")) // <- triggered by cancel
				Expect(lgr.Logged[3]["msg"]).To(Equal("testSvc stopped"))       // <- waitgroup'ed for this one!
				Expect(lgr.Logged[4]["msg"]).To(Equal("stopped"))
			})
		})

	})

})

type testSvc struct{}

func (svc testSvc) Serve(ctx context.Context, wg *sync.WaitGroup, lgr Logger) {

	lgr.Info(ctx, "starting testSvc")

	wg.Add(1)
	defer wg.Done()

	<-ctx.Done()
	lgr.Info(ctx, "shutting down testSvc")

	// as if we're doing stuff ..
	time.Sleep(99 * time.Millisecond)

	lgr.Info(ctx, "testSvc stopped")
}
