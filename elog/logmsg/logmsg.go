package logmsg

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"github.com/pkg/errors"
)

const (
	ErrorKey string = "logerror"
)

var NoValueError = errors.New("no value set")

type LogMsg struct {
	Ts        time.Time
	Level     string
	Msg       string
	CtxFields Fields
	Fields    Fields
}

type Fields map[string]Value

func (fields Fields) Store(ctx context.Context) context.Context {
	return context.WithValue(ctx, ctxKey{}, fields)
}

func GetFields(ctx context.Context) Fields {

	val := ctx.Value(ctxKey{})
	if val == nil {
		return Fields{}
	}

	flds, ok := val.(Fields)
	if !ok {
		return Fields{
			ErrorKey: NewValue(fmt.Sprintf("cannot assert fields on ctxval: %#v", val)),
		}
	}

	return flds
}

func CopyFields(ctx context.Context) Fields {

	flds := Fields{}
	for key, val := range GetFields(ctx) {
		flds[key] = val
	}

	return flds
}

type Value struct {
	Data   []byte
	Quoted bool
}

func NewValue(str string) (val Value) {

	val = Value{
		Data: make([]byte, 0, len(str)+2),
	}
	val.Data = strconv.AppendQuote(val.Data, str)
	val.Quoted = true

	return
}

func (vl Value) MarshalJSON() ([]byte, error) {

	if vl.Data == nil {
		return []byte{}, NoValueError
	}

	return vl.Data, nil
}

func (vl *Value) MarshalAppend(obj any, escape bool) error {

	data, err := json.Marshal(obj)
	if errors.Is(err, NoValueError) {
		vl.Data = append(vl.Data, []byte("null")...)
		return nil
	}
	if err != nil {
		err = errors.Wrapf(err, "failed to marshal obj: %#v\n", obj)
		return err
	}

	if escape {
		vl.Data = strconv.AppendQuote(vl.Data, string(data))
	} else {
		vl.Data = append(vl.Data, '"')
		vl.Data = append(vl.Data, data...)
		vl.Data = append(vl.Data, '"')
	}
	vl.Quoted = true

	return nil
}

// unexported

type ctxKey struct{}
