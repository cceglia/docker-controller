# 12: Notifiche D-Bus e click

**What to build:** Le notifiche desktop restano su D-Bus, con il mittente `internal/notify` e `github.com/godbus/dbus/v5`. La chiamata `org.freedesktop.Notifications.Notify` porta l'azione `default`: `[]string{"default", ""}`. Il mittente ascolta `org.freedesktop.Notifications.ActionInvoked` per l'id della notifica inviata. Sul click invoca una callback che mostra e alza la finestra. Le notifiche arrivano anche senza tray, perché passano dal bus di sessione. L'invio gira su una goroutine e non blocca il thread Qt. La decisione su cosa notificare non cambia: resta nel package puro `internal/tray`, con `Launching`, `Aborted` e `Observe`. Una notifica non compare due volte per la stessa Attività. `internal/notify` resta nel repository.

**Blocked by:** 06

**Status:** ready-for-agent

- [ ] La fine di un'Attività lanciata dalla tray produce una notifica
- [ ] Un'Attività fallita produce una notifica, da qualunque origine
- [ ] Il titolo della notifica è "Docker Controller" e il testo è quello di oggi
- [ ] Una stessa Attività non produce due notifiche
- [ ] `Notify` passa l'azione `default`: `[]string{"default", ""}`
- [ ] Il mittente ascolta `ActionInvoked` per l'id della notifica inviata
- [ ] Il click sulla notifica mostra e alza la finestra in primo piano
- [ ] La notifica arriva anche senza tray, perché passa dal bus di sessione
- [ ] L'avviso in sovrimpressione di un'Attività fallita resta quello del ticket 06: la notifica non lo sostituisce
- [ ] L'invio gira su una goroutine e non blocca il thread Qt
- [ ] Le tre chiamate restano quelle di `tray.Notifier`: `Launching`, `Aborted` e `Observe`
- [ ] La decisione su cosa notificare resta nel package puro della tray
- [ ] `internal/notify` resta nel repository, con `godbus` in `go.mod`
- [ ] Test esistenti del Notifier restano verdi con un mittente finto
- [ ] Verifica manuale su KDE: la notifica arriva e il click riporta la finestra in primo piano

## Comments

La regola "fine di un'Attività lanciata dalla tray, più ogni fallimento" resta in `internal/tray`, con i test di oggi.

`QSystemTrayIcon.ShowMessage` non riporta il click perché invia `Notify` con l'elenco `actions` vuoto. La misura sta nel ticket 01. Per questo il mittente resta `internal/notify`: passa l'azione `default` e ascolta `ActionInvoked`.
