# 15: gui: chiudere la radice

**What to build:** Dopo lo split, `internal/gui` contiene solo applicazione, shell, viste Qt e widget condivisi (`notify`, `startguard`, costanti di finestra). I simboli che i test di radice usano sono esportati, ogni package ha il commento di package, e nessun sottopackage importa `internal/gui`. Contesto: spec `06-sottopackage-compose-e-gui`, ADR 0006.

**Blocked by:** 03 (gui: estrarre design), 08 (gui: estrarre view), 09 (gui: estrarre logs), 10 (gui: estrarre project), 11 (gui: estrarre forms), 12 (gui: estrarre activity), 13 (gui: estrarre usage), 14 (gui: estrarre traymenu).

**Status:** ready-for-agent

- [ ] La radice non contiene più modelli o helper spostati.
- [ ] Nessun sottopackage importa `internal/gui`.
- [ ] `qtshell_test` compila con i simboli esportati.
- [ ] `go build ./...`, `go test ./...`, `go vet ./...` e `make test-ui` sono verdi.
