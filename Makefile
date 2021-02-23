DIST_DIR := dist
VERSION := $(shell date +%FT%T%z)-$(shell git log --pretty=format:'%h' -n 1)
LDFLAGS := '-X main.version=$(VERSION)'

all: clean daemon

clean:
	rm -rf $(DIST_DIR)

daemon:
	go build -ldflags $(LDFLAGS) -o $(DIST_DIR)/daemon cmd/daemon/main.go
