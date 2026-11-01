# 12: Packaging e rimozione di frontend e Wails

**What to build:** La build non usa più `wails3`, npm o Node. I prerequisiti diventano GTK4 e libadwaita. L'installazione mette binario, voce desktop con l'app ID e icone. Alla fine spariscono frontend, dipendenze Wails e codice morto della vecchia shell.

**Blocked by:** 05, 06, 07, 08, 09, 10, 11

**Status:** resolved

- [x] La build e la verifica girano senza `wails3` e senza npm
- [x] `go vet ./...` e `go test ./...` sono verdi
- [x] I prerequisiti documentati sono build-essential, pkg-config, libgtk-4-dev, libadwaita-1-dev
- [x] `make install` installa binario, voce desktop `io.github.cceglia.DockerController.desktop` e icone
- [ ] La voce compare nel menu di KDE con nome e icona corretti
- [x] `make uninstall` rimuove i file installati, compresa la vecchia voce `docker-controller.desktop`, e lascia il database
- [x] Frontend Svelte, CLI `wails3`, Node e dipendenze Wails sono rimossi dal repository
- [x] Nessun codice della vecchia shell resta nel repository
- [ ] La checklist manuale della spec è completata su KDE

## Comments

Implementato nel commit `bc247ee`.

Installazione e disinstallazione provate su un prefisso temporaneo: i file
nuovi, la vecchia `docker-controller.desktop` e le vecchie icone spariscono, il
database resta.

Prerequisiti: ai quattro della spec si aggiunge `gobject-introspection`, che
fornisce la voce pkg-config `gobject-introspection-1.0` e `libgirepository-1.0.so`
richiesti da `gotk4`. Senza quel pacchetto la build non parte.

Restano aperti la voce nel menu di KDE e la checklist manuale della spec.
