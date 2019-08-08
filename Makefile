SRC := $(shell find . -type f -name '*.go'; echo go.mod)
BIND :=$(shell find bind -type f)

.PHONY: build
build: dist/app.wasm dist/index.html dist/wasm_exec.js

.PHONY: serve
serve: build
	cd dist && go run ../cmd/http

dist/index.html: assets/index.html | dist
	cp "$<" "$@"

dist/wasm_exec.js: assets/wasm_exec.js assets/wasm_exec_init.js | dist
	cp "$<" "$@"
	cat assets/wasm_exec_init.js >> "$@"

dist/app.wasm: $(SRC) bound/bound.go | dist
	GOARCH=wasm GOOS=js go build -o $@ ./cmd/app

bound/bound.go: $(BIND)
	go run ./cmd/bindata

dist:
	@mkdir dist 2>/dev/null || true

.PHONY: reset
reset:
	-rm -rf dist

.PHONY: publish
publish: build
	s3cmd sync -P dist/ s3://clipee.be/wasmcarver/
	s3cmd modify -m application/wasm s3://clipee.be/wasmcarver/app.wasm
