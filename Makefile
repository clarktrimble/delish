EXECS   := $(wildcard examples/*)
TARGETS := ${EXECS:examples/%=%}

TESTA   := ${shell go list ./... | grep -v /examples/ | grep -v /test/ | grep -v /mock}

BRANCH   := ${shell git branch --show-current}
REVCNT   := ${shell git rev-list --count $(BRANCH)}
REVHASH  := ${shell git log -1 --format="%h"}

LDFLAGS  := -X main.version=${BRANCH}.${REVCNT}.${REVHASH}

all: check build

check: gen lint test

cover:
	go test -coverprofile=cover.out ${TESTA} && \
	go tool cover -func=cover.out

gen:
	go generate ./...

lint:
	golangci-lint run ./...

test:
	go test -count 1 ${TESTA}

race:
	CGO_ENABLED=1 go test -count 1 -race ${TESTA}

build: ${TARGETS}
	@echo ":: Done"

${TARGETS}:
	@echo ":: Building $@"
	CGO_ENABLED=0 go build -ldflags '${LDFLAGS}' -o bin/$@ examples/$@/main.go

.PHONY:  test tag

# Todo: sync with canaon

tag:
	@echo "Current: $$(git describe --tags --abbrev=0)"
	@read -p "New tag: " tag && git tag $$tag && \
	  read -p "Push $$tag? [y/N] " confirm && \
	  if [ "$$confirm" = "y" ]; then git push origin $$tag; fi

