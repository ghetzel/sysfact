.PHONY: build ui

PKGS        := $(shell go list ./... 2> /dev/null | grep -v '/vendor')
LOCALS      := $(shell find . -type f -name '*.go' -not -path "./vendor*/*")

.EXPORT_ALL_VARIABLES:
GO111MODULE  = on

all: fmt deps build

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
	go build -o bin/sysfact ./cmd/sysfact
	which sysfact && cp -v bin/sysfact `which sysfact` || true

copy-to-and-run:
	scp bin/sysfact $(IP):
	ssh $(IP) 'chmod +x sysfact && sysfact'