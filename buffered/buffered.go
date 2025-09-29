package buffered

/**
 * inspired by Alex Kozadaev's https://bitbucket.org/snobb/susanin
**/

import (
	"bytes"
	"net/http"

	"github.com/pkg/errors"
)

// Buffered implements http.ResponseWriter
// buffering the response and providing access to the body
type Buffered struct {
	Writer http.ResponseWriter
	Status int
	Buffer bytes.Buffer
}

// Header returns header
func (buf *Buffered) Header() http.Header {

	return buf.Writer.Header()
}

// Write buffers the response
func (buf *Buffered) Write(body []byte) (int, error) {

	if buf.Status == 0 {
		buf.Status = 200
	}

	return buf.Buffer.Write(body)
}

// WriteHeader stores the status code
func (buf *Buffered) WriteHeader(status int) {

	buf.Status = status
}

// Body gets the buffered response body
func (buf *Buffered) Body() string {

	return buf.Buffer.String()
}

// WriteResponse writes to the response writer
func (buf *Buffered) WriteResponse() (err error) {

	buf.Writer.WriteHeader(buf.Status)

	_, err = buf.Writer.Write(buf.Buffer.Bytes())
	err = errors.Wrapf(err, "failed to write response")
	return
}

/*
func (buf *Buffered) Flush() {
	// Not suitable for streaming!!
	// Todo: perhaps better off w/o Flush at all??
	// Todo: unit
	buf.WriteHeader(buf.Status)
	_, _ = buf.Writer.Write(buf.Buffer.Bytes())
	buf.Buffer.Reset()
}
*/
