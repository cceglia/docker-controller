# 01: compose: estrarre model

**What to build:** Il codice di dominio di compose vive in un package `model` dentro `internal/compose`: `Project`, `Service`, `Container`, `Port`, `Snapshot`, `Status`, `ActionKind`, `ActivityID`, `ActivityRef`, `Settings`, `Discovered`, `LogRequest`, `LogLine`, gli helper puri `StatusOf` e `Folder` e le sentinelle usate dai sottopackage. La radice riesporta con alias solo la superficie usata da gui, tray e main, così quei chiamanti non cambiano. Contesto: spec `06-sottopackage-compose-e-gui`, ADR 0006.

**Blocked by:** None (can start immediately).

**Status:** resolved

- [x] `model` contiene i tipi, gli helper e le sentinelle elencati.
- [x] `StatusOf` e `Folder` hanno test di package.
- [x] `compose.Manager`, `Config`, `New` e le sentinelle mantengono la firma per i chiamanti esterni.
- [x] I test esistenti di compose passano senza cambiamenti di merito.
- [x] `go build ./...`, `go test ./...` e `go vet ./...` sono verdi.
- [x] Il package ha un commento che dice cosa contiene e la direzione delle dipendenze.
