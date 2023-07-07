package mid_test

import (
	"net/http"
	"net/http/httptest"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	. "github.com/clarktrimble/delish/mid"
	"github.com/clarktrimble/delish/test/mock"
)

var _ = Describe("LogResponse", func() {
	var (
		handler http.Handler
		lgr     *mock.Logger
	)

	BeforeEach(func() {
		handler = jsonHandler(201, `{"ima":"pc"}`)

		lgr = mock.NewLogger()
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
				Expect(lgr.Logged).To(HaveLen(1))
				Expect(delog(lgr.Logged[0])).To(Equal(map[string]any{
					"body":    `{"ima":"pc"}`,
					"elapsed": "replaced-for-unit",
					"headers": http.Header{"Content-Type": []string{"application/json"}},
					"msg":     "sending response",
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

func delog(line map[string]any) (scrub map[string]any) {

	scrub = map[string]any{}
	for key, val := range line {

		if key == "elapsed" {
			Expect(val).To(BeNumerically(">", 100))
			Expect(val).To(BeNumerically("<", 1000000))
			val = "replaced-for-unit"
		}
		scrub[key] = val
	}

	return
}
