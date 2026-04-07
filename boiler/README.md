# boiler

Boilerplate HTTP routes for service observability and API documentation.

## Example

```go
package main

import (
	"context"
	_ "embed"
	"net/http"
	"os"
	"sync"

	"github.com/clarktrimble/delish"
	"github.com/clarktrimble/delish/boiler"
	"github.com/clarktrimble/delish/graceful"
	"github.com/clarktrimble/launch"
	"github.com/clarktrimble/sabot"
)

//go:generate apispec gen

//go:embed openapi.yaml
var apiSpec []byte

// Note: generating and embedding!
// touch openapi.yaml to avoid chicken and egg

var (
	version string
	release string
)

type config struct {
	Version string         `json:"version" ignored:"true"`
	Release string         `json:"release" ignored:"true"`
	Url     string         `json:"url" desc:"URL for API spec" default:"http://localhost:8080"`
	Logger  *sabot.Config  `json:"logger"`
	Server  *delish.Config `json:"server"`
}

func main() {
	var wg sync.WaitGroup

	cfg := &config{Version: version, Release: release}
	launch.Load(cfg, "myapp", "my excellent service")

	lgr := cfg.Logger.New(os.Stdout)
	ctx := lgr.WithFields(context.Background(), "app_id", "myapp")

	ctx = graceful.Initialize(ctx, &wg, lgr)
	rtr := http.NewServeMux()
	boiler.Register(ctx, rtr, cfg, apiSpec, lgr)

	// register additional routes on rtr ...

	server := cfg.Server.NewWithLog(ctx, rtr, lgr)
	server.Start(ctx, &wg)
	graceful.Wait(ctx)
}
```

`Register` adds the boilerplate routes to any router satisfying the `Router` interface (stdlib `*http.ServeMux` or similar).
It uses reflection to extract `Version`, `Release`, and `Url` fields from cfg, falling back gracefully when they're absent.
The docs page title is pulled from the spec's `info.title` field.

## Routes

| Route | Description |
|-------|-------------|
| `GET /config` | App config as JSON |
| `GET /monitor` | Health check |
| `GET /log` | Current log level |
| `POST /log/{level}` | Set log level |
| `GET /docs` | Interactive API docs |
| `GET /openapi.yaml` | OpenAPI spec |

## Spec Placeholders

- `${PUBLISHED_URL}` - substituted with `Url` from cfg
- `${RELEASE}` - substituted with `Release` or `Version` from cfg per below.

Stoplight prepends "v" to the release value, so a tag like `1.2.3` displays as `v1.2.3`.

When `Release` is empty, `Version` is used with an underscore prefix (e.g. `_main.42.abc1234`).
When both are empty, `_unreleased` is used.
