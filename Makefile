BINARY := infracost
ENTRYPOINT := cmd/infracost/main.go
GEN_DOCS_ENTRYPOINT := cmd/infracost-generate-docs/main.go
TERRAFORM_PROVIDER_INFRACOST_VERSION := latest
VERSION := $(shell scripts/get_version.sh HEAD)
LD_FLAGS := -ldflags="-X 'github.com/infracost/infracost/pkg/version.Version=$(VERSION)'"
BUILD_FLAGS := $(LD_FLAGS) -i -v -o
DOCS_TEMPLATES_PATH := docs/templates
DOCS_OUTPUT_PATH := docs/generated

ifndef $(GOOS)
	GOOS=$(shell go env GOOS)
endif

ifndef $(GOARCH)
	GOARCH=$(shell go env GOARCH)
endif

.PHONY: deps run build windows linux darwin build_all release install_provider clean test fmt lint docs

deps:
	go mod download

run:
	INFRACOST_ENV=dev go run $(LD_FLAGS) $(ENTRYPOINT) $(ARGS)

build:
	CGO_ENABLED=0 go build $(BUILD_FLAGS) build/$(BINARY) $(ENTRYPOINT)

windows:
	env GOOS=windows GOARCH=amd64 CGO_ENABLED=0 go build $(BUILD_FLAGS) build/$(BINARY)-windows-amd64 $(ENTRYPOINT)

linux:
	env GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build $(BUILD_FLAGS) build/$(BINARY)-linux-amd64 $(ENTRYPOINT)

darwin:
	env GOOS=darwin GOARCH=amd64 CGO_ENABLED=0 go build $(BUILD_FLAGS) build/$(BINARY)-darwin-amd64 $(ENTRYPOINT)

build_all: build windows linux darwin docs

release: build_all
	cd build; tar -czf $(BINARY)-windows-amd64.tar.gz $(BINARY)-windows-amd64
	cd build; tar -czf $(BINARY)-linux-amd64.tar.gz $(BINARY)-linux-amd64
	cd build; tar -czf $(BINARY)-darwin-amd64.tar.gz $(BINARY)-darwin-amd64
	cd docs/generated; tar -czvf docs.tar.gz *.md

install_provider:
	scripts/install_provider.sh $(TERRAFORM_PROVIDER_INFRACOST_VERSION)

clean:
	go clean
	rm -rf build/$(BINARY)*

test:
	go test -timeout 15m $(LD_FLAGS) ./... $(or $(ARGS), -v -cover)

fmt:
	go fmt ./...

lint:
	golangci-lint run

docs:
	go run $(GEN_DOCS_ENTRYPOINT) --input $(DOCS_TEMPLATES_PATH) --output $(DOCS_OUTPUT_PATH)
