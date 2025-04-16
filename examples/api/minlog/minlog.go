// Package minlog implements a minimal logger for development.
//
// See https://github.com/clarktrimble/sabot for json output and more.
package minlog

import (
	"context"
	"encoding/json"
	"fmt"
	"sort"
	"strings"
	"time"
)

// MinLog is a logger.
type MinLog struct{}

// Info logs info.
func (ml *MinLog) Info(ctx context.Context, msg string, kv ...any) {

	log(ctx, ">", msg, kv)
}

// Debug logs info.
func (ml *MinLog) Debug(ctx context.Context, msg string, kv ...any) {

	log(ctx, "-", msg, kv)
}

// Error logs an error.
func (ml *MinLog) Error(ctx context.Context, msg string, err error, kv ...any) {

	kv = append([]any{"error", err}, kv...)
	log(ctx, "*", msg, kv)
}

// WithFields adds fields to the returned context.
func (ml *MinLog) WithFields(ctx context.Context, kv ...any) context.Context {

	fields := copyFields(ctx)
	for key, val := range toFields(kv) {
		fields[key] = val
	}

	ctx = context.WithValue(ctx, ctxKey{}, fields)
	return ctx
}

// SetLevel is not demo'd here.
func (ml *MinLog) SetLevel(ctx context.Context, level string) (err error) {
	// noop
	return
}

// unexported

type ctxKey struct{}

func log(ctx context.Context, sep, msg string, kv []any) {

	now := time.Now().UTC().Format("15:04:05.0000")
	fromCtx := strings.Join(getFields(ctx).pairs(), "  ")

	fmt.Printf("%s %s %s | %s\n", now, sep, msg, fromCtx)

	for _, pair := range toFields(kv).pairs() {
		fmt.Printf("                %s\n", pair)
	}
}

type fields map[string]string

func toFields(kv []any) (flds fields) {

	if len(kv)%2 != 0 {
		kv = append(kv, "odd kv padding")
	}

	flds = map[string]string{}
	for i := 0; i < len(kv); i += 2 {
		flds[fmt.Sprintf("%s", kv[i])] = toString(kv[i+1])
	}

	return
}

func toString(val any) string {

	var out string

	switch val.(type) {
	case string, []byte, int, int64, float64, error, time.Time, time.Duration:
		out = fmt.Sprintf("%v", val)
	default:
		data, err := json.Marshal(val)
		if err != nil {
			return fmt.Sprintf("log error: failed to marshal: %#v", val)
		}
		out = string(data)
	}

	return out
}

func (flds fields) pairs() (pairs []string) {

	pairs = []string{}

	for key, val := range flds {
		pairs = append(pairs, fmt.Sprintf("%s: %s", key, val))
	}
	sort.Strings(pairs)

	return
}

func getFields(ctx context.Context) fields {

	val := ctx.Value(ctxKey{})
	if val == nil {
		return fields{}
	}

	return val.(fields)
}

func copyFields(ctx context.Context) fields {

	flds := fields{}
	for key, val := range getFields(ctx) {
		flds[key] = val
	}

	return flds
}
