package mid_test

import (
	"bytes"
	"context"
	"math/rand"
	"net/http"
	"net/http/httptest"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	. "github.com/clarktrimble/delish/mid"
	"github.com/clarktrimble/delish/mock"
)

var _ = Describe("LogRequest", func() {
	var (
		handler http.Handler
		request *http.Request
		lgr     *mock.LoggerMock
	)

	BeforeEach(func() {
		handler = http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {})

		lgr = &mock.LoggerMock{
			InfoFunc: func(ctx context.Context, msg string, kv ...any) {},
			WithFieldsFunc: func(ctx context.Context, kv ...any) context.Context {
				return ctx
			},
		}

		rand.Seed(1) //nolint:staticcheck // unit request_id
	})

	Describe("logging the request", func() {

		JustBeforeEach(func() {
			handler.ServeHTTP(httptest.NewRecorder(), request)
		})

		Describe("the hander is wrapped with the middleware", func() {
			BeforeEach(func() {
				handler = LogRequest(lgr, handler)
			})

			When("the request is empty", func() {
				BeforeEach(func() {
					request = &http.Request{}
				})

				It("logs mostly empty fields related to the request and does not panic", func() {
					ic := lgr.InfoCalls()
					Expect(ic).To(HaveLen(1))
					Expect(ic[0].Msg).To(Equal("received request"))
					Expect(ic[0].Kv).To(HaveExactElements([]any{
						"method", "",
						"path", "",
						"query", map[string][]string(nil),
						"body", "",
						"remote_ip", "",
						"remote_port", "",
						"headers", http.Header(nil),
					}))

					wfc := lgr.WithFieldsCalls()
					Expect(wfc).To(HaveLen(1))
					Expect(wfc[0].Kv).To(HaveExactElements([]any{
						"request_id", "GIehp1s",
					}))
				})
			})

			When("the request is well rounded", func() {
				BeforeEach(func() {
					bdy := bytes.NewBufferString(`{"ima":"pc"}`)
					request, _ = http.NewRequest("POST", "www.boxworld.net://baltic/latvia/riga", bdy)
					request.RemoteAddr = "10.11.12.13:34562"
					request.Header.Set("content-type", "application/json")
				})

				It("logs fields related to the request", func() {
					ic := lgr.InfoCalls()
					Expect(ic).To(HaveLen(1))
					Expect(ic[0].Msg).To(Equal("received request"))
					Expect(ic[0].Kv).To(HaveExactElements([]any{
						"method", "POST",
						"path", "/latvia/riga",
						"query", map[string][]string{},
						"body", `{"ima":"pc"}`,
						"remote_ip", "10.11.12.13",
						"remote_port", "34562",
						"headers", http.Header{"Content-Type": []string{"application/json"}},
					}))
				})

				When("and a header is flagged for redaction", func() {
					BeforeEach(func() {
						RedactHeaders = map[string]bool{"X-Authorization-Token": true}
						request.Header.Set("X-Authorization-Token", "this-is-secret")
					})

					It("redacts that header in the logging", func() {
						ic := lgr.InfoCalls()
						Expect(ic).To(HaveLen(1))
						Expect(ic[0].Msg).To(Equal("received request"))
						Expect(ic[0].Kv).To(HaveExactElements([]any{
							"method", "POST",
							"path", "/latvia/riga",
							"query", map[string][]string{},
							"body", `{"ima":"pc"}`,
							"remote_ip", "10.11.12.13",
							"remote_port", "34562",
							"headers", http.Header{
								"Content-Type":          []string{"application/json"},
								"X-Authorization-Token": []string{"--redacted--"},
							},
						}))
					})
				})

			})

		})
	})
})
