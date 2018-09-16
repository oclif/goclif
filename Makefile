MAKEFLAGS += --no-builtin-rules
.SUFFIXES:

.PHONY: test
test: lint
	go test

.PHONY: lint
lint: $(GOMETALINTER)
	gometalinter --exclude bindata.go -D errcheck

PLATFORMS := windows linux darwin
os = $(word 1, $@)

.PHONY: $(PLATFORMS)
$(PLATFORMS):
	mkdir -p dist
	GOOS=$(os) GOARCH=amd64 go build -o dist/goclif-$(VERSION)-$(os)-amd64

.PHONY: release
release: windows linux darwin

bindata.go: server.js
	go-bindata server.js

dist/goclif: *.go
	mkdir -p dist
	go build -o dist/goclif

.PHONY: build
build: dist/goclif

.PHONY: run
run: build
	DEBUG=1 ./dist/goclif $(ARGS)
