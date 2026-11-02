# 03: gui: estrarre design

**What to build:** Token di design, foglio di stile e colore della palette dei Servizi vivono in `internal/gui/design`. Il package espone `DesignTheme`, `LightTheme`, `DarkTheme`, `TargetMin`, `ApplicationStyleSheet` e `ServiceColor`; la radice li usa. Il test dei token e gli helper di contrasto si spostano nel package. Contesto: spec `06-sottopackage-compose-e-gui`, ADR 0006.

**Blocked by:** None (can start immediately).

**Status:** ready-for-agent

- [ ] `design` contiene token, foglio di stile e `ServiceColor`; non importa `internal/gui`.
- [ ] `designtokens_test` e gli helper di contrasto vivono in `design`.
- [ ] `qtshell_test` usa i simboli esportati e `make test-ui` è verde.
- [ ] `go build ./...`, `go test ./...` e `go vet ./...` sono verdi.
