package mid_test

import (
	"context"
	"net/http"
	"net/http/httptest"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	. "github.com/clarktrimble/delish/mid"
	"github.com/clarktrimble/delish/mock"
)

var _ = Describe("LogResponse", func() {
	var (
		handler http.Handler
		lgr     *mock.LoggerMock
	)

	BeforeEach(func() {
		handler = jsonHandler(201, `{"ima":"pc"}`)
		lgr = &mock.LoggerMock{
			InfoFunc: func(ctx context.Context, msg string, kv ...any) {},
		}
	})

	Describe("logging the response", func() {
		JustBeforeEach(func() {
			handler.ServeHTTP(httptest.NewRecorder(), &http.Request{})
		})

		When("the hander is wrapped with the middleware", func() {
			BeforeEach(func() {
				handler = LogResponse(lgr, handler)
			})

			It("logs fields related to the response", func() {
				ic := lgr.InfoCalls()
				Expect(ic).To(HaveLen(1))
				Expect(ic[0].Msg).To(Equal("sending response"))
				Expect(mapLog(ic[0].Kv)).To(Equal(map[string]any{
					"body":    `{"ima":"pc"}`,
					"elapsed": "replaced-for-unit",
					"headers": http.Header{"Content-Type": []string{"application/json"}},
					"status":  201,
				}))
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
		key := kv[i].(string)
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
