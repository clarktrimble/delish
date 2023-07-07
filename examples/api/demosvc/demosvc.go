package demosvc

import (
	"net/http"

	"github.com/clarktrimble/delish"
)

func AddRoute(svr *delish.Server, rtr Router) (svc *demoSvc) {

	svc = &demoSvc{
		Server: svr,
	}

	rtr.Set("GET", "/brunch", svc.getBrunch)
	return
}

// unexported

type demoSvc struct {
	Server *delish.Server
}

func (svc *demoSvc) getBrunch(writer http.ResponseWriter, request *http.Request) {

	ctx := request.Context()
	//rp := svc.Server.NewResponder(writer)
	rp := &delish.Respond{
		Writer: writer,
		Logger: svc.Server.Logger,
	}

	rp.WriteObjects(ctx, map[string]any{"brunch": []string{"green eggs", "ham"}})
}
