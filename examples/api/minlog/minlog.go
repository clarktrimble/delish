// Package minlog implements a (sub)minimal logger
// see https://github.com/clarktrimble/sabot for a featureful implementation
package minlog

import (
	"context"
	"fmt"
)

type MinLog struct{}

func (ml *MinLog) Info(ctx context.Context, msg string, kv ...any) {

	fmt.Printf("msg > %s %s\n", msg, fields(ctx))
	if len(kv) > 0 {
		fmt.Printf("kvs > %s\n\n", keyvals(kv))
	}
}

func (ml *MinLog) Error(ctx context.Context, msg string, err error, kv ...any) {

	fmt.Printf("err > %s %+v\n", msg, err)
}

func (ml *MinLog) WithFields(ctx context.Context, kv ...any) context.Context {

	val := fmt.Sprintf("%s %s", fields(ctx), keyvals(kv))
	ctx = context.WithValue(ctx, key{}, val)

	return ctx
}

// unexported

type key struct{}

func keyvals(kv []any) string {
	str := ""
	for _, item := range kv {
		str = fmt.Sprintf("%s::%v", str, item)
	}

	return str
}

func fields(ctx context.Context) string {
	val := ctx.Value(key{})
	if val == nil {
		val = ""
	}

	return fmt.Sprintf("%s", val)
}
