all: fmt deps build

deps:
	go get .

fmt:
	gofmt -w .

build: fmt
	go build -o bin/`basename ${PWD}`

clean:
	-rm -rf bin pkg

install:
	./bin/sysfact -v
	cp ./bin/sysfact /usr/bin/sysfact
	chmod +x /usr/bin/sysfact
	test -d /var/lib/sysfact || mkdir -p /var/lib/sysfact
	rsync -r --delete ./shell.d /var/lib/sysfact/

bsd:
	@mkdir -p pkg/usr/local/bin
	@mkdir -p pkg/usr/local/lib/sysfact
	@cp bin/sysfact pkg/usr/local/bin/sysfact
	@chmod +x pkg/usr/local/bin/sysfact
	@rsync -r --delete ./shell.d pkg/usr/local/lib/sysfact/
	@tar czf sysfact-freebsd.tar.gz -C pkg .
