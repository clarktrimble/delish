package format

import (
	"encoding/json"
	"time"

	"github.com/pkg/errors"

	"github.com/clarktrimble/delish/elog/logmsg"
)

type Json struct{}

func (jsn *Json) Format(lm logmsg.LogMsg) (data []byte, err error) {

	// overwrite duplicate "line" fields from ctx and boilerplate

	fields := lm.Fields
	for key, val := range lm.CtxFields {
		fields[key] = val
	}

	fields[jsonMsgKey] = logmsg.NewValue(lm.Msg)
	fields[jsonLevelKey] = logmsg.NewValue(lm.Level)
	fields[jsonTsKey] = logmsg.NewValue(lm.Ts.Format(time.RFC3339))

	data, err = json.Marshal(fields)
	if err != nil {
		err = errors.Wrap(err, "failed to marshal log message")
		return
	}

	data = append(data, '\n')
	return
}

// unexported

const (
	jsonMsgKey   string = "msg"
	jsonLevelKey string = "level"
	jsonTsKey    string = "ts"
)
