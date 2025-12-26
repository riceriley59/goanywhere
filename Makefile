# Global Variables


GIT_SHA 		:= $(shell git rev-parse --short HEAD)
GIT_TAG_LAST 	:= $(shell git tag --list 'operator*' --sort=-v:refname | head -n 1 | cut -d/ -f2)

## GO Flags
GO_LDFLAGS  := -ldflags "-X github.com/riceriley59/goanywhere/internal/version.GIT_SHA=$(GIT_SHA) \
	-X github.com/riceriley59/goanywhere/internal/version.VERSION=$(VERSION)"
GOFLAGS 	:= -mod=mod


# Default target, clean,  and help


.PHONY: all clean help
all: build

clean:
	rm -rf bin/

help:
	printf "hello"


# Build Targets


.PHONY: build build-goanywhere
build: build-goanywhere

build-goanywhere:
	go build $(GOFLAGS) $(GO_LDFLAGS) -o bin/goanywhere cmd/goanywhere/main.go
