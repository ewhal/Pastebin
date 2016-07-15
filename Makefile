## simple makefile to log workflow
.PHONY: all test clean build install

GOFLAGS ?= $(GOFLAGS:)

all: clean install build

build:
	go build $(GOFLAGS) ./...

install:
	go get github.com/dchest/uniuri
	go get github.com/ewhal/pygments
	go get github.com/go-sql-driver/mysql
	go get github.com/gorilla/mux
	go get github.com/ChannelMeter/iso8601duration

test: install
	go test $(GOFLAGS) ./...

bench: install
	go test -run=NONE -bench=. $(GOFLAGS) ./...

clean:
	go clean $(GOFLAGS) -i ./...
	rm -rf ./build 


