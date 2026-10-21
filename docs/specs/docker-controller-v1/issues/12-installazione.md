# 12: Installazione

**What to build:** un solo comando (`make install`) compila l'app e installa in `~/.local` il binario, il file `.desktop` e l'icona, così l'app compare nel menu di KDE. Il README elenca i prerequisiti di build.

**Blocked by:** 01

**Status:** ready-for-agent

- [ ] `make install` produce il binario e installa binario, `.desktop` e icona nelle cartelle XDG dell'utente
- [ ] `make uninstall` rimuove i file installati, ma non il DB
- [ ] L'app si avvia dal menu di KDE
- [ ] Il README elenca i prerequisiti (`libgtk-3-dev`, `libwebkit2gtk-4.1-dev`, `build-essential`, `pkg-config`, `wails3`)
