package format

import (
	"encoding/json"
	"time"

	"github.com/clarktrimble/delish/elog/value"
	"github.com/pkg/errors"
)

type Json struct{}

func (jsn *Json) Format(ts time.Time, level, msg string, ctxFlds, flds map[string]value.Value) (data []byte, err error) {
	//Format(ts time.Time, level, msg string, ctxFlds, flds map[string]value.Value) (data []byte, err error)

	// silently overwrite line fields from ctx and boilerplate when duplicate key

	for key, val := range ctxFlds {
		flds[key] = val
	}

	// Todo: config boiler keys
	flds["msg"] = value.NewFromString(msg)
	flds["level"] = value.NewFromString(level)
	flds["ts"] = value.NewFromString(ts.Format(time.RFC3339))

	//for key, val := range flds {
	//fmt.Printf(">>> %s :: %s ::\n", key, val)
	//}

	data, err = json.Marshal(flds)
	if err != nil {
		err = errors.Wrap(err, "failed to marshal log message")
		return
	}

	data = append(data, '\n')
	return
}
