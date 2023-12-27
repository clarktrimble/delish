package delish_test

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	. "github.com/clarktrimble/delish"
	"github.com/clarktrimble/delish/mock"
)

func TestDelish(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Delish Suite")
}

var _ = Describe("Delish", func() {
	var (
		handler http.Handler
		lgr     *mock.LoggerMock
		svr     *Server
		cfg     *Config
	)

	BeforeEach(func() {
		lgr = &mock.LoggerMock{
			InfoFunc:  func(ctx context.Context, msg string, kv ...any) {},
			ErrorFunc: func(ctx context.Context, msg string, err error, kv ...any) {},
		}

		cfg = &Config{
			Port:    8083,
			Timeout: 33 * time.Second,
		}
	})

	Describe("creating a server", func() {

		BeforeEach(func() {
			handler = http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {})
		})

		Describe("with no frills", func() {

			JustBeforeEach(func() {
				svr = cfg.New(handler, lgr)
			})

			When("all goes well", func() {

				It("creates a well formed server", func() {
					Expect(svr.Addr).To(Equal(":8083"))
					Expect(svr.Handler).ToNot(BeNil())
					Expect(svr.Logger).To(Equal(lgr))
					Expect(svr.Timeout).To(Equal(33 * time.Second))
				})
			})
		})

		Describe("with request/response logging", func() {
			var (
				ctx context.Context
			)

			JustBeforeEach(func() {
				svr = cfg.NewWithLog(ctx, handler, lgr)
			})

			When("all goes well", func() {
				BeforeEach(func() {
					ctx = context.Background()
				})

				It("creates a well formed server", func() {
					Expect(svr.Addr).To(Equal(":8083"))
					Expect(svr.Handler).ToNot(BeNil())
					Expect(svr.Logger).To(Equal(lgr))
					Expect(svr.Timeout).To(Equal(33 * time.Second))
				})
			})
		})
	})

	Describe("starting a server", func() {
		var (
			ctx    context.Context
			wg     sync.WaitGroup
			cancel context.CancelFunc
		)

		When("all goes well", func() {
			BeforeEach(func() {
				ctx, cancel = context.WithCancel(context.Background())
				handler = http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
					fmt.Fprint(writer, `{"ima": "pc"}`)
				})

				svr = cfg.New(handler, lgr)
				svr.Start(ctx, &wg)
			})

			//It("starts, serves, and stops", MustPassRepeatedly(33), func() {
			It("starts, serves, and stops", func() {

				// check for startup

				ic := lgr.InfoCalls
				Eventually(ic).Should(HaveLen(2))
				Expect(ic()[0].Msg).To(Equal("starting http service"))
				Expect(ic()[1].Msg).To(Equal("listening"))

				// make a request

				time.Sleep(9 * time.Millisecond) // srv needs a blip to actually start
				response, err := http.Get("http://:8083")
				Expect(err).To(BeNil())

				bdy, err := io.ReadAll(response.Body)
				response.Body.Close()
				Expect(err).To(BeNil())
				Expect(bdy).To(BeEquivalentTo(`{"ima": "pc"}`))

				// check for shutdown

				cancel()

				Eventually(ic).Should(HaveLen(4))
				Expect(ic()[2].Msg).To(Equal("shutting down http service"))
				Expect(ic()[3].Msg).To(Equal("http service stopped"))
			})
		})
	})

	Describe("working out the object handler", func() {
		var (
			writer  *httptest.ResponseRecorder
			request *http.Request
		)

		When("all goes well", func() {
			BeforeEach(func() {

				hf := ObjHandler("stuff", map[string]string{"thing": "one"}, lgr)
				writer = httptest.NewRecorder()
				request = &http.Request{}

				hf(writer, request)
			})

			It("responds with an named, marshalled object", func() {
				Expect(writer.Code).To(Equal(200))
				Expect(writer.Header()).To(Equal(http.Header{"Content-Type": []string{"application/json"}}))
				Expect(writer.Body.String()).To(Equal(`{"stuff":{"thing":"one"}}`))
			})
		})
	})
})
