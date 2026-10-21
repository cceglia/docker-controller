# Docker Controller - validation and build commands.
# Every Go command needs the gtk3 build tag (only GTK3/WebKitGTK-4.1 is installed).
GOBIN := $(shell go env GOPATH)/bin
export PATH := $(PATH):$(GOBIN)
TAGS := gtk3

.PHONY: test vet typecheck bindings check build dev

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
