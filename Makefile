BINARY := infracost
PKG := github.com/infracost/infracost/cmd/infracost
VERSION := $(shell scripts/get_version.sh HEAD $(NO_DIRTY))
LD_FLAGS := -ldflags="-X 'github.com/infracost/infracost/internal/version.Version=$(VERSION)'"
BUILD_FLAGS := $(LD_FLAGS) -v

DEV_ENV := dev
ifdef INFRACOST_ENV
	DEV_ENV := $(INFRACOST_ENV)
endif

.PHONY: deps run build windows linux darwin build_all install release clean test fmt lint

deps:
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	go mod download

run:
	env INFRACOST_ENV=$(DEV_ENV) go run $(LD_FLAGS) $(PKG) $(ARGS)

jsonschema:
	go run ./cmd/jsonschema/main.go --out-file ./schema/infracost.schema.json

build:
	CGO_ENABLED=0 go build $(BUILD_FLAGS) -o build/$(BINARY) $(PKG)

windows:
	env GOOS=windows GOARCH=amd64 CGO_ENABLED=0 go build $(BUILD_FLAGS) -o build/$(BINARY).exe $(PKG)
	env GOOS=windows GOARCH=arm64 CGO_ENABLED=0 go build $(BUILD_FLAGS) -o build/$(BINARY)-arm64.exe $(PKG)

linux:
	env GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build $(BUILD_FLAGS) -o build/$(BINARY)-linux-amd64 $(PKG)
	env GOOS=linux GOARCH=arm64 CGO_ENABLED=0 go build $(BUILD_FLAGS) -o build/$(BINARY)-linux-arm64 $(PKG)

darwin:
	env GOOS=darwin GOARCH=amd64 CGO_ENABLED=0 go build $(BUILD_FLAGS) -o build/$(BINARY)-darwin-amd64 $(PKG)
	env GOOS=darwin GOARCH=arm64 CGO_ENABLED=0 go build $(BUILD_FLAGS) -o build/$(BINARY)-darwin-arm64 $(PKG)

build_all: build windows linux darwin

install:
	CGO_ENABLED=0 go install $(BUILD_FLAGS) $(PKG)

release: build_all
	cd build; tar -czf $(BINARY)-windows-amd64.tar.gz $(BINARY).exe; shasum -a 256 $(BINARY)-windows-amd64.tar.gz > $(BINARY)-windows-amd64.tar.gz.sha256
	cd build; zip -r $(BINARY)-windows-amd64.zip $(BINARY).exe; shasum -a 256 $(BINARY)-windows-amd64.zip > $(BINARY)-windows-amd64.zip.sha256
	cd build; tar -czf $(BINARY)-windows-arm64.tar.gz $(BINARY)-arm64.exe; shasum -a 256 $(BINARY)-windows-arm64.tar.gz > $(BINARY)-windows-arm64.tar.gz.sha256
	cd build; mv $(BINARY)-arm64.exe $(BINARY).exe; zip -r $(BINARY)-windows-arm64.zip $(BINARY).exe; shasum -a 256 $(BINARY)-windows-arm64.zip > $(BINARY)-windows-arm64.zip.sha256
	cd build; tar -czf $(BINARY)-linux-amd64.tar.gz $(BINARY)-linux-amd64; shasum -a 256 $(BINARY)-linux-amd64.tar.gz > $(BINARY)-linux-amd64.tar.gz.sha256
	cd build; tar -czf $(BINARY)-linux-arm64.tar.gz $(BINARY)-linux-arm64; shasum -a 256 $(BINARY)-linux-arm64.tar.gz > $(BINARY)-linux-arm64.tar.gz.sha256
	cd build; tar -czf $(BINARY)-darwin-amd64.tar.gz $(BINARY)-darwin-amd64; shasum -a 256 $(BINARY)-darwin-amd64.tar.gz > $(BINARY)-darwin-amd64.tar.gz.sha256
	cd build; tar -czf $(BINARY)-darwin-arm64.tar.gz $(BINARY)-darwin-arm64; shasum -a 256 $(BINARY)-darwin-arm64.tar.gz > $(BINARY)-darwin-arm64.tar.gz.sha256

clean:
	go clean
	rm -rf build/$(BINARY)*

# Run only short unit tests
test:
	INFRACOST_LOG_LEVEL=warn go test -short $(LD_FLAGS) ./... $(or $(ARGS), -v -cover)

# Run all tests
test_all:
	INFRACOST_LOG_LEVEL=warn go test -timeout 30m $(LD_FLAGS) ./... $(or $(ARGS), -v -cover)

# Run unit tests and shared integration tests
test_shared_int:
	INFRACOST_LOG_LEVEL=warn go test -timeout 30m $(LD_FLAGS) \
		$(shell go list ./... | grep -v ./internal/providers/terraform/aws | grep -v ./internal/providers/terraform/google | grep -v ./internal/providers/terraform/azure) \
		$(or $(ARGS), -v -cover)

test_cmd:
	INFRACOST_LOG_LEVEL=warn go test -timeout 30m $(LD_FLAGS) ./cmd/infracost $(or $(ARGS), -v -cover)

test_update_cmd:
	INFRACOST_LOG_LEVEL=warn go test -timeout 30m $(LD_FLAGS) ./cmd/infracost $(or $(ARGS), -update -v -cover)

test_usage:
	INFRACOST_LOG_LEVEL=warn go test -timeout 30m $(LD_FLAGS) ./internal/usage $(or $(ARGS), -v -cover)

test_update_usage:
	INFRACOST_LOG_LEVEL=warn go test -timeout 30m $(LD_FLAGS) ./internal/usage $(or $(ARGS), -update -v -cover)

# Run AWS resource tests
test_aws:
	INFRACOST_LOG_LEVEL=warn go test -timeout 30m $(LD_FLAGS) ./internal/providers/terraform/aws $(or $(ARGS), -v -cover)

# Run Google resource tests
test_google:
	INFRACOST_LOG_LEVEL=warn go test -timeout 30m $(LD_FLAGS) ./internal/providers/terraform/google $(or $(ARGS), -v -cover)

# Run Azure resource tests
test_azure:
	INFRACOST_LOG_LEVEL=warn go test -timeout 30m $(LD_FLAGS) ./internal/providers/terraform/azure $(or $(ARGS), -v -cover)

# Run AzureRM tests
test_azurerm:
	INFRACOST_LOG_LEVEL=warn go test -timeout 30m $(LD_FLAGS) ./internal/providers/azurerm $(or $(ARGS), -v -cover)

# Update AWS golden files tests
test_update:
	INFRACOST_LOG_LEVEL=warn go test -timeout 30m $(LD_FLAGS) ./internal/providers/... $(or $(ARGS), -update -v -cover)

# Update golden files tests
test_update_aws:
	INFRACOST_LOG_LEVEL=warn go test -timeout 30m $(LD_FLAGS) ./internal/providers/terraform/aws $(or $(ARGS), -update -v -cover)

# Update Google golden files tests
test_update_google:
	INFRACOST_LOG_LEVEL=warn go test -timeout 30m $(LD_FLAGS) ./internal/providers/terraform/google $(or $(ARGS), -update -v -cover)

# Update Azure golden files tests
test_update_azure:
	INFRACOST_LOG_LEVEL=warn go test -timeout 30m $(LD_FLAGS) ./internal/providers/terraform/azure $(or $(ARGS), -update -v -cover)

# Update AzureRM golden files tests
test_update_azurerm:
	INFRACOST_LOG_LEVEL=warn go test -timeout 30m $(LD_FLAGS) ./internal/providers/azurerm $(or $(ARGS), -update -v -cover)

fmt:
	go fmt ./...
	find . -name '*.tf' -exec terraform fmt {} \;

lint:
	golangci-lint run
