# 11: gui: estrarre forms

**What to build:** I modelli dei form di aggiunta, modifica, impostazioni e chiusura vivono in `internal/gui/forms`. I dialoghi Qt della radice usano il package. I test si spostano: `addproject_test`, `edit_test` (senza il test di `DisplayNameOf`, già in `view`), `preferences_test` e `close_test`; `qtedit_test` resta in radice. Contesto: spec `06-sottopackage-compose-e-gui`, ADR 0006.

**Blocked by:** 02 (gui: creare guitest), 08 (gui: estrarre view).

**Status:** ready-for-agent

- [ ] `forms` importa `view`, non la radice.
- [ ] I test dei form vivono in `forms`; `qtedit_test` resta in radice.
- [ ] I dialoghi di aggiunta, modifica, impostazioni e chiusura usano il nuovo package.
- [ ] `go build ./...`, `go test ./...` e `go vet ./...` sono verdi.
