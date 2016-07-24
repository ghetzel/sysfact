all: vendor fmt build

update:
	test -d vendor && rm -rf vendor || exit 0
	glide up --strip-vcs --update-vendored

vendor:
	go list github.com/Masterminds/glide
	glide install --strip-vcs --update-vendored

fmt:
	gofmt -w .

build: fmt
	go build -o bin/`basename ${PWD}`

clean:
	rm -rf vendor bin pkg

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
