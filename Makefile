.PHONY: build

all: build

VERSION=v0.1.14
GORELEASER_ARGS ?= --skip=sign --snapshot --clean
GORELEASER_TARGET ?= --single-target

docker-lint:
	docker run --rm -v $(shell pwd):/app -w /app golangci/golangci-lint:latest golangci-lint run -v

lint:
	golangci-lint run -v

lintmax:
	golangci-lint run -v --max-same-issues=100

gosec:
	go install github.com/securego/gosec/v2/cmd/gosec@latest
	gosec -exclude=G101,G304,G301,G306 -exclude-dir=.history ./...

govulncheck:
	go install golang.org/x/vuln/cmd/govulncheck@latest
	govulncheck ./...

test:
	go test ./...

build: ui-build
	go build -tags "sqlite_fts5,embed" ./cmd/docmgr

.PHONY: ui-build build-embed dev-backend

ui-build:
	GOWORK=off go run ./internal/web/generate_build.go

build-embed: ui-build
	go build -tags "sqlite_fts5,embed" ./cmd/docmgr

dev-backend:
	go run -tags sqlite_fts5 ./cmd/docmgr api serve --addr 127.0.0.1:3001 --root ttmp

goreleaser:
	goreleaser release $(GORELEASER_ARGS) $(GORELEASER_TARGET)

tag-major:
	git tag $(shell svu major)

tag-minor:
	git tag $(shell svu minor)

tag-patch:
	git tag $(shell svu patch)

release:
	git push origin --tags
	GOPROXY=proxy.golang.org go list -m github.com/go-go-golems/docmgr@$(shell svu current)

bump-glazed:
	go get github.com/go-go-golems/glazed@latest
	go get github.com/go-go-golems/clay@latest
	go mod tidy

DOCMGR_BINARY=$(shell which docmgr)
install: ui-build
	go build -tags "sqlite_fts5,embed" -o ./dist/docmgr ./cmd/docmgr && \
		cp ./dist/docmgr $(DOCMGR_BINARY)
