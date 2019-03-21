export GO111MODULE=on

SOURCE = $(wildcard *.go) $(wildcard */*.go) $(wildcard */*/*.go)
TARGET = ./target
BINARY = $(TARGET)/tinycloud

.PHONY: all clean build

all: clean build

clean:
	-rm -r $(TARGET)

build: $(BINARY)

$(BINARY): $(SOURCE)
	mkdir -p $(TARGET)
	go build \
		-o $(BINARY) \
		-ldflags '-X "main.Version=$(shell git log -1 --pretty=format:"%h (%ai)")"' \
		./cmd/tinycloud
