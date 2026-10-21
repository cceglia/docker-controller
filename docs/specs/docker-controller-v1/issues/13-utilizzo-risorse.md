# 13: Utilizzo risorse

**What to build:** in cima alla finestra una barra mostra l'Utilizzo risorse globale di Docker: CPU e memoria di tutti i Container in esecuzione, sia in valore assoluto sia in percentuale sul totale della macchina, più rete in/out e I/O disco. Dentro il dettaglio del Progetto, ogni Servizio mostra il proprio Utilizzo risorse sommato sui suoi Container; un Servizio senza Container in esecuzione mostra "—". I valori si aggiornano circa ogni 2 secondi. Lo streaming parte quando la finestra è visibile e si ferma quando è nascosta. Con Docker non raggiungibile la barra mostra lo stato di errore.

**Blocked by:** 03, 04

**Status:** resolved

- [ ] Il modulo `compose` espone l'Utilizzo risorse con un flusso separato da Watch, alimentato da `docker stats --format json` in streaming (test con il fake in streaming e output reali catturati)
- [ ] Il parsing delle unità di `docker stats` è corretto (es. `1.5GiB`, `12.3kB`, `0B`) (test)
- [ ] La somma per Servizio con repliche è corretta e un Servizio fermo non ha misure (test)
- [ ] La CPU globale è normalizzata sul numero di CPU e la memoria confrontata con il totale della macchina da `docker info` (test)
- [ ] I campioni non provocano notifiche di Watch né la ricostruzione del tray (test)
- [ ] Lo streaming si ferma con la finestra nascosta e riparte quando torna visibile
- [ ] Barra globale in cima alla finestra e misure per Servizio nel dettaglio

## Comments

- 2026-10-21 (my-implement-orchestrator): implementato nel commit `9dda2ef`, correzioni dalla review finale in `0b9f34f`. Review finale (fixed point `4d05bd7`) approvata al ciclo 4. `make check` e `go test -race -tags gtk3 ./...` verdi. I criteri verificati da test automatici sono coperti; i criteri di UI, tray e desktop restano da verificare a mano con `wails3 dev` (l'app non è stata avviata per la verifica).
