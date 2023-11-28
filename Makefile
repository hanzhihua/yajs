#GITCOMMIT:=$(shell git describe --always)
BUILD_TAGS = $(shell git describe --tags)
BUILD_TIME = $(shell date)
GIT_COMMIT = $(shell git rev-parse --short HEAD)
GO_VERSION = $(shell go version)
OBJPREFIX := "main"

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
	CGO_ENABLED=$(CGO_ENABLED) $(SYSTEM) go build $(BUILDOPTS) -ldflags="-v -s -w -X '$(OBJPREFIX).BuildTags=$(BUILD_TAGS)' -X '$(OBJPREFIX).BuildTime=$(BUILD_TIME)' -X '$(OBJPREFIX).GitCommit=$(GIT_COMMIT)' -X '$(OBJPREFIX).GoVersion=$(GO_VERSION)' " -o $(BINARY) cmd/yajs.go cmd/pid.go


.PHONY: clean
clean:
	go clean
	rm -f yajs

.PHONY: dev
dev:clean
	CGO_ENABLED=$(CGO_ENABLED)  go build $(BUILDOPTS) -ldflags="-v -s -w  -X '$(OBJPREFIX).BuildTags=$(BUILD_TAGS)' -X '$(OBJPREFIX).BuildTime=$(BUILD_TIME)' -X '$(OBJPREFIX).GitCommit=$(GIT_COMMIT)' -X '$(OBJPREFIX).GoVersion=$(GO_VERSION)'" -o $(BINARY) cmd/yajs.go cmd/pid.go