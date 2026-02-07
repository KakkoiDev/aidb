PREFIX ?= /usr/local
BINDIR ?= $(PREFIX)/bin

.PHONY: build install uninstall test clean

build:
	go build -o aidb ./cmd/aidb

install:
	install -d $(BINDIR)
	install -m 755 aidb $(BINDIR)/aidb

uninstall:
	rm -f $(BINDIR)/aidb

test:
	go test ./...

clean:
	rm -f aidb
