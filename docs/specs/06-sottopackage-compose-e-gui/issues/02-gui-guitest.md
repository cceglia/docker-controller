# 02: gui: creare guitest

**What to build:** Un package `internal/gui/guitest` raccoglie gli helper condivisi dai test della GUI: i costruttori degli eventi di Attività, l'asserzione su due slice di stringhe e l'attesa di un segnale. I test di radice che oggi definiscono quegli helper li importano da qui, senza duplicarli. Contesto: spec `06-sottopackage-compose-e-gui`, ADR 0006.

**Blocked by:** None (can start immediately).

**Status:** resolved

- [x] `guitest` contiene solo helper, non test.
- [x] `notify_test` e `qtshell_test` usano gli helper condivisi.
- [x] `activitylog_test`, `addproject_test` e `usagestream_test` non definiscono più gli stessi helper.
- [x] `go build ./...`, `go test ./...` e `go vet ./...` sono verdi.
