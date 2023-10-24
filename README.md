
# Delish

![4byte_burger_onebit](https://github.com/clarktrimble/delish/assets/5055161/cdd74e04-dde1-45b7-931b-13396d53f7b1)

Coordinated startup and shutdown for Golang http server and associated services

## Why?

Why wrap the stdlib http server?

 - coordinate startup and shutdown between http server and other goroutines
 - log startup and shutdown behaviors

## More Features!

 - bring your own router
 - log requests and responses
 - redact selected headers from logging
 - response helper
 - demonstrate super handy service layer (in the spirit of Uncle Bob's Clean Architecture)

## Example

    // create logger and initialize graceful

    lgr := &minlog.MinLog{}
    ctx := lgr.WithFields(context.Background(), "run_id", hondo.Rand(7))

    ctx = graceful.Initialize(ctx, &wg, 6*cfg.Server.Timeout, lgr)

    // create router/handler, and server

    rtr := minroute.New(lgr)
    handler := mid.LogResponse(lgr, rtr)
    handler = mid.LogRequest(lgr, hondo.Rand, handler)
    handler = mid.ReplaceCtx(ctx, handler)

    server := cfg.Server.New(handler, lgr)

    // register route directly

    rtr.Set("GET", "/config", server.ObjHandler("config", cfg))

    // or via service layer

    demosvc.AddRoute(server, rtr)

    // delicious!

    server.Start(ctx, &wg)
    graceful.Wait(ctx)

## Logging Interface

    type Logger interface {
      Info(ctx context.Context, msg string, kv ...interface{})
      Error(ctx context.Context, msg string, err error, kv ...interface{})
      WithFields(ctx context.Context, kv ...interface{}) context.Context
    }

Yeah, everyone has one of these ..  The main idea with this one is to accept fields as key/value
pairs without fuss and carry context in the ctx, giving minimal invocation.
See https://github.com/clarktrimble/sabot for more.

## Test and Build

    delish % make
    go generate ./...
    CGO_ENABLED=0 golangci-lint run ./...
    CGO_ENABLED=0 go test -count 1 github.com/clarktrimble/delish ...
    ok  	github.com/clarktrimble/delish	0.425s
    ok  	github.com/clarktrimble/delish/buffered	0.257s
    ok  	github.com/clarktrimble/delish/graceful	0.901s
    ok  	github.com/clarktrimble/delish/mid	0.548s
    :: Building api
    CGO_ENABLED=0 go build -ldflags '-X main.version=main.3.e39fd95' -o bin/api examples/api/main.go
    :: Done

## And Run!

    delish % bin/api

    % curl -s localhost:8088/config | jq
    {
      "config": {
        "version": "main.3.e39fd95",
        "server": {
          "host": "",
          "port": 8088,
          "timeout": 10000000000
        }
      }
    }

### Corresponding (minimal) Logs

    msg > starting http service  ::run_id::v80QW6B
    msg > listening  ::run_id::v80QW6B
    kvs > ::address:::8088

    msg > received request  ::run_id::v80QW6B ::request_id::qWwVxnP
    kvs > ::method::GET::path::/config::query::map[]::body::::remote_ip::127.0.0.1::remote_port:...

    msg > sending response  ::run_id::v80QW6B ::request_id::qWwVxnP
    kvs > ::status::200::headers::map[Content-Type:[application/json]]::body::{"config":{"version"::...

    ^C
    msg > shutting down ..  ::run_id::v80QW6B
    msg > shutting down http service ..  ::run_id::v80QW6B
    msg > http service stopped  ::run_id::v80QW6B
    msg > stopped  ::run_id::v80QW6B

Minimal in the sense that they're logged by `minlog` as seen in examples.
Typically we'll use something sending json or other structured output.

## Header Redaction

    RedactHeaders = map[string]bool{"X-Authorization-Token": true}

And the value for these will be logged as `--redacted--`.

## Golang (Anti) Idioms

I dig the Golang community, but I might be a touch rouge with:

  - multi-char variable names
  - named return parameters
  - BDD/DSL testing
  - liberal use of vertical space

All in the name of readability, which of course, tends towards the subjective.

## License

This is free and unencumbered software released into the public domain.

Anyone is free to copy, modify, publish, use, compile, sell, or
distribute this software, either in source code form or as a compiled
binary, for any purpose, commercial or non-commercial, and by any
means.

In jurisdictions that recognize copyright laws, the author or authors
of this software dedicate any and all copyright interest in the
software to the public domain. We make this dedication for the benefit
of the public at large and to the detriment of our heirs and
successors. We intend this dedication to be an overt act of
relinquishment in perpetuity of all present and future rights to this
software under copyright law.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND,
EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF
MERCHANTABILITY, FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT.
IN NO EVENT SHALL THE AUTHORS BE LIABLE FOR ANY CLAIM, DAMAGES OR
OTHER LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE,
ARISING FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR
OTHER DEALINGS IN THE SOFTWARE.

For more information, please refer to <http://unlicense.org/>

