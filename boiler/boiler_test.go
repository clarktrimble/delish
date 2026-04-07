package boiler_test

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/clarktrimble/delish/boiler"
)

//go:generate moq -pkg boiler_test -out mock_test.go ../logger Logger

func TestBoiler(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Boiler Suite")
}

var _ = Describe("Register", func() {

	type svcCfg struct {
		Foo     string `json:"foo"`
		Version string `json:"version"`
		Release string `json:"release"`
		Url     string `json:"url"`
	}

	var (
		ctx  context.Context
		cfg  *svcCfg
		spec []byte
		lgr  *LoggerMock
		rtr  *http.ServeMux
	)

	BeforeEach(func() {
		ctx = context.Background()
		cfg = &svcCfg{Foo: "bar"}
		spec = []byte("openapi: 3.0.0\ninfo:\n  title: Test API")
		lgr = &LoggerMock{
			InfoFunc:  func(ctx context.Context, msg string, kv ...any) {},
			ErrorFunc: func(ctx context.Context, msg string, err error, kv ...any) {},
		}
	})

	JustBeforeEach(func() {
		rtr = http.NewServeMux()
		boiler.Register(ctx, rtr, cfg, spec, lgr)
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
			Expect(string(body)).To(Equal("openapi: 3.0.0\ninfo:\n  title: Test API"))
		})
	})

	When("requesting /docs", func() {
		It("returns html with title from spec", func() {
			req := httptest.NewRequest("GET", "/docs", nil)
			rec := httptest.NewRecorder()
			rtr.ServeHTTP(rec, req)

			Expect(rec.Code).To(Equal(http.StatusOK))
			Expect(rec.Header().Get("Content-Type")).To(Equal("text/html"))
			body, _ := io.ReadAll(rec.Body)
			Expect(string(body)).To(ContainSubstring("Test API"))
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

	When("cfg has version fields", func() {
		BeforeEach(func() {
			cfg = &svcCfg{Foo: "bar", Version: "main.42.abc", Release: "1.2.3", Url: "https://example.com"}
			spec = []byte("openapi: 3.0.0\ninfo:\n  title: Test API\n  version: ${RELEASE}\nservers:\n  - url: ${PUBLISHED_URL}")
		})

		It("substitutes version fields into spec", func() {
			req := httptest.NewRequest("GET", "/openapi.yaml", nil)
			rec := httptest.NewRecorder()
			rtr.ServeHTTP(rec, req)

			body, _ := io.ReadAll(rec.Body)
			Expect(string(body)).To(ContainSubstring("version: 1.2.3"))
			Expect(string(body)).To(ContainSubstring("url: https://example.com"))
		})
	})

	When("release is empty but version is set", func() {
		BeforeEach(func() {
			cfg = &svcCfg{Version: "main.42.abc"}
			spec = []byte("openapi: 3.0.0\ninfo:\n  title: Test API\n  version: ${RELEASE}")
		})

		It("falls back to version with underscore prefix", func() {
			req := httptest.NewRequest("GET", "/openapi.yaml", nil)
			rec := httptest.NewRecorder()
			rtr.ServeHTTP(rec, req)

			body, _ := io.ReadAll(rec.Body)
			Expect(string(body)).To(ContainSubstring("version: _main.42.abc"))
		})
	})

	When("spec has no title", func() {
		BeforeEach(func() {
			spec = []byte("openapi: 3.0.0")
		})

		It("falls back to default title in docs", func() {
			req := httptest.NewRequest("GET", "/docs", nil)
			rec := httptest.NewRecorder()
			rtr.ServeHTTP(rec, req)

			body, _ := io.ReadAll(rec.Body)
			Expect(string(body)).To(ContainSubstring("API Documentation"))
		})
	})
})
