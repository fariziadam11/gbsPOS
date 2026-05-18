.PHONY: test test-pos test-cms build build-pos build-cms run run-pos run-cms

build: build-pos build-cms

build-pos:
	cd gbs-pos-api && go build -o ../bin/gbs-pos-api ./cmd/server

build-cms:
	cd gbs-cms-api && go build -o ../bin/gbs-cms-api ./cmd/server

test: test-pos test-cms

test-pos:
	cd gbs-pos-api && go test ./...

test-cms:
	cd gbs-cms-api && go test ./...

run-pos:
	cd gbs-pos-api && go run ./cmd/server

run-cms:
	cd gbs-cms-api && go run ./cmd/server
