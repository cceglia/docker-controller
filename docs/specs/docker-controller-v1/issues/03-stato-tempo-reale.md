# 03: Stato in tempo reale e Docker non raggiungibile

**What to build:** lo stato dei Progetti si aggiorna appena un Container cambia, anche se il cambio arriva dal terminale, grazie a `docker events`. Un aggiornamento completo ogni 30 secondi fa da rete di sicurezza. Se Docker non risponde, la finestra mostra il banner "Docker non raggiungibile", la lista è in grigio e le Azioni sono disabilitate; l'app riprova ogni 5 secondi e torna operativa da sola.

**Blocked by:** 01

**Status:** resolved

- [ ] Un evento container fa aggiornare solo il Progetto interessato e Watch notifica (test con il fake in streaming)
- [ ] Watch non notifica se lo snapshot non è cambiato (test)
- [ ] Lo snapshot indica se Docker è raggiungibile; quando torna disponibile riprende tutto da solo (test)
- [ ] Banner, lista in grigio e Azioni disabilitate nella UI
- [ ] Verifica manuale: `docker compose stop` da terminale si riflette nell'app entro un secondo circa

## Comments

- 2026-10-21 (my-implement-orchestrator): implementato nel commit `3ce2969`, correzioni dalla review finale in `0b9f34f, 91274a2`. Review finale (fixed point `4d05bd7`) approvata al ciclo 4. `make check` e `go test -race -tags gtk3 ./...` verdi. I criteri verificati da test automatici sono coperti; i criteri di UI, tray e desktop restano da verificare a mano con `wails3 dev` (l'app non è stata avviata per la verifica).
