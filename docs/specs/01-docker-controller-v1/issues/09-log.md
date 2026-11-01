# 09: Log

**What to build:** nel dettaglio del Progetto un pannello log carica le ultime 500 righe e poi segue in tempo reale. Le righe hanno un colore per Servizio e si possono filtrare per Servizio. Ci sono pausa dello scorrimento, ricerca nel testo e un interruttore per i timestamp. Il pannello tiene al massimo 5000 righe e lo streaming si ferma quando viene chiuso.

**Blocked by:** 04

**Status:** resolved

- [ ] Logs del modulo `compose` consegna righe già attribuite al Servizio e termina alla cancellazione (test con il fake in streaming)
- [ ] Il filtro per Servizio passa il Servizio a compose (test sugli argomenti)
- [ ] Il buffer del frontend non supera le 5000 righe
- [ ] La chiusura del pannello termina il processo `compose logs`
- [ ] Colori, pausa, ricerca e timestamp funzionanti

## Comments

- 2026-10-21 (my-implement-orchestrator): implementato nel commit `f50ad1d`, correzioni dalla review finale in `0b9f34f, 81b5f9a`. Review finale (fixed point `4d05bd7`) approvata al ciclo 4. `make check` e `go test -race -tags gtk3 ./...` verdi. I criteri verificati da test automatici sono coperti; i criteri di UI, tray e desktop restano da verificare a mano con `wails3 dev` (l'app non è stata avviata per la verifica).
