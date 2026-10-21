# 12: Installazione

**What to build:** un solo comando (`make install`) compila l'app e installa in `~/.local` il binario, il file `.desktop` e l'icona, così l'app compare nel menu di KDE. Il README elenca i prerequisiti di build.

**Blocked by:** 01

**Status:** resolved

- [ ] `make install` produce il binario e installa binario, `.desktop` e icona nelle cartelle XDG dell'utente
- [ ] `make uninstall` rimuove i file installati, ma non il DB
- [ ] L'app si avvia dal menu di KDE
- [ ] Il README elenca i prerequisiti (`libgtk-3-dev`, `libwebkit2gtk-4.1-dev`, `build-essential`, `pkg-config`, `wails3`)

## Comments

- 2026-10-21 (my-implement-orchestrator): implementato nel commit `4f89284`, correzioni dalla review finale in `0b9f34f`. Install/uninstall verificati solo su un PREFIX temporaneo. Review finale (fixed point `4d05bd7`) approvata al ciclo 4. `make check` e `go test -race -tags gtk3 ./...` verdi. I criteri verificati da test automatici sono coperti; i criteri di UI, tray e desktop restano da verificare a mano con `wails3 dev` (l'app non è stata avviata per la verifica).
