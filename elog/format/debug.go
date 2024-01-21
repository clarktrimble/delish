package format

import (
	"bytes"

	"github.com/clarktrimble/delish/elog/logmsg"
)

type Debug struct{}

func (dbg *Debug) Format(lm logmsg.LogMsg) ([]byte, error) {

	//now := ts.Format("15:04:05.0000")

	buf := &bytes.Buffer{}

	buf.WriteString("msg: ")
	buf.WriteString(lm.Msg)
	buf.WriteString("\n")

	lines(buf, "ctx ", lm.CtxFields)
	lines(buf, "", lm.Fields)

	return buf.Bytes(), nil
}

func lines(buf *bytes.Buffer, prefix string, flds logmsg.Fields) {

	for key, val := range flds {
		buf.WriteString(prefix)
		buf.WriteString(key)
		buf.WriteString(": ")
		buf.Write(val.Data)
		buf.WriteString("\n")
	}
}
