BINARY := infracost
ENTRYPOINT := cmd/infracost/main.go

.PHONY: deps run build windows linux darwin build_all clean test fmt lint

deps:
	go mod download

run:
	go run $(ENTRYPOINT) $(ARGS)

build:
	go build -i -v -o build/$(BINARY) $(ENTRYPOINT)

windows:
	env GOOS=windows GOARCH=amd64 go build -i -v -o build/$(BINARY)-windows-amd64 $(ENTRYPOINT)

linux:
	env GOOS=linux GOARCH=amd64 go build -i -v -o build/$(BINARY)-linux-amd64 $(ENTRYPOINT)

darwin:
	env GOOS=darwin GOARCH=amd64 go build -i -v -o build/$(BINARY)-darwin-amd64 $(ENTRYPOINT)

build_all: build windows linux darwin

release: build_all
	cd build; tar -czf $(BINARY)-windows-amd64.tar.gz $(BINARY)-windows-amd64
	cd build; tar -czf $(BINARY)-linux-amd64.tar.gz $(BINARY)-linux-amd64
	cd build; tar -czf $(BINARY)-darwin-amd64.tar.gz $(BINARY)-darwin-amd64

clean:
	go clean
	rm -rf build/$(BINARY)*
	rm -rf release/$(BINARY)*

test:
	go test ./... $(or $(ARGS), -v -cover)

fmt:
	go fmt ./...

lint:
	golangci-lint run
