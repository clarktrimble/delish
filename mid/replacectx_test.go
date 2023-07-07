package mid_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	. "github.com/clarktrimble/delish/mid"
)

func TestMid(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Mid Suite")
}

var _ = Describe("ReplaceCtx", func() {
	var (
		ctx     context.Context
		handler http.Handler
		val     string
	)

	BeforeEach(func() {
		ctx = context.WithValue(context.Background(), key{}, "Friedrich Georg Wilhelm von Struve")
		handler = http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
			val, _ = request.Context().Value(key{}).(string)
		})
	})

	Describe("replacing the request ctx", func() {

		JustBeforeEach(func() {
			handler.ServeHTTP(httptest.NewRecorder(), &http.Request{})
		})

		When("value is stored in ctx and handler is wrapped", func() {
			BeforeEach(func() {
				handler = ReplaceCtx(ctx, handler)
			})

			It("has the request context with the value", func() {
				Expect(val).To(Equal("Friedrich Georg Wilhelm von Struve"))
			})
		})

		When("value is stored in ctx and handler is not wrapped", func() {

			It("does not have the request context with the value", func() {
				Expect(val).To(Equal(""))
			})
		})

	})
})

type key struct{}
