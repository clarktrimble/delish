package format

import (
	"encoding/json"
	"time"

	"github.com/pkg/errors"
)

type Json struct{}

func (jsn *Json) Format(ts time.Time, level, msg string, ctxFlds, flds map[string]string) (data []byte, err error) {

	// silently overwrite line fields from ctx and boilerplate when duplicate key

	for key, val := range ctxFlds {
		flds[key] = val
	}

	// Todo: config boiler keys
	flds["msg"] = msg
	flds["level"] = level
	flds["ts"] = ts.Format(time.RFC3339)

	data, err = json.Marshal(flds)
	err = errors.Wrap(err, "failed to marshal log message")
	return
}
