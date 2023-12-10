package delish_test

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"sync"
	"testing"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	. "github.com/clarktrimble/delish"
	"github.com/clarktrimble/delish/test/mock"
)

func TestDelish(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Delish Suite")
}

var _ = Describe("Server", func() {
	var (
		handler http.Handler
		lgr     *mock.Logger
		svr     *Server
	)

	BeforeEach(func() {
		lgr = mock.NewLogger()
	})

	Describe("creating a server", func() {
		var (
			cfg *Config
		)

		BeforeEach(func() {
			cfg = &Config{
				Port:    8083,
				Timeout: 33 * time.Second,
			}
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
			bdy    []byte
		)

		JustBeforeEach(func() {
			// give the server a few cycles to start
			time.Sleep(19 * time.Millisecond)

			response, err := http.Get("http://:8083")
			Expect(err).To(BeNil())

			bdy, err = io.ReadAll(response.Body)
			response.Body.Close()
			Expect(err).To(BeNil())

			cancel()
			// give shutdown a few to complete
			time.Sleep(19 * time.Millisecond)
		})

		When("all goes well", func() {
			BeforeEach(func() {
				ctx, cancel = context.WithCancel(context.Background())
				handler = http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
					fmt.Fprint(writer, `{"ima": "pc"}`)
				})

				svr = (&Config{Port: 8083}).New(handler, lgr)
				svr.Start(ctx, &wg)
			})

			It("starts, serves, and stops", func() {
				Expect(lgr.Logged).To(HaveLen(4))
				Expect(lgr.Logged[0]["msg"]).To(Equal("starting http service"))
				Expect(lgr.Logged[1]["msg"]).To(Equal("listening"))

				Expect(bdy).To(BeEquivalentTo(`{"ima": "pc"}`))

				Expect(lgr.Logged[2]["msg"]).To(Equal("shutting down http service .."))
				Expect(lgr.Logged[3]["msg"]).To(Equal("http service stopped"))
				// Todo: sometimes getting "shutdown failed" ??
				//       doubling sleeps to 19 above did not help, but "feels" like a timing issue ..
			})
		})
	})

})
