# 01: Scheletro: lista dei Progetti scoperti

**What to build:** all'avvio, l'app apre una finestra con l'elenco di tutti i Progetti che Docker conosce, ciascuno con nome, cartella, stato (Attivo, Parziale, Fermo, Rimosso) e numero di Container in esecuzione sul totale. I Progetti scoperti vengono salvati in automatico nel DB, quindi restano in lista ai riavvii successivi anche quando Docker non li vede più (in quel caso con stato Rimosso). È il tracer bullet che attraversa tutti i livelli: CLI Docker → modulo `compose` → SQLite → binding Wails → frontend.

**Blocked by:** None (can start immediately)

**Status:** resolved

- [ ] Progetto Wails v3 con frontend Svelte + TypeScript, versione di Wails fissata come da ADR-0001; `wails3 dev` avvia l'app
- [ ] Il modulo `compose` espone snapshot e Watch secondo la spec; il port verso Docker ha l'adapter a processi esterni e un fake per i test
- [ ] Le fixture JSON sono catturate da `compose ls -a` / `compose ps -a` dei Progetti reali del computer
- [ ] Il DB SQLite viene creato in `~/.local/share/docker-controller/docker-controller.db` con migrazioni versionate
- [ ] Un Progetto scoperto viene salvato e, se sparisce da Docker, resta nello snapshot come Rimosso
- [ ] I quattro stati sono calcolati correttamente (test con le fixture di `sunlinkdocker` e `stellantisvpn`)
- [ ] Il frontend riceve lo snapshot tramite un evento e lo mostra tramite il modulo store
- [ ] Il tema chiaro/scuro segue quello del sistema
- [ ] `go test ./...` passa

## Comments

- 2026-10-21 (my-implement-orchestrator): implementato nel commit `a4f0981`, correzioni dalla review finale in `0b9f34f, 91274a2`. Review finale (fixed point `4d05bd7`) approvata al ciclo 4. `make check` e `go test -race -tags gtk3 ./...` verdi. I criteri verificati da test automatici sono coperti; i criteri di UI, tray e desktop restano da verificare a mano con `wails3 dev` (l'app non è stata avviata per la verifica).
