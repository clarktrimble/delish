package boiler

import (
	"bytes"
	"context"
	_ "embed"
	"net/http"

	"github.com/clarktrimble/delish"
	"github.com/clarktrimble/delish/logger"
)

// Config specifies boiler options.
// Embed in app config to satisfy AppConfig interface.
type Config struct {
	Version string `json:"version" ignored:"true"`
	Release string `json:"release" ignored:"true"`
	Url     string `json:"url" desc:"URL referenced from API spec" default:"http://localhost:3031"`
}

// AppRelease gets version.
func (cfg *Config) AppRelease() (release string) {

	release = cfg.Release
	if release == "untagged" {
		release = cfg.Version
	}
	if release == "" {
		release = "unreleased"
	}
	return
}

// AppUrl returns the published URL for the API spec.
func (cfg *Config) AppUrl() string {
	return cfg.Url
}

// AppConfig is satisfied by embedding Config.
type AppConfig interface {
	AppRelease() string
	AppUrl() string
}

// Todo: golang runtime stats ftw
// Todo: unit and doc
// Todo: I canhaz pluggable js/css?

//go:embed docs.html
var docsHtml []byte

//go:embed elements.min.js.gz
var elementsJs []byte

//go:embed elements.min.css.gz
var elementsCss []byte

// NewRouter creates a router with boilerplate routes.
// The openapiSpec may contain ${RELEASE} and ${PUBLISHED_URL} placeholders
// which are replaced with values from cfg.
func NewRouter(ctx context.Context, cfg AppConfig, openapiSpec []byte, lgr logger.Logger) (rtr *http.ServeMux) {

	spec := bytes.Replace(openapiSpec, []byte("${RELEASE}"), []byte(cfg.AppRelease()), 1)
	spec = bytes.Replace(spec, []byte("${PUBLISHED_URL}"), []byte(cfg.AppUrl()), 1)

	rtr = http.NewServeMux()
	rtr.HandleFunc("GET /config", delish.ObjHandler("config", cfg, lgr))
	rtr.HandleFunc("GET /monitor", delish.ObjHandler("status", "ok", lgr))
	rtr.HandleFunc("POST /log/{level}", delish.LogLevel(ctx, lgr))
	rtr.HandleFunc("GET /log", delish.GetLogLevel(ctx, lgr))
	rtr.HandleFunc("GET /docs", getDocs)
	rtr.HandleFunc("GET /elements.min.js", getDocsJs)
	rtr.HandleFunc("GET /elements.min.css", getDocsCss)
	rtr.HandleFunc("GET /openapi.yaml", func(writer http.ResponseWriter, request *http.Request) {
		writer.Header().Set("Content-Type", "application/x-yaml")
		writer.Header().Set("Access-Control-Allow-Origin", "*")
		_, _ = writer.Write(spec)
	})

	return
}

// unexported

func getDocs(writer http.ResponseWriter, request *http.Request) {
	writer.Header().Set("Content-Type", "text/html")
	_, _ = writer.Write(docsHtml)
}

func getDocsJs(writer http.ResponseWriter, request *http.Request) {
	writer.Header().Set("Content-Type", "application/javascript")
	writer.Header().Set("Content-Encoding", "gzip")
	writer.Header().Set("Cache-Control", "public, max-age=31536000")
	_, _ = writer.Write(elementsJs)
}

func getDocsCss(writer http.ResponseWriter, request *http.Request) {
	writer.Header().Set("Content-Type", "text/css")
	writer.Header().Set("Content-Encoding", "gzip")
	writer.Header().Set("Cache-Control", "public, max-age=31536000")
	_, _ = writer.Write(elementsCss)
}
