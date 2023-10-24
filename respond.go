package delish

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/pkg/errors"
)

// Respond provides convinience methods when responding to a request
type Respond struct {
	Writer http.ResponseWriter
	Logger Logger
}

// Ok responds with 200
func (rp *Respond) Ok(ctx context.Context) {

	header(rp.Writer, 200)
	rp.Write(ctx, []byte(`{"status":"ok"}`))
}

// NotOk logs an error and responds with it
func (rp *Respond) NotOk(ctx context.Context, code int, err error) {

	header(rp.Writer, code)

	rp.Logger.Error(ctx, "returning error to client", err)
	rp.WriteObjects(ctx, map[string]any{"error": fmt.Sprintf("%v", err)})
}

// NotFound responds with 404
func (rp *Respond) NotFound(ctx context.Context) {

	header(rp.Writer, 404)
	rp.Write(ctx, []byte(`{"not":"found"}`))
}

// WriteObjects responds with marshalled objects by key
func (rp *Respond) WriteObjects(ctx context.Context, objects map[string]any) {

	header(rp.Writer, 0)

	data, err := json.Marshal(objects)
	if err != nil {
		err = errors.Wrapf(err, "somehow failed to encode: %#v", objects)
		rp.Logger.Error(ctx, "failed to encode response", err)

		rp.Writer.WriteHeader(http.StatusInternalServerError)
		rp.Write(ctx, []byte(`{"error": "failed to encode response"}`))
	}

	rp.Write(ctx, data)
}

// Write respondes with arbitrary data, logging if error
func (rp *Respond) Write(ctx context.Context, data []byte) {

	// leaving content-type as exercise for handler

	_, err := rp.Writer.Write(data)
	if err != nil {
		err = errors.Wrapf(err, "failed to write response")
		rp.Logger.Error(ctx, "failed to write response", err)
	}
}

// unexported

func header(writer http.ResponseWriter, code int) {

	writer.Header().Set("content-type", "application/json")
	if code != 0 {
		writer.WriteHeader(code)
	}
}
