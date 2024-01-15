
# Delish

![4byte_burger_onebit](https://github.com/clarktrimble/delish/assets/5055161/cdd74e04-dde1-45b7-931b-13396d53f7b1)

Coordinated startup and shutdown for Golang http server and associated services.

## Why?

Why wrap the stdlib http server?

 - log startup and shutdown behaviors
 - coordinate startup and shutdown between http server and other goroutines
 - log requests and responses

## More Features!

 - bring your own router
 - redact selected headers from logging
 - optionally skip logging of request and response bodies
 - response helper

## Logging Interface

```go
type Logger interface {
	Info(ctx context.Context, msg string, kv ...any)
	Error(ctx context.Context, msg string, err error, kv ...any)
	WithFields(ctx context.Context, kv ...any) context.Context
}
```

The logging interface is meant to support structured, contextual logging.
Through it `delish` logs startup/shutdown, handles errors, and optionally request and response.
The example api includes `minlog`, aiming for a modicum of readability in support of development.
See https://github.com/clarktrimble/sabot for json output, truncation and more.

## Example

```go
  // setup logger and initialize graceful

  lgr := &minlog.MinLog{}
  ctx := lgr.WithFields(context.Background(),
    "app_id", "api_demo",
    "run_id", hondo.Rand(7),
  )

  ctx = graceful.Initialize(ctx, &wg, lgr, "config", cfg)

  // setup router

  rtr := minroute.New(ctx, lgr)
  rtr.HandleFunc("GET /config", delish.ObjHandler("config", cfg, lgr))

  // start demo service

  svc := cfg.Service.New(rtr, lgr)
  svc.Start(ctx, &wg)

  // start api server and wait for interrupt

  svr := cfg.Server.NewWithLog(ctx, rtr, lgr)
  svr.Start(ctx, &wg)
  graceful.Wait(ctx)

  // delicious!
```

## Test and Build

```bash
proj/delish$ make
go generate ./...
golangci-lint run ./...
go test -count 1 github.com/clarktrimble/delish github.com/clarktrimble/delish/buffered github.com/clarktrimble/delish/graceful github.com/clarktrimble/delish/mid github.com/clarktrimble/delish/minroute github.com/clarktrimble/delish/respond
ok      github.com/clarktrimble/delish  0.039s
ok      github.com/clarktrimble/delish/buffered 0.008s
ok      github.com/clarktrimble/delish/graceful 0.110s
ok      github.com/clarktrimble/delish/mid      0.007s
ok      github.com/clarktrimble/delish/minroute 0.005s
ok      github.com/clarktrimble/delish/respond  0.005s
:: Building api
CGO_ENABLED=0 go build -ldflags '-X main.version=main.24.e73714b' -o bin/api examples/api/main.go
:: Done
```

## And Run!

```
proj/delish$ bin/api
14:53:52.2693 > starting up | app_id: api_demo  run_id: aAiz8lV
                config: {"version":"","server":{"host":"","port":8088,"timeout":10000000000},"service":{"interval":5000000000}}
14:53:52.2694 > worker starting | app_id: api_demo  run_id: aAiz8lV  worker_id: pcGzjPx
                name: service
14:53:52.2694 > starting http service | app_id: api_demo  run_id: aAiz8lV
14:53:52.2695 > listening | app_id: api_demo  run_id: aAiz8lV
                address: :8088
14:53:57.2742 > worker gonna work | app_id: api_demo  run_id: aAiz8lV  worker_id: pcGzjPx
14:54:17.2727 > worker gonna work | app_id: api_demo  run_id: aAiz8lV  worker_id: pcGzjPx
14:54:18.7671 > received request | app_id: api_demo  request_id: x3XAlkx  run_id: aAiz8lV
                body:
                headers: {"Accept":["*/*"],"User-Agent":["curl/7.88.1"]}
                method: GET
                path: /report
                query: {}
                remote_ip: 127.0.0.1
                remote_port: 45226
14:54:18.7672 > sending response | app_id: api_demo  request_id: x3XAlkx  run_id: aAiz8lV
                body: {"worked":4}
                elapsed: 11.498µs
                headers: {"Content-Type":["application/json"]}
                status: 200
14:54:22.2725 > worker gonna work | app_id: api_demo  run_id: aAiz8lV  worker_id: pcGzjPx
14:54:27.2721 > worker gonna work | app_id: api_demo  run_id: aAiz8lV  worker_id: pcGzjPx
14:55:07.2729 * worker canna work | app_id: api_demo  run_id: aAiz8lV  worker_id: pcGzjPx
                error: oops
14:55:12.2742 > worker gonna work | app_id: api_demo  run_id: aAiz8lV  worker_id: pcGzjPx
14:55:17.2729 > worker gonna work | app_id: api_demo  run_id: aAiz8lV  worker_id: pcGzjPx
!!interrupt!!
14:55:18.1824 > shutting down | app_id: api_demo  run_id: aAiz8lV
14:55:18.1825 > shutting down http service | app_id: api_demo  run_id: aAiz8lV
14:55:18.1827 > http service stopped | app_id: api_demo  run_id: aAiz8lV
14:55:18.9398 > worker shutting down | app_id: api_demo  run_id: aAiz8lV  worker_id: pcGzjPx
14:55:21.4416 > worker stopped | app_id: api_demo  run_id: aAiz8lV  worker_id: pcGzjPx
14:55:21.4417 > stopped | app_id: api_demo  run_id: aAiz8lV
```

## Request Logging Options

```go
mid.RedactHeaders = map[string]bool{"X-Authorization-Token": true}
mid.SkipBody = true
```

Sets middleware to redact a header and to skip body logging.

