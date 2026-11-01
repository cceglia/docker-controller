# 01: Prototipo miqt e tray

**What to build:** Un'app minima in miqt dimostra che la scelta regge su questa macchina e su KDE. Il prototipo compila contro Qt 6.10.2, apre una finestra vuota, crea un'icona di sistema con un menu di due voci e mostra una notifica con `QSystemTrayIcon.ShowMessage`. Verifica il click sinistro e il tasto destro, il click sulla notifica, un dialogo modale aperto da una voce della tray e un aggiornamento della finestra da una goroutine con `mainthread.Start`. Prova tutto su X11 e su Wayland. Alla fine del ticket il prototipo non resta nel repository: i risultati finiscono nella spec.

**Blocked by:** None (can start immediately)

**Status:** ready-for-agent

- [ ] `qt6-base-dev` è installato e `pkg-config --modversion Qt6Widgets` risponde
- [ ] `go get github.com/mappu/miqt@v0.15.0` compila un programma minimo con cgo
- [ ] Il prototipo apre una `QMainWindow` vuota e la chiude senza errori
- [ ] Il prototipo crea un `QSystemTrayIcon` con tooltip "Docker Controller"
- [ ] Il menu della tray ha due voci e risponde al click
- [ ] Il click sinistro mostra la finestra e il tasto destro apre il menu
- [ ] `QSystemTrayIcon.ShowMessage` mostra una notifica desktop con titolo e testo
- [ ] Il click sulla notifica richiama la finestra
- [ ] Una voce del menu apre un dialogo modale sul thread Qt
- [ ] Una goroutine aggiorna un'etichetta con `mainthread.Start`
- [ ] L'icona, il menu e la notifica funzionano su X11
- [ ] L'icona, il menu e la notifica funzionano su Wayland, oppure il limite è scritto
- [ ] Il tempo del primo build con cgo è misurato e scritto
- [ ] La memoria del processo a riposo è misurata con la finestra aperta
- [ ] L'esito è scritto nella spec e la versione di miqt è fissata
- [ ] Il prototipo non resta nel repository

## Comments

Ticket di rischio. Se miqt non compila contro Qt 6.10.2, si fissa il commit più recente di miqt e si aggiorna la spec prima del port. Se la tray non funziona su Wayland, la spec dichiara Wayland non supportato e il port resta su X11.
