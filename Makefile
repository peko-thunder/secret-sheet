BINARY  := sheet-env
MODULE  := github.com/peko-thunder/sheet-password
VERSION := $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
LDFLAGS := -ldflags "-s -w -X main.version=$(VERSION)"

# ── Local build ──────────────────────────────────────────────────────────────

.PHONY: build
build:
	go build $(LDFLAGS) -o $(BINARY) .

.PHONY: install
install:
	go install $(LDFLAGS) .

# ── Cross-compilation (outputs to dist/) ─────────────────────────────────────

DIST := dist

.PHONY: build-all
build-all: clean-dist
	GOOS=linux   GOARCH=amd64   go build $(LDFLAGS) -o $(DIST)/$(BINARY)-linux-amd64     .
	GOOS=linux   GOARCH=arm64   go build $(LDFLAGS) -o $(DIST)/$(BINARY)-linux-arm64     .
	GOOS=darwin  GOARCH=amd64   go build $(LDFLAGS) -o $(DIST)/$(BINARY)-darwin-amd64    .
	GOOS=darwin  GOARCH=arm64   go build $(LDFLAGS) -o $(DIST)/$(BINARY)-darwin-arm64    .
	GOOS=windows GOARCH=amd64   go build $(LDFLAGS) -o $(DIST)/$(BINARY)-windows-amd64.exe .
	@echo ""
	@echo "Built binaries:"
	@ls -lh $(DIST)/

# ── Release archives (tar.gz / zip) ─────────────────────────────────────────

.PHONY: release
release: build-all
	cd $(DIST) && \
	  tar czf $(BINARY)-$(VERSION)-linux-amd64.tar.gz    $(BINARY)-linux-amd64   && \
	  tar czf $(BINARY)-$(VERSION)-linux-arm64.tar.gz    $(BINARY)-linux-arm64   && \
	  tar czf $(BINARY)-$(VERSION)-darwin-amd64.tar.gz   $(BINARY)-darwin-amd64  && \
	  tar czf $(BINARY)-$(VERSION)-darwin-arm64.tar.gz   $(BINARY)-darwin-arm64  && \
	  zip     $(BINARY)-$(VERSION)-windows-amd64.zip     $(BINARY)-windows-amd64.exe
	@echo ""
	@echo "Release archives:"
	@ls -lh $(DIST)/*.tar.gz $(DIST)/*.zip

# ── Utilities ─────────────────────────────────────────────────────────────────

.PHONY: clean
clean:
	rm -f $(BINARY)
	rm -rf $(DIST)

.PHONY: clean-dist
clean-dist:
	mkdir -p $(DIST)
	rm -f $(DIST)/$(BINARY)-*

.PHONY: test
test:
	go test ./...

.PHONY: vet
vet:
	go vet ./...
