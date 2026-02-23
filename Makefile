# Makefile
.PHONY: build

install:
	go mod tidy
	$(MAKE) build/custom-gcl

# this compile the custom linter `custom-gcl` into the `build` only once
build/custom-gcl:
	go tool golangci-lint custom

lint: build/custom-gcl
	./build/custom-gcl run ./...

test:
	@go clean -testcache
	@go test -v -cover -coverprofile coverage.out -race ./...
	@go tool cover -func coverage.out

build:
	go build ./...
