package logmsg

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"github.com/pkg/errors"
)

// rfi:
// - pointerize Value and it's Data in support of Write method, or just keep appending?
// - Value sync.Pool ftw or at least know len for Marshall
// - think about more/diff Value metadata in support of depthy obj logging in Marshal
// - https://dave.cheney.net/2016/04/27/dont-just-check-errors-handle-them-gracefully (behave!)

const (
	ErrorKey string = "logerror"
)

var NoValueError = errors.New("no value set") //nolint: errname // not now

type LogMsg struct {
	Ts        time.Time
	Level     string
	Msg       string
	CtxFields Fields
	Fields    Fields
}

type Fields map[string]Value

func (fields Fields) Marshal() []byte {

	data := make([]byte, 0, 1024)

	data = append(data, '{')
	for key, val := range fields {
		data = strconv.AppendQuote(data, key)
		data = append(data, ':')
		data = append(data, val.Data...)
		data = append(data, ',')
	}
	data[len(data)-1] = '}'

	return data
}

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
