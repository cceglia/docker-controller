# 10: gui: estrarre project

**What to build:** Azioni ed etichette, `Confirmation`, rimozione dalla lista, formattazione del dettaglio e stato del disegno vivono in `internal/gui/project`. Le viste Qt del dettaglio e della lista usano il package. I test si spostano. Contesto: spec `06-sottopackage-compose-e-gui`, ADR 0006.

**Blocked by:** 08 (gui: estrarre view).

**Status:** ready-for-agent

- [ ] `project` importa `view`, non la radice.
- [ ] `actions_test`, `remove_test`, `detail_test` e `detailstate_test` vivono in `project`.
- [ ] La lista, il dettaglio e le tacche della radice usano il nuovo package.
- [ ] `go build ./...`, `go test ./...` e `go vet ./...` sono verdi.
