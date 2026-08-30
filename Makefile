BIN         ?= sysfact

all: deps fmt test build docs

fmt:
	go fmt ./...
	go generate -x ./...
	-go mod tidy

deps:
	go get ./...
	go vet -printf=false ./...

test: fmt deps
	go test -printf=false ./...
	test -x ./bin/$(BIN) && ./bin/$(BIN) apply -s test/src/ -d test/dest/ || true

build: fmt
	go build -o bin/$(BIN) ./cmd/sysfact
	which sysfact 2> /dev/null && cp -v bin/sysfact `which sysfact` || true

binaries: fmt deps
	GOOS=linux BIN=sysfact make build
	GOOS=freebsd BIN=sysfact.freebsd make build
	GOOS=darwin BIN=sysfact.darwin make build

docs:
	owndoc render --property rootpath=/sysfact/

copy-to-and-run:
	scp bin/$(BIN) $(IP):sysfact
	ssh $(IP) 'chmod +x sysfact && sysfact -L debug'

.PHONY: deps fmt build docs binaries copy-to-and-run
.EXPORT_ALL_VARIABLES: