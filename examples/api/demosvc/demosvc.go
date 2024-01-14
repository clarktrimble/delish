package demosvc

import (
	"net/http"

	"github.com/clarktrimble/delish"
	"github.com/clarktrimble/delish/respond"
)

// Todo: example svc here is odd, make regular with New, replace Server with Logger, etc

// Router specifies an http router.
type Router interface {
	HandleFunc(pattern string, handler http.HandlerFunc)
}

func AddRoute(svr *delish.Server, rtr Router) (svc *demoSvc) {

	svc = &demoSvc{
		Server: svr,
	}

	rtr.HandleFunc("GET /brunch", svc.getBrunch)
	return
}

// unexported

type demoSvc struct {
	Server *delish.Server
}

func (svc *demoSvc) getBrunch(writer http.ResponseWriter, request *http.Request) {

	ctx := request.Context()

	rp := &respond.Respond{
		Writer: writer,
		Logger: svc.Server.Logger,
	}

	rp.WriteObjects(ctx, map[string]any{"brunch": []string{"green eggs", "ham"}})
}
