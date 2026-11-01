# Docker Controller - validation, build and packaging commands.
# The build is Qt 6 via miqt with cgo: Go, a C++ compiler and the Qt libraries.
GOBIN := $(shell go env GOPATH)/bin
export PATH := $(PATH):$(GOBIN)

APP_ID := io.github.cceglia.DockerController
APP_BIN := docker-controller
ICON_SIZES := 256x256 512x512

PREFIX ?= $(HOME)/.local
DESTDIR ?=

BIN_INSTALL := $(DESTDIR)$(PREFIX)/bin/$(APP_BIN)
DESKTOP_DIR := $(PREFIX)/share/applications
ICON_ROOT := $(PREFIX)/share/icons/hicolor
DESKTOP_INSTALL := $(DESTDIR)$(DESKTOP_DIR)/$(APP_ID).desktop
ICON_INSTALL := $(foreach s,$(ICON_SIZES),$(DESTDIR)$(ICON_ROOT)/$(s)/apps/$(APP_ID).png)

# Names the previous version installed. `make uninstall` removes them too, so an
# upgrade leaves no second entry and no second icon in the menu.
LEGACY_DESKTOP := $(DESTDIR)$(DESKTOP_DIR)/docker-controller.desktop
LEGACY_ICON := $(foreach s,$(ICON_SIZES),$(DESTDIR)$(ICON_ROOT)/$(s)/apps/docker-controller.png)

# gtk-update-icon-cache only rebuilds the cache of the hicolor theme. It is a
# program of the desktop, not a library the build links, so the build needs no
# GTK: a machine without the tool installs and runs the app all the same.
# The prefix `-` and the `command -v` test make a missing tool no error.
ICON_CACHE := -command -v gtk-update-icon-cache >/dev/null && gtk-update-icon-cache -q -t -f $(DESTDIR)$(PREFIX)/share/icons/hicolor

.PHONY: test test-ui vet check build install uninstall

## test: Go tests
test:
	go test ./...

## test-ui: UI tests without a display: offscreen smoke test and design tokens
test-ui:
	QT_QPA_PLATFORM=offscreen go vet -tags ui ./internal/gui
	QT_QPA_PLATFORM=offscreen go test -tags ui ./internal/gui

vet:
	go vet ./...

## check: everything the orchestrator runs after a ticket: vet + tests + build
check: vet test build

## build: production binary in bin/
build:
	CGO_ENABLED=1 go build -o bin/dockercontroller .

## install: build the production binary and install it with the desktop entry and the icons under $(PREFIX) (default ~/.local)
install: build
	install -Dm755 bin/dockercontroller $(BIN_INSTALL)
	install -d $(DESTDIR)$(DESKTOP_DIR)
	sed 's|@BINDIR@|$(PREFIX)/bin|' build/linux/docker-controller.desktop.in > $(DESKTOP_INSTALL)
	chmod 644 $(DESKTOP_INSTALL)
	$(foreach s,$(ICON_SIZES),install -Dm644 build/linux/icons/$(s)/docker-controller.png $(DESTDIR)$(ICON_ROOT)/$(s)/apps/$(APP_ID).png;)
	-command -v update-desktop-database >/dev/null && update-desktop-database -q $(DESTDIR)$(DESKTOP_DIR)
	$(ICON_CACHE)
	@echo "Installato in $(DESTDIR)$(PREFIX). Il database non viene toccato."

## uninstall: remove the files installed by `make install`, including the names of the previous version (never the database in ~/.local/share/docker-controller)
uninstall:
	rm -f $(BIN_INSTALL) $(DESKTOP_INSTALL) $(LEGACY_DESKTOP) $(ICON_INSTALL) $(LEGACY_ICON)
	-command -v update-desktop-database >/dev/null && update-desktop-database -q $(DESTDIR)$(DESKTOP_DIR)
	$(ICON_CACHE)
	@echo "Disinstallato da $(DESTDIR)$(PREFIX). Il database e' stato conservato."
