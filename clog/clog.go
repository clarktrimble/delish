package clog

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"os"
	"time"

	"github.com/clarktrimble/delish/clog/format"
	"github.com/pkg/errors"
)

type formatter interface {
	Format(ts time.Time, level, msg string, ctxFlds, flds map[string]string) (data []byte, err error)
}

const (
	logError string = "logerror"
	trunc    string = "--truncated--"
)

type MinLog struct {
	Writer    io.Writer
	AltWriter io.Writer
	Formatter formatter
	MaxLen    int
}

func New() *MinLog {

	return &MinLog{
		Writer:    os.Stdout,
		AltWriter: os.Stderr,
		Formatter: &format.Json{},
	}
}

// off to a decent start with slog.Attr
//
// rfi:
// x convert to string right away, dont put in ctx and dont send to formatter
// x switch back to map[string] in ctx store (prolly need copy with map?)
// x dont forget to look over logerror's, at min a field would be nice
// - put error trace in it's own field
// - look at map[string][]byte ??

func (ml *MinLog) Info(ctx context.Context, msg string, kv ...any) {
	ml.log(ctx, "info", msg, kv)
}

func (ml *MinLog) Error(ctx context.Context, msg string, err error, kv ...any) {

	kv = append(kv, "error", fmt.Sprintf("%+v", err))
	ml.log(ctx, "error", msg, kv)
}

func (ml *MinLog) WithFields(ctx context.Context, kv ...any) context.Context {

	fields := copyFields(ctx)
	for key, val := range toFields(kv) {
		fields[key] = val
	}

	ctx = context.WithValue(ctx, ctxKey{}, fields)
	return ctx
}

// unexported

func (ml *MinLog) log(ctx context.Context, level, msg string, kv []any) {

	line, err := ml.Formatter.Format(time.Now().UTC(), level, msg, getFields(ctx), toFields(kv))
	if err != nil {
		line = []byte(fmt.Sprintf("%s: %+v", logError, err))
	}

	// Todo: buff or sommat pls!!
	_, err = ml.Writer.Write(append(line, []byte("\n")...))
	if err != nil && ml.AltWriter != nil {
		err = errors.Wrapf(err, "failed to write")
		_, _ = fmt.Fprintf(ml.AltWriter, "%s: %+v with line: %s\n", logError, err, line)
	}
}

type ctxKey struct{}

type fields map[string]string

func toFields(kv []any) (flds fields) {

	flds = fields{}
	for _, attr := range argsToAttrSlice(kv) {

		val, err := toString(attr.Value)
		if err != nil {
			flds[logError] = err.Error()
			continue
		}

		if attr.Key == badKey {
			flds[logError] = fmt.Sprintf("no field name found for: %s", val)
			continue
		}

		flds[attr.Key] = val
	}

	return
}

func toString(val slog.Value) (out string, err error) {

	switch val.Kind() {
	case slog.KindString, slog.KindBool, slog.KindDuration, slog.KindTime,
		slog.KindInt64, slog.KindUint64, slog.KindFloat64:
		out = val.String()
	default:
		var data []byte
		data, err = json.Marshal(val.Any())
		if err != nil {
			err = errors.Wrapf(err, "failed to marshal value: %#v", val.Any())
			return
		}
		out = string(data)
	}

	return
}

func getFields(ctx context.Context) fields {

	val := ctx.Value(ctxKey{})
	if val == nil {
		return fields{}
	}

	flds, ok := val.(fields)
	if !ok {
		return fields{
			"logerror": fmt.Sprintf("cannot assert fields on ctxval: %#v", val),
		}
	}

	return flds
}

func copyFields(ctx context.Context) fields {

	flds := fields{}
	for key, val := range getFields(ctx) {
		flds[key] = val
	}

	return flds
}

// copied from the olde slog as of go.1.21.6

const badKey = "!BADKEY"

func argsToAttrSlice(args []any) []slog.Attr {

	var (
		attr  slog.Attr
		attrs []slog.Attr
	)

	for len(args) > 0 {
		attr, args = argsToAttr(args)
		attrs = append(attrs, attr)
	}

	return attrs
}

func argsToAttr(args []any) (slog.Attr, []any) {

	switch x := args[0].(type) {
	case string:
		if len(args) == 1 {
			return slog.String(badKey, x), nil
		}
		return slog.Any(x, args[1]), args[2:]

	case slog.Attr:
		return x, args[1:]

	default:
		return slog.Any(badKey, x), args[1:]
	}
}
