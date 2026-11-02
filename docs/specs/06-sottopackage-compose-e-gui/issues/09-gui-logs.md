# 09: gui: estrarre logs

**What to build:** Ciclo di vita degli stream, modello della vista Log, `serviceHueOf` e i messaggi degli errori di fine stream vivono in `internal/gui/logs`. La radice usa il package per la pagina Log; i test degli stream e della vista si spostano; la parte dei Log di `user_errors_test` si sposta qui. Contesto: spec `06-sottopackage-compose-e-gui`, ADR 0006.

**Blocked by:** 03 (gui: estrarre design), 08 (gui: estrarre view).

**Status:** ready-for-agent

- [ ] `logs` importa `design` e `view`, non la radice.
- [ ] `logstreams_test`, `logview_test` e la parte Log di `user_errors_test` vivono in `logs`.
- [ ] La pagina Log della radice usa il nuovo package e `make test-ui` è verde.
- [ ] `go build ./...`, `go test ./...` e `go vet ./...` sono verdi.
