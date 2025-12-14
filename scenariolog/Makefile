.PHONY: help build test install uninstall clean

GO ?= go
CMD_PKG ?= ./cmd/scenariolog

# Enable FTS5 by default (requires CGO and a C toolchain).
BUILD_TAGS ?= sqlite_fts5
CGO_ENABLED ?= 1
GOFLAGS ?=

PREFIX ?= $(HOME)/.local
BINDIR ?= $(PREFIX)/bin
BIN ?= scenariolog

BUILD_DIR ?= build
BUILD_OUT ?= $(BUILD_DIR)/$(BIN)
INSTALL_PATH ?= $(BINDIR)/$(BIN)

help:
	@echo "Targets:"
	@echo "  build      Build $(BUILD_OUT) (BUILD_TAGS='$(BUILD_TAGS)')"
	@echo "  test       Run unit tests"
	@echo "  install    Install to $(INSTALL_PATH)"
	@echo "  uninstall  Remove $(INSTALL_PATH)"
	@echo "  clean      Remove build artifacts"
	@echo ""
	@echo "Variables (override with e.g. make BUILD_TAGS= ...):"
	@echo "  BUILD_TAGS   (default: sqlite_fts5; set empty to disable)"
	@echo "  CGO_ENABLED  (default: 1)"
	@echo "  GOFLAGS      (default: empty)"
	@echo "  PREFIX       (default: $$HOME/.local)"

build:
	@mkdir -p "$(BUILD_DIR)"
	CGO_ENABLED="$(CGO_ENABLED)" $(GO) build $(GOFLAGS) -tags "$(BUILD_TAGS)" -o "$(BUILD_OUT)" "$(CMD_PKG)"

test:
	$(GO) test ./...

install: build
	@mkdir -p "$(BINDIR)"
	install -m 0755 "$(BUILD_OUT)" "$(INSTALL_PATH)"
	@echo "Installed: $(INSTALL_PATH)"

uninstall:
	rm -f "$(INSTALL_PATH)"
	@echo "Removed: $(INSTALL_PATH)"

clean:
	rm -rf "$(BUILD_DIR)"


