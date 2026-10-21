# 05: Azioni distruttive

**What to build:** nel dettaglio compaiono Down, con un modal di conferma, e un pulsante rosso separato Down con volumi, il cui modal richiede di digitare il nome del Progetto. Un Progetto Rimosso può essere tolto con "Rimuovi dalla lista" (anche qui con modal), che non tocca Docker. Se i File compose di un Progetto non esistono più, il Progetto mostra l'avviso "File compose mancante" e "Rimuovi dalla lista" viene messo in evidenza.

**Blocked by:** 02

**Status:** resolved

- [ ] Down corrisponde a `down`, Down con volumi a `down -v` (test sugli argomenti)
- [ ] Dopo un Down il Progetto resta in lista come Rimosso (test)
- [ ] "Rimuovi dalla lista" viene rifiutato se il Progetto ha Container (test)
- [ ] File compose mancante viene rilevato nello snapshot (test con un file temporaneo cancellato)
- [ ] Modal di conferma per tutte e tre le Azioni distruttive; Down con volumi si abilita solo dopo aver digitato il nome esatto

## Comments

- 2026-10-21 (my-implement-orchestrator): implementato nel commit `b022d84`, correzioni dalla review finale in `0b9f34f`. Review individuale (rischio perdita dati, `down -v`): approvata, nessun blocco. Review finale (fixed point `4d05bd7`) approvata al ciclo 4. `make check` e `go test -race -tags gtk3 ./...` verdi. I criteri verificati da test automatici sono coperti; i criteri di UI, tray e desktop restano da verificare a mano con `wails3 dev` (l'app non è stata avviata per la verifica).
