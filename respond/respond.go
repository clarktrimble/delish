// Package respond provides logging help when responding to a json request.
package respond

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/pkg/errors"
)

// Respond provides convinience methods when responding to a json request.
type Respond struct {
	Writer http.ResponseWriter
	Logger logger
}

// Ok responds with 200 ok.
func (rp *Respond) Ok(ctx context.Context) {

	rp.header(200)
	rp.Write(ctx, []byte(`{"status":"ok"}`))
}

// NotOk logs an error and responds with it.
func (rp *Respond) NotOk(ctx context.Context, code int, err error) {

	rp.header(code)

	rp.Logger.Error(ctx, "returning error to client", err)
	rp.WriteObjects(ctx, map[string]any{"error": fmt.Sprintf("%v", err)})
}

// NotFound responds with 404 not found.
func (rp *Respond) NotFound(ctx context.Context) {

	rp.header(404)
	rp.Write(ctx, []byte(`{"not":"found"}`))
}

// WriteObjects responds with marshalled objects by key.
func (rp *Respond) WriteObjects(ctx context.Context, objects map[string]any) {

	rp.header(0)

	data, err := json.Marshal(objects)
	if err != nil {
		err = errors.Wrapf(err, "somehow failed to encode: %#v", objects)
		rp.Logger.Error(ctx, "failed to encode response", err)

		rp.header(500)
		rp.Write(ctx, []byte(`{"error": "failed to encode response"}`))
		return
	}

	rp.Write(ctx, data)
}

// Write respondes with arbitrary data, logging if error.
func (rp *Respond) Write(ctx context.Context, data []byte) {

	// leaving content-type as exercise for handler

	_, err := rp.Writer.Write(data)
	if err != nil {
		err = errors.Wrapf(err, "failed to write response")
		rp.Logger.Error(ctx, "failed to write response", err)
	}
}

// unexported

type logger interface {
	Error(ctx context.Context, msg string, err error, kv ...any)
}

func (rp *Respond) header(code int) {

	rp.Writer.Header().Set("content-type", "application/json")
	if code != 0 {
		rp.Writer.WriteHeader(code)
	}
}
