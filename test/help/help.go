// Package help for tests
package help

import (
	"fmt"
	"net/http"
)

// Todo: worth pkg here?

type ErrorResponder struct{}

func (buf *ErrorResponder) Header() (hdr http.Header) {

	hdr = http.Header{}
	return
}

func (buf *ErrorResponder) Write(body []byte) (count int, err error) {

	err = fmt.Errorf("oops")
	return
}

func (buf *ErrorResponder) WriteHeader(status int) {}
