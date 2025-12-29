package respond

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

//go:generate moq -pkg respond -out mock_test.go ../logger Logger

func TestRespond(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Respond Suite")
}

var _ = Describe("Respond(ing)", func() {
	var (
		ctx    context.Context
		writer *httptest.ResponseRecorder
		lgr    *LoggerMock
		rp     *Respond
	)

	BeforeEach(func() {
		ctx = context.Background()
		writer = httptest.NewRecorder()

		lgr = &LoggerMock{
			ErrorFunc: func(ctx context.Context, msg string, err error, kv ...any) {},
		}

		rp = &Respond{
			Writer: writer,
			Logger: lgr,
		}
	})

	Describe("with ok", func() {

		JustBeforeEach(func() {
			rp.Ok(ctx)
		})

		When("all goes well", func() {
			It("responds with http status and blerb", func() {

				Expect(writer.Code).To(Equal(200))
				Expect(writer.Body.String()).To(Equal(`{"status":"ok"}`))
			})
		})
	})

	Describe("with not ok", func() {
		var (
			code int
			err  error
		)

		JustBeforeEach(func() {
			rp.NotOk(ctx, code, err)
		})

		When("all goes well", func() {
			BeforeEach(func() {
				code = 500
				err = fmt.Errorf("oops")
			})

			It("responds with http status, error body, and logs the error", func() {

				Expect(writer.Code).To(Equal(500))
				Expect(writer.Body.String()).To(Equal(`{"error":"oops"}`))

				ec := lgr.ErrorCalls()
				Expect(ec).To(HaveLen(1))
				Expect(ec[0].Msg).To(Equal("returning error to client"))
				Expect(ec[0].Err.Error()).To(Equal("oops"))
			})
		})
	})

	Describe("with not found", func() {

		JustBeforeEach(func() {
			rp.NotFound(ctx)
		})

		When("all goes well", func() {

			It("responds with 404 and not found body", func() {

				Expect(writer.Code).To(Equal(404))
				Expect(writer.Body.String()).To(Equal(`{"not":"found"}`))
			})
		})
	})

	Describe("with objects", func() {
		var (
			objects map[string]any
		)

		JustBeforeEach(func() {
			rp.WriteObjects(ctx, objects)
		})

		When("all goes well", func() {
			BeforeEach(func() {
				objects = map[string]any{"ima": "pc"}
			})

			It("responds with http ok, application/json, and marshalled objects", func() {

				Expect(writer.Code).To(Equal(200))
				Expect(writer.Header()).To(Equal(http.Header{"Content-Type": []string{"application/json"}}))
				Expect(writer.Body.String()).To(Equal(`{"ima":"pc"}`))
			})
		})

		When("marshal fails", func() {
			BeforeEach(func() {
				objects = map[string]any{"foo": make(chan int)}
			})

			It("responds with 500, error body, and logs the error", func() {

				Expect(writer.Code).To(Equal(500))
				Expect(writer.Header()).To(Equal(http.Header{"Content-Type": []string{"application/json"}}))
				Expect(writer.Body.String()).To(Equal(`{"error": "failed to encode response"}`))

				ec := lgr.ErrorCalls()
				Expect(ec).To(HaveLen(1))
				Expect(ec[0].Msg).To(Equal("failed to encode response"))
				Expect(ec[0].Err.Error()).To(ContainSubstring("unsupported type"))
			})
		})

	})

	Describe("with bytes", func() {
		var (
			data []byte
		)

		JustBeforeEach(func() {
			rp.Write(ctx, data)
		})

		When("all goes well", func() {
			BeforeEach(func() {
				data = []byte(`{"ima":"pc"}`)
			})

			It("responds with ok and body", func() {

				Expect(writer.Code).To(Equal(200))
				Expect(writer.Header()).To(Equal(http.Header{"Content-Type": []string{"text/plain; charset=utf-8"}}))
				Expect(writer.Body.String()).To(Equal(`{"ima":"pc"}`))
			})
		})

		When("write fails", func() {
			BeforeEach(func() {
				data = []byte(`{"ima":"pc"}`)
				rp.Writer = &errorResponder{}
			})

			It("logs the error", func() {
				ec := lgr.ErrorCalls()
				Expect(ec).To(HaveLen(1))
				Expect(ec[0].Msg).To(Equal("failed to write response"))
				Expect(ec[0].Err.Error()).To(Equal("failed to write response: oops"))
			})
		})

	})
})

type errorResponder struct{}

func (er *errorResponder) Header() (hdr http.Header) {
	return http.Header{}
}

func (er *errorResponder) Write(body []byte) (count int, err error) {
	return 0, fmt.Errorf("oops")
}

func (er *errorResponder) WriteHeader(status int) {}
