package mid

import (
	"net/http"
)

// Todo: this might be the way re-streaming?? cleanup!

//type ResponseWriter interface {
//Header() Header
//Write([]byte) (int, error)
//WriteHeader(statusCode int)
//}

// Streaming captures response meta.
type Streaming struct {
	writer http.ResponseWriter
	status int
	size   int
	err    error
}

// NewStreaming creates a Streaming.
func NewStreaming(writer http.ResponseWriter) *Streaming {
	return &Streaming{
		writer: writer,
		status: 200,
	}
}

// implement http.ResponseWriter

// Header returns header.
func (str *Streaming) Header() http.Header {
	return str.writer.Header()
}

// Write writes to the writer.
func (str *Streaming) Write(body []byte) (int, error) {
	size, err := str.writer.Write(body)
	if err != nil {
		str.err = err
	}

	str.size = size // Todo: wha??
	return size, nil
}

// WriteHeader stores the status code and writes headers
func (str *Streaming) WriteHeader(status int) {

	str.status = status
	str.writer.WriteHeader(status)
}

// end implement http.ResponseWriter

//func (rw *responseWriter) Write(b []byte) (int, error) {
//size, err := rw.ResponseWriter.Write(b)
//rw.size += size
//return size, err
//}

// Status returns the captured status code
func (str *Streaming) Status() int {

	return str.status
}

// Body returns blank as streaming does not support.
//func (str *Streaming) Body() string {
//return ""
//}
