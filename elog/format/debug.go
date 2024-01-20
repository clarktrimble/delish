package format

import (
	"bytes"
	"time"

	"github.com/clarktrimble/delish/elog/value"
)

type Debug struct{}

func (lt *Debug) Format(ts time.Time, level, msg string, ctxFlds, flds map[string]value.Value) ([]byte, error) {

	//now := ts.Format("15:04:05.0000")

	buf := &bytes.Buffer{}

	buf.WriteString("msg: ")
	buf.WriteString(msg)
	buf.WriteString("\n")

	lines(buf, "ctx ", ctxFlds)
	lines(buf, "", flds)

	return buf.Bytes(), nil
}

func lines(buf *bytes.Buffer, prefix string, flds map[string]value.Value) {

	for key, val := range flds {
		buf.WriteString(prefix)
		buf.WriteString(key)
		buf.WriteString(": ")
		buf.Write(val.Data)
		buf.WriteString("\n")
	}
}
