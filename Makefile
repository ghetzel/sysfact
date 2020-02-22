.PHONY: build ui

PKGS        := $(shell go list ./... 2> /dev/null | grep -v '/vendor')
LOCALS      := $(shell find . -type f -name '*.go' -not -path "./vendor*/*")
BIN         ?= sysfact

.EXPORT_ALL_VARIABLES:
GO111MODULE  = on

all: deps fmt build docs

fmt:
	gofmt -w $(LOCALS)
	go generate -x ./...
	-go mod tidy

deps:
	go get ./...
	go vet ./...

test: fmt deps
	go test $(PKGS)

build: fmt
	go build -o bin/$(BIN) ./cmd/sysfact
	which sysfact && cp -v bin/sysfact `which sysfact` || true

binaries: fmt deps
	GOOS=linux BIN=sysfact make build
	GOOS=freebsd BIN=sysfact.freebsd make build
	GOOS=darwin BIN=sysfact.darwin make build

docs:
	owndoc render --property rootpath=/sysfact/

copy-to-and-run:
	scp bin/$(BIN) $(IP):sysfact
	ssh $(IP) 'chmod +x sysfact && sysfact -L debug'