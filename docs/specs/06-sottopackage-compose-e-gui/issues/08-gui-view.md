# 08: gui: estrarre view

**What to build:** `internal/gui/view` è il package foglia della presentazione: formattazione numerica, stato e tacche del Progetto, sezioni della lista, errori generici, apertura di cartelle, terminali e URL, e i testi condivisi `ProjectName`, `DisplayNameOf` e `MissingComposeText`. `view` non importa nessun package di feature. I test si spostano o si spezzano: `open_test` e la parte generica di `user_errors_test` vanno qui, il test di `DisplayNameOf` si separa da `edit_test`. Contesto: spec `06-sottopackage-compose-e-gui`, ADR 0006.

**Blocked by:** 02 (gui: creare guitest).

**Status:** ready-for-agent

- [ ] `view` non importa `logs`, `project`, `forms`, `activity`, `usage` o `traymenu`.
- [ ] `ProjectName`, `DisplayNameOf` e `MissingComposeText` vivono in `view`.
- [ ] `open_test` usa `guitest`; la parte dei Log di `user_errors_test` resta fuori da `view`.
- [ ] `go build ./...`, `go test ./...` e `go vet ./...` sono verdi.
