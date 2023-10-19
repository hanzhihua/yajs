GITCOMMIT:=$(shell git describe --always)
BINARY:=yajs
SYSTEM:=GOOS=linux  GOARCH=amd64
CHECKS:=check
BUILDOPTS:=-v
GOPATH?=go
CGO_ENABLED?=0


.PHONY: all
all: yajs

.PHONY: yajs
yajs:clean
	CGO_ENABLED=$(CGO_ENABLED) $(SYSTEM) go build $(BUILDOPTS) -ldflags="-s -w -X main.GitCommit=$(GITCOMMIT)" -o $(BINARY) cmd/yajs.go


.PHONY: clean
clean:
	go clean
	rm -f yajs

.PHONY: dev
dev:clean
	CGO_ENABLED=$(CGO_ENABLED)  go build $(BUILDOPTS) -ldflags="-s -w -X main.GitCommit=$(GITCOMMIT)" -o $(BINARY) cmd/yajs.go