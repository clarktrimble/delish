package mid

import (
	"context"
	"net/http"
)

// ReplaceCtx replaces the request ctx.
//
// This can give logging middlewares access to contextual logging fields
// such as "app_id" and "run_id".
// And of course attachs any cancel or timeouts associated with the new ctx.
func ReplaceCtx(ctx context.Context, next http.Handler) http.HandlerFunc {

	return func(writer http.ResponseWriter, request *http.Request) {

		next.ServeHTTP(writer, request.WithContext(ctx))
	}
}
