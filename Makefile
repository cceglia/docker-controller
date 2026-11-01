# Docker Controller - validation and build commands.
# The build is GTK4 via gotk4.
GOBIN := $(shell go env GOPATH)/bin
export PATH := $(PATH):$(GOBIN)

.PHONY: test vet check build install uninstall

## test: Go tests
test:
	go test ./...

vet:
	go vet ./...

## check: everything the orchestrator runs after a ticket: vet + tests
check: vet test

## build: production binary in bin/
build:
	CGO_ENABLED=1 go build -o bin/dockercontroller .

## install: build the production binary and install it with the desktop entry and icon under $(PREFIX) (default ~/.local)
PREFIX ?= $(HOME)/.local
DESTDIR ?=
APP_BIN := docker-controller
ICON_SIZES := 256x256 512x512
ICON_ROOT := $(PREFIX)/share/icons/hicolor
DESKTOP_DIR := $(PREFIX)/share/applications
BIN_INSTALL := $(DESTDIR)$(PREFIX)/bin/$(APP_BIN)
DESKTOP_INSTALL := $(DESTDIR)$(DESKTOP_DIR)/docker-controller.desktop
ICON_INSTALL := $(foreach s,$(ICON_SIZES),$(DESTDIR)$(ICON_ROOT)/$(s)/apps/docker-controller.png)

install: build
	install -Dm755 bin/dockercontroller $(BIN_INSTALL)
	install -d $(DESTDIR)$(DESKTOP_DIR)
	sed 's|@BINDIR@|$(PREFIX)/bin|' build/linux/docker-controller.desktop.in > $(DESKTOP_INSTALL)
	chmod 644 $(DESKTOP_INSTALL)
	$(foreach s,$(ICON_SIZES),install -Dm644 build/linux/icons/$(s)/docker-controller.png $(DESTDIR)$(ICON_ROOT)/$(s)/apps/docker-controller.png;)
	-command -v update-desktop-database >/dev/null && update-desktop-database -q $(DESTDIR)$(DESKTOP_DIR)
	-command -v gtk-update-icon-cache >/dev/null && gtk-update-icon-cache -q -t -f $(DESTDIR)$(PREFIX)/share/icons/hicolor
	@echo "Installato in $(DESTDIR)$(PREFIX). Il database non viene toccato."

## uninstall: remove the files installed by `make install` (never the database in ~/.local/share/docker-controller)
uninstall:
	rm -f $(BIN_INSTALL) $(DESKTOP_INSTALL) $(ICON_INSTALL)
	-command -v update-desktop-database >/dev/null && update-desktop-database -q $(DESTDIR)$(DESKTOP_DIR)
	-command -v gtk-update-icon-cache >/dev/null && gtk-update-icon-cache -q -t -f $(DESTDIR)$(PREFIX)/share/icons/hicolor
	@echo "Disinstallato da $(DESTDIR)$(PREFIX). Il database e' stato conservato."
