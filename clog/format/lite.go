package format

import (
	"fmt"
	"sort"
	"strings"
	"time"
)

type Lite struct{}

func (lt *Lite) Format(ts time.Time, level, msg string, ctxFlds, flds map[string]string) []byte {

	bldr := &strings.Builder{}
	now := ts.Format("15:04:05.0000")
	fromCtx := strings.Join(pairs(ctxFlds), "  ")

	sep := ">"
	if level == "error" {
		sep = "*"
	}

	fmt.Fprintf(bldr, "%s %s %s | %s\n", now, sep, msg, fromCtx)
	for _, pair := range pairs(flds) {
		fmt.Fprintf(bldr, "                %s\n", pair)
	}

	return []byte(bldr.String())
}

func pairs(flds map[string]string) (pairs []string) {

	pairs = []string{}

	for key, val := range flds {
		pairs = append(pairs, fmt.Sprintf("%s: %s", key, val))
	}
	sort.Strings(pairs)

	return
}
