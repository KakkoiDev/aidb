VERSION ?= dev
LDFLAGS := -s -w -X github.com/KakkoiDev/aidb/cmd/aidb/cmd.version=$(VERSION)
BINDIR  ?= $(or $(GOBIN),/usr/local/bin)

.PHONY: build install uninstall test clean release publish

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

publish:
	@test -n "$(V)" || (echo "Usage: make publish V=fix|feature|major|x.y.z" && exit 1)
	@LATEST=$$(git describe --tags --abbrev=0 2>/dev/null | sed 's/^v//') || LATEST="0.0.0"; \
	IFS='.' read -r MAJ MIN PAT <<< "$$LATEST"; \
	case "$(V)" in \
		fix)     PAT=$$((PAT + 1)) ;; \
		feature) MIN=$$((MIN + 1)); PAT=0 ;; \
		major)   MAJ=$$((MAJ + 1)); MIN=0; PAT=0 ;; \
		*)       IFS='.' read -r MAJ MIN PAT <<< "$(V)" ;; \
	esac; \
	TAG="v$$MAJ.$$MIN.$$PAT"; \
	echo "Tagging $$TAG and pushing..."; \
	git tag "$$TAG" && git push origin "$$TAG"
