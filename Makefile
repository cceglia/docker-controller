# Docker Controller - validation and build commands.
# Every Go command needs the gtk3 build tag (only GTK3/WebKitGTK-4.1 is installed).
GOBIN := $(shell go env GOPATH)/bin
export PATH := $(PATH):$(GOBIN)
TAGS := gtk3

.PHONY: test vet typecheck bindings check build dev install uninstall

## test: Go tests (run `make bindings` or `make check` first only if you changed bound types)
test:
	go test -tags $(TAGS) ./...

vet:
	go vet -tags $(TAGS) ./...

## bindings: regenerate the Wails TypeScript bindings (frontend/bindings, git-ignored)
bindings:
	wails3 generate bindings -f '-tags $(TAGS)' -clean=true -ts -i

## typecheck: svelte-check on the frontend (needs the bindings)
typecheck: bindings
	cd frontend && npm install --silent && npm run check

## check: everything the orchestrator runs after a ticket: vet + tests + frontend typecheck/build
check: vet test typecheck
	cd frontend && npm run build

## build: production binary in bin/
build:
	wails3 build

## dev: run the app with hot reload
dev:
	wails3 dev -config ./build/config.yml

## install: build the production binary and install it with the desktop entry and icon under $(PREFIX) (default ~/.local)
PREFIX ?= $(HOME)/.local
DESTDIR ?=
APP_BIN := docker-controller
ICON_DIR := $(PREFIX)/share/icons/hicolor/1024x1024/apps
DESKTOP_DIR := $(PREFIX)/share/applications
BIN_INSTALL := $(DESTDIR)$(PREFIX)/bin/$(APP_BIN)
DESKTOP_INSTALL := $(DESTDIR)$(DESKTOP_DIR)/docker-controller.desktop
ICON_INSTALL := $(DESTDIR)$(ICON_DIR)/docker-controller.png

install: build
	install -Dm755 bin/dockercontroller $(BIN_INSTALL)
	install -d $(DESTDIR)$(DESKTOP_DIR)
	sed 's|@BINDIR@|$(PREFIX)/bin|' build/linux/docker-controller.desktop.in > $(DESKTOP_INSTALL)
	chmod 644 $(DESKTOP_INSTALL)
	install -Dm644 build/appicon.png $(ICON_INSTALL)
	-command -v update-desktop-database >/dev/null && update-desktop-database -q $(DESTDIR)$(DESKTOP_DIR)
	-command -v gtk-update-icon-cache >/dev/null && gtk-update-icon-cache -q -t -f $(DESTDIR)$(PREFIX)/share/icons/hicolor
	@echo "Installato in $(DESTDIR)$(PREFIX). Il database non viene toccato."

## uninstall: remove the files installed by `make install` (never the database in ~/.local/share/docker-controller)
uninstall:
	rm -f $(BIN_INSTALL) $(DESKTOP_INSTALL) $(ICON_INSTALL)
	-command -v update-desktop-database >/dev/null && update-desktop-database -q $(DESTDIR)$(DESKTOP_DIR)
	-command -v gtk-update-icon-cache >/dev/null && gtk-update-icon-cache -q -t -f $(DESTDIR)$(PREFIX)/share/icons/hicolor
	@echo "Disinstallato da $(DESTDIR)$(PREFIX). Il database e' stato conservato."
