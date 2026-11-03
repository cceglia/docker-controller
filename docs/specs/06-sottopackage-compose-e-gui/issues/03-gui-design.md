# 03: gui: estrarre design

**What to build:** Token di design, foglio di stile e colore della palette dei Servizi vivono in `internal/gui/design`. Il package espone `DesignTheme`, `LightTheme`, `DarkTheme`, `TargetMin`, `ApplicationStyleSheet` e `ServiceColor`; la radice li usa. Il test dei token e gli helper di contrasto si spostano nel package. Contesto: spec `06-sottopackage-compose-e-gui`, ADR 0006.

**Blocked by:** None (can start immediately).

**Status:** resolved

- [x] `design` contiene token, foglio di stile e `ServiceColor`; non importa `internal/gui`.
- [x] `designtokens_test` e gli helper di contrasto vivono in `design`.
- [x] `qtshell_test` usa i simboli esportati e `make test-ui` è verde.
- [x] `go build ./...`, `go test ./...` e `go vet ./...` sono verdi.
