# 07: compose: chiudere la radice

**What to build:** Dopo lo split, `internal/compose` contiene solo `Manager`, la composizione e i metodi del modulo. Il commento di package è aggiornato, la direzione delle dipendenze è `model` ← `store`/`parse` ← radice, con `docker` indipendente, e non restano file o alias morti. Contesto: spec `06-sottopackage-compose-e-gui`, ADR 0006.

**Blocked by:** 04 (compose: estrarre docker), 05 (compose: estrarre store), 06 (compose: estrarre parse).

**Status:** ready-for-agent

- [ ] La radice non contiene più codice spostato né alias non usati.
- [ ] Nessun sottopackage importa `internal/compose`.
- [ ] `go test ./internal/compose/...` è verde.
- [ ] `go build ./...`, `go test ./...` e `go vet ./...` sono verdi.
