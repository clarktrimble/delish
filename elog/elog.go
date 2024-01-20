package elog

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"os"
	"strconv"
	"time"

	"github.com/clarktrimble/delish/elog/format"
	"github.com/clarktrimble/delish/elog/value"
	"github.com/pkg/errors"
)

// logger is flat as pancake
// no groups ..
// ctx ftw

type formatter interface {
	Format(ts time.Time, level, msg string, ctxFlds, flds map[string]value.Value) (data []byte, err error)
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
		//Formatter: &format.Lite{},
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
// - format wants to return reader
// - cfg lvl strings

func (ml *MinLog) Info(ctx context.Context, msg string, kv ...any) {
	ml.log(ctx, "info", msg, kv)
}

func (ml *MinLog) Error(ctx context.Context, msg string, err error, kv ...any) {

	kv = append(kv, "error", fmt.Sprintf("%+v", err))
	ml.log(ctx, "error", msg, kv)
}

func (ml *MinLog) WithFields(ctx context.Context, kv ...any) context.Context {

	fields := copyFields(ctx)
	for key, val := range toFieldsToo(kv) {
		fields[key] = val
	}

	ctx = context.WithValue(ctx, ctxKey{}, fields)
	return ctx
}

// unexported

func (ml *MinLog) log(ctx context.Context, level, msg string, kv []any) {

	line, err := ml.Formatter.Format(time.Now().UTC(), level, msg, getFields(ctx), toFieldsToo(kv))
	if err != nil {
		line = []byte(fmt.Sprintf("%s: %+v", logError, err))
	}

	// Todo: buff or sommat pls!!
	_, err = ml.Writer.Write(line)
	//_, err = ml.Writer.Write(append(line, []byte("\n")...))
	if err != nil && ml.AltWriter != nil {
		err = errors.Wrapf(err, "failed to write")
		_, _ = fmt.Fprintf(ml.AltWriter, "%s: %+v with line: %s\n", logError, err, line)
	}
}

type ctxKey struct{}

type fields map[string]value.Value

func toFieldsToo(kv []any) (flds fields) {

	// rfi: sync pool, json.encode

	flds = fields{}
	for _, attr := range argsToAttrSlice(kv) {

		val := value.Value{
			Data: []byte{},
		}

		switch attr.Value.Kind() {
		case slog.KindString:
			val.Data = strconv.AppendQuote(val.Data, attr.Value.String())
			val.Quoted = true
		case slog.KindBool:
			val.Data = strconv.AppendBool(val.Data, attr.Value.Bool())
		case slog.KindTime:
			val.Data = append(val.Data, '"')
			val.Data = attr.Value.Time().AppendFormat(val.Data, time.RFC3339)
			val.Data = append(val.Data, '"')
			val.Quoted = true
		case slog.KindDuration:
			val.Data = strconv.AppendInt(val.Data, int64(attr.Value.Duration()), 10)
		case slog.KindInt64:
			val.Data = strconv.AppendInt(val.Data, attr.Value.Int64(), 10)
		case slog.KindUint64:
			val.Data = strconv.AppendUint(val.Data, attr.Value.Uint64(), 10)
		case slog.KindFloat64:
			val.Data = strconv.AppendFloat(val.Data, attr.Value.Float64(), 'g', -1, 64)
		default:
			data, err := json.Marshal(attr.Value.Any())
			if err != nil {
				panic(err)
			}

			//val.Data = append(val.Data, '"')
			//val.Data = append(val.Data, data...)
			//val.Data = append(val.Data, '"')
			val.Data = strconv.AppendQuote(val.Data, string(data)) // escaped, woot!
			val.Quoted = true
		}

		flds[attr.Key] = val
	}

	return
}

func toFields(kv []any) (flds fields) {

	flds = fields{}
	for _, attr := range argsToAttrSlice(kv) {

		// Todo: switch here, but furst a knap!!

		val, err := toString(attr.Value)
		if err != nil {
			flds[logError] = value.NewFromString(err.Error())
			continue
		}

		// fmt.Printf(">>> toString: %s\n", val)

		if attr.Key == badKey {
			flds[logError] = value.NewFromString(fmt.Sprintf("no field name found for: %s", val))
			continue
		}

		flds[attr.Key] = value.NewFromString(val)
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
			"logerror": value.NewFromString(fmt.Sprintf("cannot assert fields on ctxval: %#v", val)),
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
