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
		lgr = &mock.LoggerMock{
			InfoFunc: func(ctx context.Context, msg string, kv ...any) {},
		}

		ctx = Initialize(context.Background(), &wg, lgr)
	})

	Describe("initializing the package", func() {

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

				svc := &testSvc{}
				go svc.Start(ctx, &wg, lgr)

				// once service is started, signal shutdown

				go func() {
					Eventually(svc.Started).Should(BeTrue())

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
				Eventually(ic).Should(HaveLen(6))
				Expect(ic()[0].Msg).To(Equal("starting up"))
				Expect(ic()[1].Msg).To(Equal("starting testSvc"))
				Expect(ic()[2].Msg).To(Equal("shutting down"))
				Expect(ic()[3].Msg).To(Equal("shutting down testSvc")) // <- triggered by cancel
				Expect(ic()[4].Msg).To(Equal("testSvc stopped"))       //
				Expect(ic()[5].Msg).To(Equal("stopped"))               // <- waitgroup'ed for this one!
			})
		})
	})

})

type testSvc struct {
	started bool
	mu      sync.RWMutex
}

func (svc *testSvc) Started() bool {
	svc.mu.RLock()
	defer svc.mu.RUnlock()
	return svc.started
}

func (svc *testSvc) Start(ctx context.Context, wg *sync.WaitGroup, lgr Logger) {

	lgr.Info(ctx, "starting testSvc")

	wg.Add(1)
	defer wg.Done()

	svc.mu.Lock()
	svc.started = true
	svc.mu.Unlock()

	<-ctx.Done()
	lgr.Info(ctx, "shutting down testSvc")

	// as if we're finishing something up ..
	time.Sleep(99 * time.Millisecond)

	lgr.Info(ctx, "testSvc stopped")
}
