// Package respond provides logging help when responding to a json request.
package respond

import (
	"context"
	"encoding/json"
	"html/template"
	"net/http"

	"github.com/clarktrimble/delish/logger"
	"github.com/pkg/errors"
)

// Todo: think about respond as it's own tiny module

// Respond provides convinience methods when responding to a json request.
type Respond struct {
	Writer http.ResponseWriter
	Logger logger.Logger
}

// New creates a Respond.
func New(writer http.ResponseWriter, lgr logger.Logger) *Respond {
	// Todo: unit!!

	return &Respond{
		Writer: writer,
		Logger: lgr,
	}
}

// Ok responds with 200 ok.
func (rp *Respond) Ok(ctx context.Context) {

	rp.jsonHeader(200)
	rp.Write(ctx, []byte(`{"status":"ok"}`))
}

// NotOk logs an error and responds with it.
func (rp *Respond) NotOk(ctx context.Context, code int, err error) {

	rp.jsonHeader(code)

	rp.Logger.Error(ctx, "returning error to client", err)
	rp.WriteObjects(ctx, map[string]any{"error": err.Error()})
}

// GoNoGo calls NotOk or Ok.
func (rp *Respond) GoNoGo(ctx context.Context, code int, err error) {
	// Todo: unit!!

	if err != nil {
		rp.NotOk(ctx, code, err)
		return
	}

	rp.Ok(ctx)
}

// NotFound responds with 404 not found.
func (rp *Respond) NotFound(ctx context.Context) {

	rp.jsonHeader(404)
	rp.Write(ctx, []byte(`{"not":"found"}`))
}

// WriteObjects responds with marshalled objects by key.
func (rp *Respond) WriteObjects(ctx context.Context, objects map[string]any) {

	rp.jsonHeader(0)

	data, err := json.Marshal(objects)
	if err != nil {
		err = errors.Wrapf(err, "somehow failed to encode: %#v", objects)
		rp.Logger.Error(ctx, "failed to encode response", err)

		rp.jsonHeader(500)
		rp.Write(ctx, []byte(`{"error": "failed to encode response"}`))
		return
	}

	rp.Write(ctx, data)
}

// WriteObject responds with marshalled object.
func (rp *Respond) WriteObject(ctx context.Context, obj any) {

	// Todo: no need for "Objects" version, yeah?
	//       in any case, unit!

	rp.jsonHeader(0)

	data, err := json.Marshal(obj)
	if err != nil {
		err = errors.Wrapf(err, "somehow failed to encode: %#v", obj)
		rp.Logger.Error(ctx, "failed to encode response", err)

		rp.jsonHeader(500)
		rp.Write(ctx, []byte(`{"error": "failed to encode response"}`))
		return
	}

	rp.Write(ctx, data)
}

func (rp *Respond) WriteHtml(ctx context.Context, content template.HTML) {

	// Todo: want string and sanitize here?
	//       in any case, unit!

	rp.Writer.Header().Set("Content-Type", "text/html")
	rp.Write(ctx, []byte(content))
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

func (rp *Respond) jsonHeader(code int) {

	rp.Writer.Header().Set("content-type", "application/json")
	if code != 0 {
		rp.Writer.WriteHeader(code)
	}
}
