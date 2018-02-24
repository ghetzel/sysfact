VERSION := `./bin/sysfact --version | cut -d' ' -f3`

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
	rsync -rv --delete ./shell.d /var/lib/sysfact/

prep-package:
	-rm -rf pkg
	mkdir -p pkg/usr/bin
	mkdir -p pkg/var/lib/sysfact
	cp ./bin/sysfact pkg/usr/bin/sysfact
	rsync -rv --delete ./shell.d pkg/var/lib/sysfact/
	-rm *.deb *.tar.gz

package: prep-package
	fpm -s dir -t deb -n sysfact -v $(VERSION) -C pkg usr var
	tar czvf sysfact-$(VERSION).tar.gz -C pkg ./

bsd:
	@mkdir -p pkg/usr/local/bin
	@mkdir -p pkg/usr/local/lib/sysfact
	GOOS=freebsd GOARCH=amd64 go build -o pkg/usr/local/bin/sysfact
	@chmod +x pkg/usr/local/bin/sysfact
	@rsync -r --delete ./shell.d pkg/usr/local/lib/sysfact/
	@tar czf sysfact-freebsd.tar.gz --owner=0 --group=0 -C pkg .
