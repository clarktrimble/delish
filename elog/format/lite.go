package format

import (
	"fmt"
	"sort"
	"strings"

	"github.com/clarktrimble/delish/elog/logmsg"
)

type Lite struct{}

func (lt *Lite) Format(lm logmsg.LogMsg) ([]byte, error) {

	bldr := &strings.Builder{}
	now := lm.Ts.Format("15:04:05.0000")
	fromCtx := strings.Join(pairs(lm.CtxFields), "  ")

	fmt.Fprintf(bldr, "%s %s %s | %s\n", now, lm.Level, lm.Msg, fromCtx)
	for _, pair := range pairs(lm.Fields) {
		fmt.Fprintf(bldr, "                %s\n", pair)
	}

	return []byte(bldr.String()), nil
}

func pairs(flds logmsg.Fields) (pairs []string) {

	pairs = []string{}

	for key, val := range flds {

		data := val.Data
		if val.Quoted && len(data) > 1 {
			data = data[1 : len(data)-1]
		}

		pairs = append(pairs, fmt.Sprintf("%s: %s", key, data))
	}
	sort.Strings(pairs)

	return
}
