VERSION := $(shell cat ../VERSION)

LDFLAGS         := -ldflags "-X main.version=$(VERSION)"
RELEASE_LDFLAGS := -ldflags "-s -w -X main.version=$(VERSION)"

.PHONY: build release test clean

build:
	mkdir -p bin
	CGO_ENABLED=0 go build $(LDFLAGS) -o bin/crafter .

release:
	mkdir -p bin
	CGO_ENABLED=0 GOOS=darwin  GOARCH=arm64 go build $(RELEASE_LDFLAGS) -o bin/crafter-darwin-arm64 .
	CGO_ENABLED=0 GOOS=darwin  GOARCH=amd64 go build $(RELEASE_LDFLAGS) -o bin/crafter-darwin-amd64 .
	CGO_ENABLED=0 GOOS=linux   GOARCH=amd64 go build $(RELEASE_LDFLAGS) -o bin/crafter-linux-amd64 .
	CGO_ENABLED=0 GOOS=linux   GOARCH=arm64 go build $(RELEASE_LDFLAGS) -o bin/crafter-linux-arm64 .

test:
	go test ./...

clean:
	rm -rf bin/
