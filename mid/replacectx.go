package mid

import (
	"context"
	"net/http"
)

// ReplaceCtx replaces the request ctx
func ReplaceCtx(ctx context.Context, next http.Handler) http.HandlerFunc {

	return func(writer http.ResponseWriter, request *http.Request) {

		next.ServeHTTP(writer, request.WithContext(ctx))
	}
}
