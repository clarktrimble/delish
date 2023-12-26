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

	"github.com/clarktrimble/delish/mock"
)

func TestMid(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Graceful Suite")
}

var _ = Describe("Graceful", func() {
	var (
		ctx context.Context
		lgr *mock.LoggerMock
		wg  sync.WaitGroup
	)

	BeforeEach(func() {
		ctx = context.Background()
		lgr = &mock.LoggerMock{
			InfoFunc: func(ctx context.Context, msg string, kv ...any) {},
		}
	})

	Describe("initializing the package", func() {

		JustBeforeEach(func() {
			ctx = Initialize(ctx, &wg, lgr)
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

		When("all is well", func() {
			BeforeEach(func() {

				// init graceful and start test service

				ctxCancel := Initialize(context.Background(), &wg, lgr)
				go testSvc{}.Start(ctxCancel, &wg, lgr)

				// once service is started, signal shutdown

				go func() {
					ic := lgr.InfoCalls
					Eventually(ic).Should(HaveLen(1))
					Expect(ic()[0].Msg).To(Equal("starting testSvc"))

					proc, err := os.FindProcess(os.Getpid())
					Expect(err).ToNot(HaveOccurred())
					err = proc.Signal(syscall.SIGQUIT)
					Expect(err).ToNot(HaveOccurred())
				}()

				// block with graceful, waiting for signal

				Wait(ctx)
			})

			It("starts, blocks, cancels, waits, and stops", func() {
				ic := lgr.InfoCalls
				Eventually(ic).Should(HaveLen(5))
				Expect(ic()[1].Msg).To(Equal("shutting down"))
				Expect(ic()[2].Msg).To(Equal("shutting down testSvc")) // <- triggered by cancel
				Expect(ic()[3].Msg).To(Equal("testSvc stopped"))       //
				Expect(ic()[4].Msg).To(Equal("stopped"))               // <- waitgroup'ed for this one!
			})
		})
	})

})

type testSvc struct{}

func (svc testSvc) Start(ctx context.Context, wg *sync.WaitGroup, lgr Logger) {

	lgr.Info(ctx, "starting testSvc")

	wg.Add(1)
	defer wg.Done()

	<-ctx.Done()
	lgr.Info(ctx, "shutting down testSvc")

	// as if we're finishing something up ..
	time.Sleep(99 * time.Millisecond)

	lgr.Info(ctx, "testSvc stopped")
}
