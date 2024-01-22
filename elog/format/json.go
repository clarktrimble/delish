package format

import (
	"time"

	"github.com/clarktrimble/delish/elog/logmsg"
)

// rfi:
// - Value sync.Pool ftw or at least know len for Marshall
// - prolly append ts, msg, level here, just get frags from fields "Marshal"

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

	data = fields.Marshal()

	data = append(data, '\n')
	return
}

// unexported

const (
	jsonMsgKey   string = "msg"
	jsonLevelKey string = "level"
	jsonTsKey    string = "ts"
)
