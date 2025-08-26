package mid

import (
	"bytes"
	"context"
	"io"
	"math/rand"
	"net/http"
	"net/http/httptest"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("LogRequest", func() {
	var (
		handler  http.Handler
		request  *http.Request
		received *http.Request
		lgr      *loggerMock
	)

	BeforeEach(func() {
		handler = http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
			received = request
		})

		lgr = &loggerMock{
			DebugFunc: func(ctx context.Context, msg string, kv ...any) {},
			WithFieldsFunc: func(ctx context.Context, kv ...any) context.Context {
				return ctx
			},
		}

		RedactHeaders = map[string]bool{}
		SkipBody = false

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
					ic := lgr.DebugCalls()
					Expect(ic).To(HaveLen(1))
					Expect(ic[0].Msg).To(Equal("received request"))
					Expect(ic[0].Kv).To(HaveExactElements([]any{
						"method", "",
						"path", "",
						"query", map[string][]string(nil),
						"remote_ip", "",
						"remote_port", "",
						"headers", http.Header(nil),
						"body", "",
					}))

					wfc := lgr.WithFieldsCalls()
					Expect(wfc).To(HaveLen(1))
					// Todo: golang upgrade broke sommat, fix!!
					//Expect(wfc[0].Kv).To(HaveExactElements([]any{
					//	"request_id", "GIehp1s",
					//}))
				})
			})

			When("the request is well rounded", func() {
				BeforeEach(func() {
					bdy := bytes.NewBufferString(`{"ima":"pc"}`)
					request, _ = http.NewRequest("POST", "www.boxworld.net://baltic/latvia/riga", bdy)
					request.RemoteAddr = "10.11.12.13:34562"
					request.Header.Set("content-type", "application/json")
				})

				It("logs fields related to the request and body is intact", func() {
					ic := lgr.DebugCalls()
					Expect(ic).To(HaveLen(1))
					Expect(ic[0].Msg).To(Equal("received request"))
					Expect(ic[0].Kv).To(HaveExactElements([]any{
						"method", "POST",
						"path", "/latvia/riga",
						"query", map[string][]string{},
						"remote_ip", "10.11.12.13",
						"remote_port", "34562",
						"headers", http.Header{"Content-Type": []string{"application/json"}},
						"body", `{"ima":"pc"}`,
					}))

					body, err := io.ReadAll(received.Body)
					received.Body.Close()
					Expect(err).ToNot(HaveOccurred())
					Expect(string(body)).To(Equal(`{"ima":"pc"}`))
				})

				When("and a header is flagged for redaction", func() {
					BeforeEach(func() {
						RedactHeaders = map[string]bool{"X-Authorization-Token": true}
						request.Header.Set("X-Authorization-Token", "this-is-secret")
					})

					It("redacts that header in the logging", func() {
						ic := lgr.DebugCalls()
						Expect(ic).To(HaveLen(1))
						Expect(ic[0].Msg).To(Equal("received request"))
						Expect(ic[0].Kv).To(HaveExactElements([]any{
							"method", "POST",
							"path", "/latvia/riga",
							"query", map[string][]string{},
							"remote_ip", "10.11.12.13",
							"remote_port", "34562",
							"headers", http.Header{
								"Content-Type":          []string{"application/json"},
								"X-Authorization-Token": []string{"--redacted--"},
							},
							"body", `{"ima":"pc"}`,
						}))
					})
				})

				When("and body skipping is enabled", func() {
					BeforeEach(func() {
						SkipBody = true
					})

					It("does not log the body and body is intact", func() {
						ic := lgr.DebugCalls()
						Expect(ic).To(HaveLen(1))
						Expect(ic[0].Msg).To(Equal("received request"))
						Expect(ic[0].Kv).To(HaveExactElements([]any{
							"method", "POST",
							"path", "/latvia/riga",
							"query", map[string][]string{},
							"remote_ip", "10.11.12.13",
							"remote_port", "34562",
							"headers", http.Header{
								"Content-Type": []string{"application/json"},
							},
						}))

						body, err := io.ReadAll(received.Body)
						received.Body.Close()
						Expect(err).ToNot(HaveOccurred())
						Expect(string(body)).To(Equal(`{"ima":"pc"}`))
					})
				})

			})

		})
	})
})
