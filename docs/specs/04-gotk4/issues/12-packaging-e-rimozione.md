# 12: Packaging e rimozione di frontend e Wails

**What to build:** La build non usa più `wails3`, npm o Node. I prerequisiti diventano GTK4 e libadwaita. L'installazione mette binario, voce desktop con l'app ID e icone. Alla fine spariscono frontend, dipendenze Wails e codice morto della vecchia shell.

**Blocked by:** 05, 06, 07, 08, 09, 10, 11

**Status:** ready-for-agent

- [ ] La build e la verifica girano senza `wails3` e senza npm
- [ ] `go vet ./...` e `go test ./...` sono verdi
- [ ] I prerequisiti documentati sono build-essential, pkg-config, libgtk-4-dev, libadwaita-1-dev
- [ ] `make install` installa binario, voce desktop `io.github.cceglia.DockerController.desktop` e icone
- [ ] La voce compare nel menu di KDE con nome e icona corretti
- [ ] `make uninstall` rimuove i file installati, compresa la vecchia voce `docker-controller.desktop`, e lascia il database
- [ ] Frontend Svelte, CLI `wails3`, Node e dipendenze Wails sono rimossi dal repository
- [ ] Nessun codice della vecchia shell resta nel repository
- [ ] La checklist manuale della spec è completata su KDE
