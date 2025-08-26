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

		if skipLogging(request) {
			next.ServeHTTP(writer, request)
			return
		}

		start := time.Now()
		ctx := request.Context()
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
