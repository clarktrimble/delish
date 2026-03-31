package boiler

import (
	"bytes"
	"context"
	_ "embed"
	"net/http"

	"github.com/clarktrimble/delish"
	"github.com/clarktrimble/delish/logger"
)

//go:embed paths.yaml
var pathsYaml []byte

// SubSpec substitutes ${RELEASE} and ${PUBLISHED_URL} placeholders in an OpenAPI spec.
// Version is a branch.revcount.revhash string (e.g. "main.42.abc1234"), used as a fallback
// when release is "untagged". Release is a git tag (e.g. "1.2.3") or "untagged".
// Todo: replace with apispec
func SubSpec(spec []byte, version, release, url string) []byte {

	apiRelease := release
	if release == "untagged" {
		apiRelease = "_" + version
	}
	if apiRelease == "" {
		apiRelease = "_unreleased"
	}
	result := bytes.Replace(spec, []byte("${RELEASE}"), []byte(apiRelease), 1)
	return bytes.Replace(result, []byte("${PUBLISHED_URL}"), []byte(url), 1)
}

//go:embed docs.html
var docsHtml []byte

//go:embed elements.min.js.gz
var elementsJs []byte

//go:embed elements.min.css.gz
var elementsCss []byte

// NewRouter creates a router with boilerplate routes.
func NewRouter(ctx context.Context, cfg any, title string, spec []byte, lgr logger.Logger) (rtr *http.ServeMux) {

	docs := bytes.Replace(docsHtml, []byte("${TITLE}"), []byte(title), 1)

	rtr = http.NewServeMux()
	rtr.HandleFunc("GET /config", delish.ObjHandler("config", cfg, lgr))
	rtr.HandleFunc("GET /monitor", delish.ObjHandler("status", "ok", lgr))
	rtr.HandleFunc("POST /log/{level}", delish.LogLevel(ctx, lgr))
	rtr.HandleFunc("GET /log", delish.GetLogLevel(ctx, lgr))
	rtr.HandleFunc("GET /docs", func(writer http.ResponseWriter, request *http.Request) {
		writer.Header().Set("Content-Type", "text/html")
		_, _ = writer.Write(docs)
	})
	rtr.HandleFunc("GET /elements.min.js", getDocsJs)
	rtr.HandleFunc("GET /elements.min.css", getDocsCss)
	rtr.HandleFunc("GET /openapi.yaml", func(writer http.ResponseWriter, request *http.Request) {
		writer.Header().Set("Content-Type", "application/x-yaml")
		writer.Header().Set("Access-Control-Allow-Origin", "*")
		_, _ = writer.Write(spec)
	})

	return
}

// ApiSpec documents endpoints provided.
func ApiSpec() ([]byte, map[string]any) {
	return pathsYaml, nil
}

// unexported

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
