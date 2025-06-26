package mid

import (
	"bytes"
	"net/http"
	"time"

	"github.com/clarktrimble/delish/buffered"
)

// LogResponse is a middleware which logs the response
func LogResponse(lgr logger, next http.Handler) http.HandlerFunc {

	return func(writer http.ResponseWriter, request *http.Request) {

		start := time.Now()
		ctx := request.Context()

		if SkipPaths[request.URL.Path] {
			// Stream directly, no buffering
			next.ServeHTTP(writer, request)
			//lgr.Debug(ctx, "streaming response", "path", request.URL.Path, "elapsed", time.Since(start))
			return
		}

		buf := &buffered.Buffered{
			Writer: writer,
			Buffer: bytes.Buffer{},
		}

		next.ServeHTTP(buf, request)

		fields := []any{
			"status", buf.Status,
			"headers", buf.Header(),
			"elapsed", time.Since(start),
		}

		if !SkipBody {
			fields = append(fields, "body")
			fields = append(fields, buf.Body())
		}

		lgr.Debug(ctx, "sending response", fields...)

		err := buf.WriteResponse()
		if err != nil {
			lgr.Error(ctx, "failed to write response", err)
		}
	}
}
