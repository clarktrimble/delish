package boiler

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

//go:generate moq -pkg boiler -out mock_test.go ../logger Logger

func TestBoiler(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Boiler Suite")
}

var _ = Describe("SubSpec", func() {

	var (
		spec    []byte
		version string
		release string
		url     string
		result  []byte
	)

	BeforeEach(func() {
		spec = []byte("version: ${RELEASE}\nurl: ${PUBLISHED_URL}")
		version = "abc123"
		release = "1.2.3"
		url = "https://example.com"
	})

	JustBeforeEach(func() {
		result = SubSpec(spec, version, release, url)
	})

	When("release is a proper version", func() {
		It("substitutes release and url", func() {
			Expect(string(result)).To(Equal("version: 1.2.3\nurl: https://example.com"))
		})
	})

	When("release is untagged", func() {
		BeforeEach(func() {
			release = "untagged"
		})

		It("falls back to version with underscore prefix", func() {
			Expect(string(result)).To(Equal("version: _abc123\nurl: https://example.com"))
		})
	})

	When("release is empty", func() {
		BeforeEach(func() {
			version = ""
			release = ""
		})

		It("uses _unreleased", func() {
			Expect(string(result)).To(Equal("version: _unreleased\nurl: https://example.com"))
		})
	})
})

var _ = Describe("NewRouter", func() {

	var (
		ctx   context.Context
		cfg   any
		title string
		spec  []byte
		lgr   *LoggerMock
		rtr   *http.ServeMux
	)

	BeforeEach(func() {
		ctx = context.Background()
		cfg = map[string]string{"foo": "bar"}
		title = "Test API"
		spec = []byte("openapi: 3.0.0")
		lgr = &LoggerMock{
			InfoFunc:  func(ctx context.Context, msg string, kv ...any) {},
			ErrorFunc: func(ctx context.Context, msg string, err error, kv ...any) {},
		}
	})

	JustBeforeEach(func() {
		rtr = NewRouter(ctx, cfg, title, spec, lgr)
	})

	It("returns a non-nil router", func() {
		Expect(rtr).ToNot(BeNil())
	})

	When("requesting /monitor", func() {
		It("returns ok status", func() {
			req := httptest.NewRequest("GET", "/monitor", nil)
			rec := httptest.NewRecorder()
			rtr.ServeHTTP(rec, req)

			Expect(rec.Code).To(Equal(http.StatusOK))
			body, _ := io.ReadAll(rec.Body)
			Expect(string(body)).To(ContainSubstring(`"status":"ok"`))
		})
	})

	When("requesting /config", func() {
		It("returns config as json", func() {
			req := httptest.NewRequest("GET", "/config", nil)
			rec := httptest.NewRecorder()
			rtr.ServeHTTP(rec, req)

			Expect(rec.Code).To(Equal(http.StatusOK))
			body, _ := io.ReadAll(rec.Body)
			Expect(string(body)).To(ContainSubstring(`"foo":"bar"`))
		})
	})

	When("requesting /openapi.yaml", func() {
		It("returns the spec with correct content type", func() {
			req := httptest.NewRequest("GET", "/openapi.yaml", nil)
			rec := httptest.NewRecorder()
			rtr.ServeHTTP(rec, req)

			Expect(rec.Code).To(Equal(http.StatusOK))
			Expect(rec.Header().Get("Content-Type")).To(Equal("application/x-yaml"))
			body, _ := io.ReadAll(rec.Body)
			Expect(string(body)).To(Equal("openapi: 3.0.0"))
		})
	})

	When("requesting /docs", func() {
		It("returns html", func() {
			req := httptest.NewRequest("GET", "/docs", nil)
			rec := httptest.NewRecorder()
			rtr.ServeHTTP(rec, req)

			Expect(rec.Code).To(Equal(http.StatusOK))
			Expect(rec.Header().Get("Content-Type")).To(Equal("text/html"))
		})
	})

	When("requesting /elements.min.js", func() {
		It("returns gzipped javascript with cache headers", func() {
			req := httptest.NewRequest("GET", "/elements.min.js", nil)
			rec := httptest.NewRecorder()
			rtr.ServeHTTP(rec, req)

			Expect(rec.Code).To(Equal(http.StatusOK))
			Expect(rec.Header().Get("Content-Type")).To(Equal("application/javascript"))
			Expect(rec.Header().Get("Content-Encoding")).To(Equal("gzip"))
			Expect(rec.Header().Get("Cache-Control")).To(Equal("public, max-age=31536000"))
		})
	})

	When("requesting /elements.min.css", func() {
		It("returns gzipped css with cache headers", func() {
			req := httptest.NewRequest("GET", "/elements.min.css", nil)
			rec := httptest.NewRecorder()
			rtr.ServeHTTP(rec, req)

			Expect(rec.Code).To(Equal(http.StatusOK))
			Expect(rec.Header().Get("Content-Type")).To(Equal("text/css"))
			Expect(rec.Header().Get("Content-Encoding")).To(Equal("gzip"))
			Expect(rec.Header().Get("Cache-Control")).To(Equal("public, max-age=31536000"))
		})
	})
})
