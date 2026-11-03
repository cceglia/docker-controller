# 14: gui: estrarre traymenu

**What to build:** L'albero del menu della tray vive in `internal/gui/traymenu`, con `TrayNode` esportato. `qttray` usa il package. Il test dell'albero si sposta. Contesto: spec `06-sottopackage-compose-e-gui`, ADR 0006.

**Blocked by:** 08 (gui: estrarre view), 10 (gui: estrarre project).

**Status:** resolved

- [x] `traymenu` importa `view` e `project`, non la radice.
- [x] `TrayNode` è esportato e `traytree_test` vive in `traymenu`.
- [x] Il menu della tray della radice usa il nuovo package e `make test-ui` è verde.
- [x] `go build ./...`, `go test ./...` e `go vet ./...` sono verdi.
