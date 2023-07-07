package mid_test

import (
	"bytes"
	"net/http"
	"net/http/httptest"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	. "github.com/clarktrimble/delish/mid"
	"github.com/clarktrimble/delish/test/mock"
)

var _ = Describe("LogRequest", func() {
	var (
		handler http.Handler
		lgr     *mock.Logger
		request *http.Request
	)

	BeforeEach(func() {
		handler = http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		})
		lgr = mock.NewLogger()
	})

	Describe("logging the request", func() {

		JustBeforeEach(func() {
			handler.ServeHTTP(httptest.NewRecorder(), request)
		})

		Describe("the hander is wrapped with the middleware", func() {
			BeforeEach(func() {
				handler = LogRequest(lgr, notRand, handler)
			})

			When("the request is empty", func() {
				BeforeEach(func() {
					request = &http.Request{}
				})

				It("logs mostly empty fields related to the request and does not panic", func() {
					Expect(lgr.Logged).To(HaveLen(1))
					Expect(lgr.Logged[0]["msg"]).To(Equal("received request"))
					Expect(lgr.Logged[0]).To(Equal(map[string]any{
						"body":        "",
						"headers":     http.Header(nil),
						"method":      "",
						"msg":         "received request",
						"path":        "",
						"query":       map[string][]string(nil),
						"remote_ip":   "",
						"remote_port": "",
						"request_id":  "123123123",
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
					Expect(lgr.Logged).To(HaveLen(1))
					Expect(lgr.Logged[0]).To(Equal(map[string]any{
						"body":        `{"ima":"pc"}`,
						"headers":     http.Header{"Content-Type": []string{"application/json"}},
						"method":      "POST",
						"msg":         "received request",
						"path":        "/latvia/riga",
						"query":       map[string][]string{},
						"remote_ip":   "10.11.12.13",
						"remote_port": "34562",
						"request_id":  "123123123",
					}))
				})
			})
		})

	})
})

func notRand(n int) string {

	return "123123123"
}
