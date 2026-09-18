package boiler

import (
	"bytes"
	"context"
	_ "embed"
	"net/http"
	"reflect"

	"github.com/clarktrimble/delish"
	"github.com/clarktrimble/delish/logger"
	"gopkg.in/yaml.v3"
)

//go:embed docs.html
var docsHtml []byte

//go:embed elements.min.js.gz
var elementsJs []byte

//go:embed elements.min.css.gz
var elementsCss []byte

// Router specifies a router interface à la stdlib http.ServeMux.
type Router interface {
	HandleFunc(pattern string, handler func(http.ResponseWriter, *http.Request))
}

// Register adds boilerplate routes to a router.
func Register(ctx context.Context, rtr Router, cfg any, spec []byte, lgr logger.Logger) {

	fingerprint := stringField(cfg, "Fingerprint")
	// Todo: Remove Version support once consumers have converted to Fingerprint.
	if fingerprint == "" {
		fingerprint = stringField(cfg, "Version")
	}
	release := stringField(cfg, "Release")
	url := stringField(cfg, "Url")

	title := specTitle(spec, "API Documentation")
	spec = subSpec(spec, fingerprint, release, url)

	docs := bytes.ReplaceAll(docsHtml, []byte("${TITLE}"), []byte(title))

	rtr.HandleFunc("GET /config", delish.ObjHandler("config", cfg, lgr))
	rtr.HandleFunc("GET /monitor", delish.ObjHandler("status", "ok", lgr))
	rtr.HandleFunc("POST /log/{level}", delish.LogLevel(ctx, lgr))
	rtr.HandleFunc("GET /log", delish.GetLogLevel(ctx, lgr))
	rtr.HandleFunc("GET /docs", staticHandler(docs, "text/html"))
	rtr.HandleFunc("GET /openapi.yaml", staticHandler(spec, "application/x-yaml"))
	rtr.HandleFunc("GET /elements.min.js", gzipHandler(elementsJs, "application/javascript"))
	rtr.HandleFunc("GET /elements.min.css", gzipHandler(elementsCss, "text/css"))
}

// unexported

func staticHandler(body []byte, contentType string) http.HandlerFunc {
	return func(writer http.ResponseWriter, request *http.Request) {
		writer.Header().Set("Content-Type", contentType)
		_, _ = writer.Write(body)
	}
}

func gzipHandler(body []byte, contentType string) http.HandlerFunc {
	inner := staticHandler(body, contentType)
	return func(writer http.ResponseWriter, request *http.Request) {
		writer.Header().Set("Content-Encoding", "gzip")
		writer.Header().Set("Cache-Control", "public, max-age=31536000")
		inner(writer, request)
	}
}

func subSpec(spec []byte, fingerprint, release, url string) []byte {

	var label string
	switch {
	case release != "":
		label = release
	case fingerprint != "":
		label = "_" + fingerprint
	default:
		label = "_unreleased"
	}
	result := bytes.ReplaceAll(spec, []byte("${RELEASE}"), []byte(label))
	return bytes.ReplaceAll(result, []byte("${PUBLISHED_URL}"), []byte(url))
}

func specTitle(spec []byte, fallback string) string {
	var doc struct {
		Info struct {
			Title string `yaml:"title"`
		} `yaml:"info"`
	}
	if yaml.Unmarshal(spec, &doc) == nil && doc.Info.Title != "" {
		return doc.Info.Title
	}
	return fallback
}

func stringField(cfg any, name string) string {
	v := reflect.ValueOf(cfg)
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}
	if v.Kind() == reflect.Struct {
		f := v.FieldByName(name)
		if f.IsValid() && f.Kind() == reflect.String {
			return f.String()
		}
	}
	return ""
}
