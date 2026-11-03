# 05: compose: estrarre store

**What to build:** La persistenza dei Progetti vive in `internal/compose/store`: `Store`, `New`, `UpsertDiscovered`, `Add`, `UpdateSettings`, `List` e `Remove`. `Manager` riceve lo `Store` e continua a comportarsi come prima. I test aprono un database SQLite temporaneo con le migrazioni esistenti. Contesto: spec `06-sottopackage-compose-e-gui`, ADR 0006.

**Blocked by:** 01 (compose: estrarre model).

**Status:** resolved

- [x] `store` contiene la persistenza e non importa la radice.
- [x] I test di `store` coprono salvataggio, elenco, impostazioni e rimozione su un database temporaneo.
- [x] I test di `Manager` passano senza cambiare il comportamento osservabile.
- [x] `go build ./...`, `go test ./...` e `go vet ./...` sono verdi.
