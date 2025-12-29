package mid

import (
	"bytes"
	"io"
	"net/http"
	"net/url"
	"strings"

	"github.com/clarktrimble/delish/logger"
	"github.com/clarktrimble/hondo"
	"github.com/pkg/errors"
)

const (
	idLen int = 7
)

// LogRequest is a middleware which logs the request.
func LogRequest(lgr logger.Logger, next http.Handler) http.HandlerFunc {

	return func(writer http.ResponseWriter, request *http.Request) {

		if skipLogging(request) {
			next.ServeHTTP(writer, request)
			return
		}

		ctx := request.Context()
		ctx = lgr.WithFields(ctx, "request_id", hondo.Rand(idLen))
		request = request.WithContext(ctx)

		ip, port := ipPort(request.RemoteAddr)
		path, query := pathQuery(request.URL)

		fields := []any{
			"method", request.Method,
			"path", path,
			"query", query,
			"remote_ip", ip,
			"remote_port", port,
			"headers", redact(request.Header),
		}

		// Todo: config skip body yeah??
		if !SkipBody {
			body, err := requestBody(request)
			if err != nil {
				lgr.Error(ctx, "request logger failed to get body", err)
			} else {
				fields = append(fields, "body")
				fields = append(fields, string(body))
			}
		}

		lgr.Trace(ctx, "received request", fields...)
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

// read and restore body

func requestBody(req *http.Request) (body []byte, err error) {

	body, err = read(req.Body)
	if err != nil {
		return
	}

	req.Body = io.NopCloser(bytes.NewBuffer(body))
	return
}

func read(reader io.Reader) (data []byte, err error) {

	if reader == nil {
		return
	}

	data, err = io.ReadAll(reader)
	if err != nil {
		err = errors.Wrapf(err, "failed to read from: %#v", reader)
		return
	}
	if len(data) == 0 {
		data = nil
	}

	return
}
