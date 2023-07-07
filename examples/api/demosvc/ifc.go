package demosvc

import (
	"net/http"
)

type Router interface {
	// for example, customize for router of choice
	Set(method, path string, handler http.HandlerFunc) (err error)
}
