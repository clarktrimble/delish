package boiler

import (
	"context"
	_ "embed"
	"net/http"

	"github.com/clarktrimble/delish"
	"github.com/clarktrimble/delish/logger"
)

// Todo: version made up in apispec
// Todo: how does url work in stoplight page?
// Todo: golang runtime stats ftw
// Todo: unit and doc
// Todo: I canhaz pluggable js/css?

//go:embed docs.html
var docsHtml []byte

//go:embed elements.min.js.gz
var elementsJs []byte

//go:embed elements.min.css.gz
var elementsCss []byte

func NewRouter(ctx context.Context, cfg any, openapiSpec []byte, lgr logger.Logger) (rtr *http.ServeMux) {

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
		_, _ = writer.Write(openapiSpec)
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
