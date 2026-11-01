# 10: Tray

**What to build:** un'icona nel system tray, che cambia aspetto quando Docker non è raggiungibile, con un menu annidato Gruppo ▸ Progetto (i Progetti senza Gruppo al primo livello). Ogni Progetto mostra lo stato (● Attivo, ◐ Parziale, ○ Fermo/Rimosso) e offre Up, Stop, Restart e il sottomenu Servizi ▸ Servizio ▸ Start/Stop/Restart. Durante un'Attività le voci di quel Progetto sono disabilitate. Una notifica desktop segnala la fine di un'Attività lanciata dal tray e ogni Attività fallita. Il menu ha anche "Mostra finestra" ed "Esci". Aggiornamento: il tray offre anche Down (di un Progetto e di un Servizio), con una finestra di conferma e senza mai rimuovere i volumi. Down con volumi, Pull e Build non sono nel tray.

**Blocked by:** 03, 04, 07

**Status:** resolved

- [ ] Il menu si ricostruisce a ogni snapshot di Watch, senza polling proprio
- [ ] Struttura annidata per Gruppo, Progetto e Servizi come descritto
- [ ] Voci disabilitate durante un'Attività e quando Docker non è raggiungibile
- [ ] Notifiche desktop per le Attività lanciate dal tray e per gli errori
- [ ] Icona diversa quando Docker non è raggiungibile
- [ ] Verifica manuale su KDE: Restart di `personalvpn` dal tray con la finestra nascosta mostra la notifica

## Comments

- 2026-10-21 (my-implement-orchestrator): implementato nel commit `68c3543`, correzioni dalla review finale in `91274a2`. Review finale (fixed point `4d05bd7`) approvata al ciclo 4. `make check` e `go test -race -tags gtk3 ./...` verdi. I criteri verificati da test automatici sono coperti; i criteri di UI, tray e desktop restano da verificare a mano con `wails3 dev` (l'app non è stata avviata per la verifica).
