package minroute

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestMinRoute(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "MinRoute Suite")
}

var _ = Describe("MinRoute", func() {
	var (
		ctx context.Context
		lgr *LoggerMock
		rtr *MinRoute
	)

	BeforeEach(func() {
		ctx = context.Background()

		lgr = &LoggerMock{
			ErrorFunc: func(ctx context.Context, msg string, err error, kv ...any) {},
		}

		rtr = New(ctx, lgr)
	})

	Describe("creating a router", func() {

		When("all goes well", func() {
			It("creates a router with empty table", func() {
				Expect(rtr).To(Equal(&MinRoute{
					Ctx:    ctx,
					Logger: lgr,
					Routes: map[string]map[string]http.HandlerFunc{
						"GET":    map[string]http.HandlerFunc{},
						"PUT":    map[string]http.HandlerFunc{},
						"POST":   map[string]http.HandlerFunc{},
						"DELETE": map[string]http.HandlerFunc{},
					},
				}))
			})
		})
	})

	Describe("setting a route", func() {
		var (
			hf http.HandlerFunc
		)

		BeforeEach(func() {
			hf = func(http.ResponseWriter, *http.Request) {}
		})

		When("all goes well", func() {
			BeforeEach(func() {
				rtr.HandleFunc("GET /stuff", hf)
			})

			It("registers the handler for that route", func() {
				expectPtr := fmt.Sprintf("%v", hf)
				gotPtr := fmt.Sprintf("%v", rtr.Routes["GET"]["/stuff"])
				Expect(gotPtr).To(Equal(expectPtr))

				Expect(lgr.ErrorCalls()).To(HaveLen(0))
			})
		})

		When("method is not supported", func() {
			BeforeEach(func() {
				rtr.HandleFunc("PATCH /stuff", hf)
			})

			It("logs an error", func() {
				ec := lgr.ErrorCalls()
				Expect(ec).To(HaveLen(1))
				Expect(ec[0].Err).To(MatchError("unsupported method from pattern: 'PATCH /stuff'"))
			})
		})

		When("method is not provided", func() {
			BeforeEach(func() {
				rtr.HandleFunc("/stuff", hf)
			})

			It("logs an error", func() {
				ec := lgr.ErrorCalls()
				Expect(ec).To(HaveLen(1))
				Expect(ec[0].Err).To(MatchError("failed to split pattern: '/stuff' into method and path"))
			})
		})

	})

	Describe("handling a request", func() {
		var (
			hf      http.HandlerFunc
			writer  *httptest.ResponseRecorder
			request *http.Request
			err     error
		)

		BeforeEach(func() {
			hf = func(writer http.ResponseWriter, request *http.Request) {
				_, err = writer.Write([]byte(`{"thing":"one"}`))
				Expect(err).ToNot(HaveOccurred())
			}

			rtr.HandleFunc("GET /stuff", hf)
			writer = httptest.NewRecorder()
		})

		JustBeforeEach(func() {
			rtr.ServeHTTP(writer, request)
		})

		When("all goes well", func() {
			BeforeEach(func() {
				request, err = http.NewRequest("GET", "http://boxworld.com/stuff", nil)
				Expect(err).ToNot(HaveOccurred())
			})

			It("responds via the handler", func() {
				Expect(writer.Code).To(Equal(200))
				Expect(writer.Body.String()).To(Equal(`{"thing":"one"}`))
			})
		})

		When("no handler for path", func() {
			BeforeEach(func() {
				request, err = http.NewRequest("PUT", "http://boxworld.com/stuff", nil)
				Expect(err).ToNot(HaveOccurred())
			})

			It("responds not found", func() {
				Expect(writer.Code).To(Equal(404))
				Expect(writer.Body.String()).To(Equal(`{"not":"found"}`))
			})
		})

		When("method is not supported", func() {
			BeforeEach(func() {
				request, err = http.NewRequest("PATCH", "http://boxworld.com/stuff", nil)
				Expect(err).ToNot(HaveOccurred())
			})

			It("responds not found", func() {
				Expect(writer.Code).To(Equal(404))
				Expect(writer.Body.String()).To(Equal(`{"not":"found"}`))
			})
		})

	})

})
