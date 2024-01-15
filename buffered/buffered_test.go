package buffered

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestBuffered(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Buffered Suite")
}

var _ = Describe("Buf", func() {
	var (
		buf         *Buffered
		contentType http.Header
	)

	BeforeEach(func() {
		contentType = http.Header{"Content-Type": []string{"application/json"}}
		buf = &Buffered{}
	})

	Describe("getting the header", func() {
		var (
			hdr http.Header
		)

		JustBeforeEach(func() {
			hdr = buf.Header()
		})

		When("all is well", func() {
			BeforeEach(func() {
				buf = &Buffered{
					Writer: &httptest.ResponseRecorder{
						HeaderMap: contentType,
					},
				}
			})

			It("returns the header", func() {
				Expect(hdr).To(Equal(contentType))
			})
		})

	})

	Describe("setting the status", func() {
		var (
			sts int
		)

		JustBeforeEach(func() {
			buf.WriteHeader(sts)
		})

		When("all is well", func() {
			BeforeEach(func() {
				sts = 201
			})

			It("stores the status", func() {
				Expect(buf.Status).To(Equal(201))
			})
		})

	})

	Describe("setting the body", func() {
		var (
			bdy []byte
			err error
		)

		JustBeforeEach(func() {
			_, err = buf.Write(bdy)
		})

		When("all is well", func() {
			BeforeEach(func() {
				bdy = []byte(`{"ima": "pc"}`)
				buf = &Buffered{
					Buffer: bytes.Buffer{},
				}
			})

			It("stores the body", func() {
				Expect(err).ToNot(HaveOccurred())
				Expect(buf.Buffer.String()).To(Equal(`{"ima": "pc"}`))
			})
		})

	})

	Describe("getting the body", func() {
		var (
			bdy string
		)

		JustBeforeEach(func() {
			bdy = buf.Body()
		})

		When("all is well", func() {
			BeforeEach(func() {
				buf = &Buffered{
					Buffer: *(bytes.NewBufferString(`{"ima": "pc"}`)),
				}
			})

			It("gets the body", func() {
				Expect(bdy).To(Equal(`{"ima": "pc"}`))
			})
		})

	})

	Describe("writing to the response writer", func() {
		var (
			err error
		)

		JustBeforeEach(func() {
			err = buf.WriteResponse()
		})

		When("all is well", func() {
			BeforeEach(func() {
				buf = &Buffered{
					Writer: &httptest.ResponseRecorder{
						Body: &bytes.Buffer{},
					},
					Status: 201,
					Buffer: *(bytes.NewBufferString(`{"ima": "pc"}`)),
				}
			})

			It("writes to its response writer", func() {
				Expect(err).ToNot(HaveOccurred())

				result := buf.Writer.(*httptest.ResponseRecorder).Result()
				Expect(result.StatusCode).To(Equal(201))

				resultBody, err := io.ReadAll(result.Body)
				Expect(err).ToNot(HaveOccurred())
				Expect(resultBody).To(BeEquivalentTo(`{"ima": "pc"}`))
			})
		})

		When("write fails", func() {
			BeforeEach(func() {
				buf = &Buffered{
					Writer: &errorResponder{},
				}
			})

			It("returns an error", func() {
				Expect(err).To(HaveOccurred())
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
