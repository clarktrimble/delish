package elog

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"os"
	"strconv"
	"time"

	"github.com/pkg/errors"

	"github.com/clarktrimble/delish/elog/format"
	"github.com/clarktrimble/delish/elog/logmsg"
)

// logger is flat as pancake
// no groups ..
// ctx ftw
// performance, no (never paid for such?) is interesting though!

type formatter interface {
	Format(logMsg logmsg.LogMsg) ([]byte, error)
}

const (
	trunc string = "--truncated--"
)

type MinLog struct {
	Writer    io.Writer
	AltWriter io.Writer
	Formatter formatter
	Escape    bool
	MaxLen    int
	InfoStr   string
	ErrorStr  string
}

func New() *MinLog {

	return &MinLog{
		Writer:    os.Stdout,
		AltWriter: os.Stderr,
		Formatter: &format.Lite{},
		InfoStr:   ">",
		ErrorStr:  "*",
		//Formatter: &format.Json{},
		//Escape:    true,
		//InfoStr:   "info",
		//ErrorStr:  "error",
	}
}

// off to a decent start with slog.Attr
//
// rfi:
// x convert to string right away, dont put in ctx and dont send to formatter
// x switch back to map[string] in ctx store (prolly need copy with map?)
// x dont forget to look over logerror's, at min a field would be nice
// - put error trace in it's own field
// x look at map[string][]byte ??
// - format wants to return reader
// - cfg lvl strings
// - sync pool values, json.encode
// - trunc

func (ml *MinLog) Info(ctx context.Context, msg string, kv ...any) {
	ml.log(ctx, ml.InfoStr, msg, kv)
}

func (ml *MinLog) Error(ctx context.Context, msg string, err error, kv ...any) {

	kv = append(kv, "error", fmt.Sprintf("%+v", err))
	ml.log(ctx, ml.ErrorStr, msg, kv)
}

func (ml *MinLog) WithFields(ctx context.Context, kv ...any) context.Context {

	fields := logmsg.CopyFields(ctx)
	for key, val := range ml.toFields(kv) {
		fields[key] = val
	}

	ctx = fields.Store(ctx)
	return ctx
}

// unexported

func (ml *MinLog) log(ctx context.Context, level, msg string, kv []any) {

	lm := logmsg.LogMsg{
		Ts:        time.Now().UTC(),
		Level:     level,
		Msg:       msg,
		CtxFields: logmsg.GetFields(ctx),
		Fields:    ml.toFields(kv),
	}

	line, err := ml.Formatter.Format(lm)
	if err != nil {
		line = []byte(fmt.Sprintf("%s: %+v", logmsg.ErrorKey, err))
	}

	_, err = ml.Writer.Write(line)
	if err != nil && ml.AltWriter != nil {
		err = errors.Wrapf(err, "failed to write")
		_, _ = fmt.Fprintf(ml.AltWriter, "%s: %+v with line: %s\n", logmsg.ErrorKey, err, line)
	}
}

func (ml *MinLog) toFields(kv []any) (flds logmsg.Fields) {

	flds = logmsg.Fields{}
	for _, attr := range argsToAttrSlice(kv) {

		val, err := ml.toValue(attr.Value)
		if err != nil {
			flds[logmsg.ErrorKey] = logmsg.NewValue(err.Error())
			continue
		}

		flds[attr.Key] = val
	}

	return
}

func (ml *MinLog) toValue(attrValue slog.Value) (val logmsg.Value, err error) {

	val = logmsg.Value{
		Data: []byte{},
	}

	switch attrValue.Kind() {
	case slog.KindString:
		val.Data = strconv.AppendQuote(val.Data, attrValue.String())
		val.Quoted = true
	case slog.KindBool:
		val.Data = strconv.AppendBool(val.Data, attrValue.Bool())
	case slog.KindTime:
		val.Data = append(val.Data, '"')
		val.Data = attrValue.Time().AppendFormat(val.Data, time.RFC3339)
		val.Data = append(val.Data, '"')
		val.Quoted = true
	case slog.KindDuration:
		val.Data = strconv.AppendInt(val.Data, int64(attrValue.Duration()), 10)
	case slog.KindInt64:
		val.Data = strconv.AppendInt(val.Data, attrValue.Int64(), 10)
	case slog.KindUint64:
		val.Data = strconv.AppendUint(val.Data, attrValue.Uint64(), 10)
	case slog.KindFloat64:
		val.Data = strconv.AppendFloat(val.Data, attrValue.Float64(), 'g', -1, 64)
	default:
		err = (&val).MarshalAppend(attrValue.Any(), ml.Escape)
	}

	return
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
