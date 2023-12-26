package mid

import (
	"net/http"
	"net/url"
	"strings"

	"github.com/clarktrimble/hondo"
)

const (
	idLen int = 7
)

var (
	RedactHeaders = map[string]bool{}
)

// LogRequest is a middleware which logs the request.
func LogRequest(logger Logger, next http.Handler) http.HandlerFunc {

	return func(writer http.ResponseWriter, request *http.Request) {

		ctx := request.Context()
		ctx = logger.WithFields(ctx, "request_id", hondo.Rand(idLen))
		request = request.WithContext(ctx)

		body, err := requestBody(request)
		if err != nil {
			logger.Error(ctx, "request logger failed to get body", err)
		}

		ip, port := ipPort(request.RemoteAddr)
		path, query := pathQuery(request.URL)

		logger.Info(ctx, "received request",
			"method", request.Method,
			"path", path,
			"query", query,
			"body", string(body),
			"remote_ip", ip,
			"remote_port", port,
			"headers", redact(request.Header),
		)

		next.ServeHTTP(writer, request)
	}
}

// unexported

func redact(header http.Header) (redacted http.Header) {

	redacted = header.Clone()
	for key := range header {

		redacted[key] = header[key]
		if RedactHeaders[key] {
			redacted[key] = []string{"--redacted--"}
		}
	}

	return
}

func ipPort(addr string) (ip, port string) {

	ipPort := strings.Split(addr, ":")
	ip = ipPort[0]
	if len(ipPort) > 1 {
		port = ipPort[1]
	}

	return
}

func pathQuery(url *url.URL) (path string, query map[string][]string) {

	if url != nil {
		path = url.Path
		query = url.Query()
	}

	return
}
