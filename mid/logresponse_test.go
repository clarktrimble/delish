package mid

import (
	"context"
	"net/http"
	"net/http/httptest"
	"net/url"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("LogResponse", func() {
	var (
		handler  http.Handler
		recorder *httptest.ResponseRecorder
		lgr      *loggerMock
	)

	BeforeEach(func() {
		handler = jsonHandler(201, `{"ima":"pc"}`)
		recorder = httptest.NewRecorder()

		lgr = &loggerMock{
			DebugFunc: func(ctx context.Context, msg string, kv ...any) {},
		}

		SkipBody = false
	})

	Describe("logging the response", func() {

		JustBeforeEach(func() {
			url, err := url.Parse("http://test.com/blah")
			Expect(err).ToNot(HaveOccurred())
			handler.ServeHTTP(recorder, &http.Request{URL: url})
		})

		When("the hander is wrapped with the middleware", func() {
			BeforeEach(func() {
				handler = LogResponse(lgr, handler)
			})

			It("logs fields related to the response and body is intact", func() {
				ic := lgr.DebugCalls()
				Expect(ic).To(HaveLen(1))
				Expect(ic[0].Msg).To(Equal("sending response"))
				Expect(mapLog(ic[0].Kv)).To(Equal(map[string]any{
					"body":    `{"ima":"pc"}`,
					"elapsed": "replaced-for-unit",
					"headers": http.Header{"Content-Type": []string{"application/json"}},
					"status":  201,
				}))

				Expect(recorder.Body.String()).To(Equal(`{"ima":"pc"}`))
			})

			When("and skip body is enabled", func() {
				BeforeEach(func() {
					SkipBody = true
				})

				It("does not log body and body is intact", func() {
					ic := lgr.DebugCalls()
					Expect(ic).To(HaveLen(1))
					Expect(ic[0].Msg).To(Equal("sending response"))
					Expect(mapLog(ic[0].Kv)).To(Equal(map[string]any{
						"elapsed": "replaced-for-unit",
						"headers": http.Header{"Content-Type": []string{"application/json"}},
						"status":  201,
					}))

					Expect(recorder.Body.String()).To(Equal(`{"ima":"pc"}`))
				})
			})
		})

	})
})

func jsonHandler(code int, msg string) http.HandlerFunc {

	return func(writer http.ResponseWriter, request *http.Request) {

		writer.Header().Set("content-type", "application/json")
		writer.WriteHeader(code)

		_, err := writer.Write([]byte(msg))
		Expect(err).ToNot(HaveOccurred())
	}
}

func mapLog(kv []any) (mapped map[string]any) {

	mapped = map[string]any{}

	for i := 0; i < len(kv); i += 2 {
		key := kv[i].(string) //nolint:forcetypeassert // ok w panic from scuff helper
		val := kv[i+1]

		if key == "elapsed" {
			Expect(val).To(BeNumerically(">", 100))
			Expect(val).To(BeNumerically("<", 1000000))
			val = "replaced-for-unit"
		}
		mapped[key] = val
	}

	return
}
