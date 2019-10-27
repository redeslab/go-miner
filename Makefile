SHELL=PATH='$(PATH)' /bin/sh

GOBUILD=CGO_ENABLED=0 go build -ldflags '-w -s'

PLATFORM := $(shell uname -o)

NAME := HOP.exe
OS := windows

ifeq ($(PLATFORM), Msys)
    INCLUDE := ${shell echo "$(GOPATH)"|sed -e 's/\\/\//g'}
else ifeq ($(PLATFORM), Cygwin)
    INCLUDE := ${shell echo "$(GOPATH)"|sed -e 's/\\/\//g'}
else
	INCLUDE := $(GOPATH)
	NAME=HOP
	OS=linux
endif

# enable second expansion
.SECONDEXPANSION:

.PHONY: all
.PHONY: test

BINDIR=$(INCLUDE)/bin

all: build

build:
	GOOS=$(OS) GOARCH=amd64 $(GOBUILD) -o $(BINDIR)/$(NAME)

test:
	GOOS=darwin go build -ldflags '-w -s' -o $(BINDIR)/$(NAME)

clean:
	rm $(BINDIR)/$(NAME)
