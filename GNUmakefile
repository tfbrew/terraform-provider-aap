# SPECIAL: Hardcoded provider prefix required in this file
REPO_BUILD_TAG_VAL=repoAAP
GOFLAGS=-tags=$(REPO_BUILD_TAG_VAL)
export GOFLAGS
export GOEXPERIMENT=jsonv2

default: fmt lint install generate

build:
	go build -v ./...

install: build
	go install -v ./...

lint:
	golangci-lint run --build-tags=$(REPO_BUILD_TAG_VAL)

generate:
	go run generate-examples/main.go
	(cd tools && go generate ./...)

fmt:
	gofmt -s -w -e .

test:
	go test -v -cover -timeout=120s -parallel=10 ./internal/provider

testacc:
	TF_ACC=1 go test -v -cover ./internal/provider

.PHONY: fmt lint test testacc build install generate
