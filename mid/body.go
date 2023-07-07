package mid

import (
	"bytes"
	"io"
	"net/http"

	"github.com/pkg/errors"
)

// unexported

// read and restore body

func requestBody(req *http.Request) (body []byte, err error) {

	body, err = read(req.Body)
	if err != nil {
		return
	}

	req.Body = io.NopCloser(bytes.NewBuffer(body))
	return
}

func read(reader io.Reader) (data []byte, err error) {

	if reader == nil {
		return
	}

	data, err = io.ReadAll(reader)
	if err != nil {
		err = errors.Wrapf(err, "failed to read from: %#v", reader)
		return
	}
	if len(data) == 0 {
		data = nil
	}

	return
}
