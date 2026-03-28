BINARY_NAME=arch_forge
VERSION?=dev
BUILD_DIR=./bin

.PHONY: build test lint run clean docker-build docker-push release-snapshot release-dry-run release

build:
	go build -ldflags="-X github.com/archforge/cli/internal/adapter/cli.Version=$(VERSION)" -o $(BUILD_DIR)/$(BINARY_NAME) ./cmd/archforge

test:
	go test ./... -v -race -coverprofile=coverage.out

lint:
	golangci-lint run ./...

run:
	go run ./cmd/archforge $(ARGS)

clean:
	rm -rf $(BUILD_DIR) coverage.out

docker-build:
	docker build --build-arg VERSION=$(VERSION) -t archforge/cli:$(VERSION) .

docker-push:
	docker push archforge/cli:$(VERSION)
	docker tag archforge/cli:$(VERSION) archforge/cli:latest
	docker push archforge/cli:latest

release-snapshot:
	goreleaser release --snapshot --clean

release-dry-run:
	goreleaser release --skip=publish --clean

release:
	goreleaser release --clean
