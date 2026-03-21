# boiler

Boilerplate HTTP routes for service observability and API documentation.

## Example

```go
package main

import (
	"context"
	"sync"

	"github.com/clarktrimble/delish"
	"github.com/clarktrimble/delish/boiler"
	"github.com/clarktrimble/delish/graceful"
	"github.com/clarktrimble/launch"
	"github.com/clarktrimble/sabot"
)

var (
	version string
	release string
)

type config struct {
	Version string        `json:"version" ignored:"true"`
	Release string        `json:"release" ignored:"true"`
	Url     string        `json:"url" desc:"URL for API spec" default:"http://localhost:8080"`
	Logger  *sabot.Config `json:"logger"`
	Server  *delish.Config `json:"server"`
}

//go:embed openapi.yaml
var apiSpec []byte

func main() {
	var wg sync.WaitGroup

	cfg := &config{Version: version, Release: release}
	launch.Load(cfg, "myapp", "my excellent service")

	lgr := cfg.Logger.New(os.Stdout)
	ctx := lgr.WithFields(context.Background(), "app_id", "myapp")

	ctx = graceful.Initialize(ctx, &wg, lgr)
	spec := boiler.SubSpec(apiSpec, version, release, cfg.Url)
	rtr := boiler.NewRouter(ctx, cfg, "My Excellent Service", spec, lgr)

	// register additional routes on rtr ...

	server := cfg.Server.NewWithLog(ctx, rtr, lgr)
	server.Start(ctx, &wg)
	graceful.Wait(ctx)
}
```

## Routes

| Route | Description |
|-------|-------------|
| `GET /config` | App config as JSON |
| `GET /monitor` | Health check |
| `GET /log` | Current log level |
| `POST /log/{level}` | Set log level |
| `GET /docs` | Interactive API docs |
| `GET /openapi.yaml` | OpenAPI spec |

## SubSpec Placeholders

- `${RELEASE}` - substituted with release (or "_untagged" / "_unreleased")
- `${PUBLISHED_URL}` - substituted with provided URL
