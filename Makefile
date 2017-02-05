## simple makefile to log workflow
.PHONY: all test clean build install

GOFLAGS ?= $(GOFLAGS:)
dbtype=$(shell grep dbtype config.json | cut -d \" -f 4)
dbname=$(shell grep dbname config.json | cut -d \" -f 4)
dbtable=$(shell grep dbtable config.json | cut -d \" -f 4)

all: clean install build

build:
	gofmt -w pastebin.go
	go build $(GOFLAGS) ./...
ifeq ($(dbtype),sqlite3)
	cat database.sql | sed 's/pastebin/$(dbtable)/' | sqlite3 $(dbname)
endif

install:
	go get github.com/dchest/uniuri
	go get github.com/ewhal/pygments
	go get github.com/mattn/go-sqlite3
	go get github.com/gorilla/mux
	go get github.com/go-sql-driver/mysql
	go get github.com/lib/pq
	go get golang.org/x/crypto/bcrypt
	go get github.com/gorilla/securecookie

test: install
	go install $(GOFLAGS) ./...

bench: install
	go test -run=NONE -bench=. $(GOFLAGS) ./...

clean:
	go clean $(GOFLAGS) -i ./...
	rm -rf ./build
	rm -rf pastebin.db
