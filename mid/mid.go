package mid

import (
	"net/http"
	"regexp"
)

// Todo: canonicalize redact headers, see giant for example
// Todo: replace pkg vars maybe with "mid.NewLogger(lgr, opts)" (sets up both)

var (
	RedactHeaders = map[string]bool{}
	SkipPattern   *regexp.Regexp
	SkipBody      bool
)

func skipLogging(request *http.Request) bool {

	// Todo: unit
	// Todo: log just a little? body is really the heavy lift here
	// lgr.Trace(ctx, "streaming response", "path", request.URL.Path, "elapsed", time.Since(start))

	return SkipPattern != nil &&
		request.URL != nil &&
		SkipPattern.MatchString(request.URL.Path)
}
