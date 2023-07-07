package mid

import (
	"bytes"
	"net/http"
	"time"

	"github.com/clarktrimble/delish/buffered"
)

// LogResponse is a middleware which logs the response
func LogResponse(lgr Logger, next http.Handler) http.HandlerFunc {

	return func(writer http.ResponseWriter, request *http.Request) {

		start := time.Now()
		ctx := request.Context()
		buf := &buffered.Buffered{
			Writer: writer,
			Buffer: bytes.Buffer{},
		}

		next.ServeHTTP(buf, request)

		lgr.Info(ctx, "sending response",
			"status", buf.Status,
			"headers", buf.Header(),
			"body", buf.Body(),
			"elapsed", time.Since(start),
		)
		// Todo: opt-out body logging

		err := buf.WriteResponse()
		if err != nil {
			lgr.Error(ctx, "failed to write response", err)
		}
	}
}
