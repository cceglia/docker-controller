# 13: gui: estrarre usage

**What to build:** Feed, stream, stato e testi dell'Utilizzo risorse vivono in `internal/gui/usage`, con `MeterValue` accanto al calcolo della barra. La radice usa il package per la barra e il dettaglio. I test si spostano: `usagefeed_test`, `usagestream_test`, `usagetext_test` e `usagebar_test`. Contesto: spec `06-sottopackage-compose-e-gui`, ADR 0006.

**Blocked by:** 02 (gui: creare guitest), 08 (gui: estrarre view).

**Status:** ready-for-agent

- [ ] `usage` importa `view`, non la radice.
- [ ] `MeterValue` vive in `usage` e `usagebar_test` si sposta con lui.
- [ ] La barra e il dettaglio della radice usano il nuovo package.
- [ ] `go build ./...`, `go test ./...` e `go vet ./...` sono verdi.
