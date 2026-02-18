VERSION ?= dev
LDFLAGS := -s -w -X github.com/KakkoiDev/aidb/cmd/aidb/cmd.version=$(VERSION)
BINDIR  ?= $(or $(GOBIN),/usr/local/bin)

.PHONY: build install uninstall test clean release

build:
	go build -ldflags '$(LDFLAGS)' -o aidb ./cmd/aidb

install:
	install -d $(BINDIR)
	install -m 755 aidb $(BINDIR)/aidb

uninstall:
	rm -f $(BINDIR)/aidb

test:
	go test ./...

clean:
	rm -f aidb

release:
	goreleaser release --snapshot --clean
