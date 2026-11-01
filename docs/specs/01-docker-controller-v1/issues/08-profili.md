# 08: Profili

**What to build:** il dettaglio del Progetto mostra i Profili disponibili come interruttori. I Profili attivi vengono ricordati e usati da Up, Stop, Restart, Pull, Build e dalle Azioni sui Servizi. Down e Down con volumi agiscono invece su tutti i Profili. Cambiare un interruttore non avvia né ferma nulla.

**Blocked by:** 02

**Status:** resolved

- [ ] I Profili disponibili sono letti da `compose config --profiles` ed esposti nello snapshot (test)
- [ ] I Profili attivi sono passati come `--profile` alle Azioni (test sugli argomenti)
- [ ] Down e Down con volumi passano tutti i Profili (test sugli argomenti)
- [ ] I Profili attivi sopravvivono alla riapertura del DB (test)
- [ ] Cambiare i Profili non lancia alcuna Attività (test)
- [ ] Interruttori nella UI

## Comments

- 2026-10-21 (my-implement-orchestrator): implementato nel commit `af2ef1f`. Review finale (fixed point `4d05bd7`) approvata al ciclo 4. `make check` e `go test -race -tags gtk3 ./...` verdi. I criteri verificati da test automatici sono coperti; i criteri di UI, tray e desktop restano da verificare a mano con `wails3 dev` (l'app non è stata avviata per la verifica).
