# 04: Dettaglio del Progetto e azioni sui Servizi

**What to build:** selezionando un Progetto si apre il dettaglio, con i Container raggruppati per Servizio (immagine, stato, health, porte pubblicate). Ogni Servizio ha i pulsanti Start, Stop e Restart, che agiscono su tutte le sue repliche come un'Attività del Progetto. Una porta pubblicata apre `http://localhost:<porta>` nel browser. "Apri cartella" apre la cartella del Progetto nel file manager.

**Blocked by:** 02

**Status:** resolved

- [ ] Lo snapshot contiene i Container per Servizio con stato, health e porte (test con le fixture di `sunlinkdocker`)
- [ ] Un Servizio con repliche mostra tutti i suoi Container
- [ ] Le Azioni sul Servizio passano il nome del Servizio a compose (test sugli argomenti)
- [ ] Le Azioni sul Servizio rispettano la regola di una sola Attività per Progetto
- [ ] Porte cliccabili e "Apri cartella" funzionanti

## Comments

- 2026-10-21 (my-implement-orchestrator): implementato nel commit `8445f4b`. Review finale (fixed point `4d05bd7`) approvata al ciclo 4. `make check` e `go test -race -tags gtk3 ./...` verdi. I criteri verificati da test automatici sono coperti; i criteri di UI, tray e desktop restano da verificare a mano con `wails3 dev` (l'app non è stata avviata per la verifica).
