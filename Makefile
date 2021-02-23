DIST_DIR := dist
VERSION := $(shell date +%FT%T%z)-$(shell git log --pretty=format:'%h' -n 1)
LDFLAGS := '-X main.version=$(VERSION)'

all: clean server client clientproxy

clean:
	rm -rf $(DIST_DIR)

server:
	go build -ldflags $(LDFLAGS) -o $(DIST_DIR)/server cmd/server/main.go

client:
	go build -ldflags $(LDFLAGS) -o $(DIST_DIR)/client cmd/client/main.go

clientproxy:
	go build -ldflags $(LDFLAGS) -o $(DIST_DIR)/clientproxy cmd/clientproxy/main.go
