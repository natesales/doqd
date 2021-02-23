DIST_DIR := dist
VERSION := $(shell date +%FT%T%z)-$(shell git log --pretty=format:'%h' -n 1)
LDFLAGS := '-X main.version=$(VERSION)'

all: clean daemon client

clean:
	rm -rf $(DIST_DIR)

daemon:
	go build -ldflags $(LDFLAGS) -o $(DIST_DIR)/daemon cmd/daemon/main.go

client:
	go build -ldflags $(LDFLAGS) -o $(DIST_DIR)/client cmd/client/main.go
