MAKEFLAGS += --no-builtin-rules
.SUFFIXES:

.PHONY: build
build: dist/goclif

.PHONY: test
test: lint build
	go test

.PHONY: lint
lint: $(GOMETALINTER)
	gometalinter --exclude bindata.go -D gosec errcheck

PLATFORMS := windows linux darwin
os = $(word 1, $@)

.PHONY: $(PLATFORMS)
$(PLATFORMS):
	mkdir -p dist
	GOOS=$(os) GOARCH=amd64 go build -o dist/goclif-$(VERSION)-$(os)-amd64

.PHONY: release
release: windows linux darwin

worker.js: worker.ts
	tsc

bindata.go: worker.js
	go-bindata worker.js

dist/goclif: *.go
	mkdir -p dist
	go build -o dist/goclif

.PHONY: run
run: build
	# DEBUG=1 ./dist/goclif $(ARGS)
	./dist/goclif $(ARGS)
