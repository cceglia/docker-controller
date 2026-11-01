# 01: Prototipo miqt e tray

**What to build:** Un'app minima in miqt dimostra che la scelta regge su questa macchina e su KDE. Il prototipo compila contro Qt 6.10.2, apre una finestra vuota, crea un'icona di sistema con un menu di due voci e mostra una notifica desktop. Verifica il click sinistro e il tasto destro, il click sulla notifica, un dialogo modale aperto da una voce della tray e un aggiornamento della finestra da una goroutine con `mainthread.Start`. Prova tutto sul desktop KDE. Alla fine del ticket il prototipo non resta nel repository: i risultati finiscono nella spec.

**Blocked by:** None (can start immediately)

**Status:** resolved

- [x] `qt6-base-dev` è installato e `pkg-config --modversion Qt6Widgets` risponde
- [x] `go get github.com/mappu/miqt@v0.15.0` compila un programma minimo con cgo
- [x] Il prototipo apre una `QMainWindow` vuota e la chiude senza errori
- [x] Il prototipo crea un `QSystemTrayIcon` con tooltip "Docker Controller"
- [x] Il menu della tray ha due voci e risponde al click
- [x] Il click sinistro mostra la finestra e il tasto destro apre il menu
- [x] La notifica arriva su D-Bus con `org.freedesktop.Notifications.Notify`, con titolo e testo corretti. `QSystemTrayIcon.ShowMessage` lascia vuoto l'elenco `actions`
- [x] Il click sulla notifica usa l'azione `default` e l'ascolto di `ActionInvoked` su D-Bus. È la via scelta, perché `ShowMessage` non riporta il click
- [x] Una voce del menu apre un dialogo modale sul thread Qt
- [x] Una goroutine aggiorna un'etichetta con `mainthread.Start`
- [x] L'icona, il menu e la notifica funzionano su Wayland con KDE
- [x] La verifica manuale dell'utente è su Wayland: X11 resta senza misura in questo ticket
- [x] Il tempo del primo build con cgo è misurato e scritto
- [x] La memoria del processo a riposo è misurata con la finestra aperta
- [x] L'esito è scritto nella spec e la versione di miqt è fissata
- [x] Il prototipo non resta nel repository

## Comments

Ticket di rischio. Se miqt non compila contro Qt 6.10.2, si fissa il commit più recente di miqt e si aggiorna la spec prima del port. Se la tray non funziona su Wayland, la spec dichiara Wayland non supportato e il port resta su X11.

## Esito

Il prototipo sta in `/tmp/opencode/miqt-proto` e non entra nel repository.

**Build e binario.** miqt `v0.15.0` compila e gira contro Qt 6.10.2, con Go 1.27.3 e g++ 15.2.0. Il primo build con la cache Go vuota dura 256,53 s. Il build con la cache calda dura 7,3 s. Il binario pesa 100 MB senza strip e 28,8 MB con `-s -w`. La versione di miqt resta fissata a `v0.15.0`.

**Memoria.** A riposo su Wayland, con la finestra aperta: PSS 27,0 MB, RSS 136,7 MB, 11 thread.

**Tray su Wayland.** `QSystemTrayIcon_IsSystemTrayAvailable()` risponde `true`. `org.kde.StatusNotifierWatcher` è presente, con `kded6` come proprietario e `plasmashell` come host. L'oggetto `/StatusNotifierItem` è esportato e attivo. Il click sinistro riporta la finestra in primo piano.

**Menu su Wayland.** Il tasto destro apre il menu. Il menu lo disegna KDE tramite `com.canonical.dbusmenu` sull'oggetto `/MenuBar`. Qt non apre un popup proprio e stampa `Failed to create grabbing popup`: è un limite di Wayland, non un difetto del codice.

**Notifica.** `org.freedesktop.Notifications.Notify` arriva al bus con titolo e testo corretti.

**Verifica manuale.** Sul desktop Wayland dell'utente tutto funziona, tranne il click sulla notifica.

**Causa del click mancante.** `QSystemTrayIcon.ShowMessage` invia `Notify` con l'elenco `actions` vuoto. KDE non può quindi segnalare il click, e `QSystemTrayIcon.MessageClicked` non arriva mai. Anche `internal/notify` invia oggi `actions` vuoto e non ascolta `ActionInvoked`: nemmeno la versione GTK ha mai avuto il click.

**Decisione sulle notifiche.** Le notifiche tornano su D-Bus, con il mittente `internal/notify` e `github.com/godbus/dbus/v5`. La chiamata `Notify` passa l'azione `default`: `[]string{"default", ""}`. Il mittente ascolta `ActionInvoked` per l'id della notifica inviata e sul click invoca una callback che mostra e alza la finestra. Le notifiche arrivano anche senza tray, perché passano dal bus di sessione. L'invio resta su una goroutine e non blocca il thread Qt. La decisione su cosa notificare resta nel package puro `internal/tray`, con `Launching`, `Aborted` e `Observe`. `internal/notify` non esce dal repository e `godbus` resta in `go.mod`.
