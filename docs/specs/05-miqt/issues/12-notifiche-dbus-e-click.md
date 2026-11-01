# 12: Notifiche D-Bus e click

**What to build:** Le notifiche desktop restano su D-Bus, con il mittente `internal/notify` e `github.com/godbus/dbus/v5`. La chiamata `org.freedesktop.Notifications.Notify` porta l'azione `default`: `[]string{"default", ""}`. Il mittente ascolta `org.freedesktop.Notifications.ActionInvoked` per l'id della notifica inviata. Sul click invoca una callback che mostra e alza la finestra. Le notifiche arrivano anche senza tray, perché passano dal bus di sessione. L'invio gira su una goroutine e non blocca il thread Qt. La decisione su cosa notificare non cambia: resta nel package puro `internal/tray`, con `Launching`, `Aborted` e `Observe`. Una notifica non compare due volte per la stessa Attività. `internal/notify` resta nel repository.

**Blocked by:** 06

**Status:** resolved

- [x] La fine di un'Attività lanciata dalla tray produce una notifica
- [x] Un'Attività fallita produce una notifica, da qualunque origine
- [x] Il titolo della notifica è "Docker Controller" e il testo è quello di oggi
- [x] Una stessa Attività non produce due notifiche
- [x] `Notify` passa l'azione `default`: `[]string{"default", ""}`
- [x] Il mittente ascolta `ActionInvoked` per l'id della notifica inviata
- [x] Il click sulla notifica mostra e alza la finestra in primo piano
- [x] La notifica arriva anche senza tray, perché passa dal bus di sessione
- [x] L'avviso in sovrimpressione di un'Attività fallita resta quello del ticket 06: la notifica non lo sostituisce
- [x] L'invio gira su una goroutine e non blocca il thread Qt
- [x] Le tre chiamate restano quelle di `tray.Notifier`: `Launching`, `Aborted` e `Observe`
- [x] La decisione su cosa notificare resta nel package puro della tray
- [x] `internal/notify` resta nel repository, con `godbus` in `go.mod`
- [x] Test esistenti del Notifier restano verdi con un mittente finto
- [x] Verifica manuale su KDE: la notifica arriva e il click riporta la finestra in primo piano

## Comments

La regola "fine di un'Attività lanciata dalla tray, più ogni fallimento" resta in `internal/tray`, con i test di oggi.

`QSystemTrayIcon.ShowMessage` non riporta il click perché invia `Notify` con l'elenco `actions` vuoto. La misura sta nel ticket 01. Per questo il mittente resta `internal/notify`: passa l'azione `default` e ascolta `ActionInvoked`.

## Esito

Una notifica parte con l'azione `default`, un click su di essa riporta la finestra
in primo piano.

### Struttura

| File | Ruolo |
| --- | --- |
| `internal/notify/notify.go` | Il mittente D-Bus: invio, `pending`, ascolto di `ActionInvoked`, callback del click |
| `internal/notify/notify_test.go` | La regola degli id, la gestione del segnale, gli argomenti di `Notify` |
| `internal/gui/notify.go` | `Notifier.OnClick`: inoltra la callback a un mittente che sa riportare il click |
| `internal/gui/qtapp.go` | Registra il click sulla finestra dopo `newShell` |
| `internal/gui/notify_test.go` | L'inoltro della callback e il caso di un mittente che non sa riportarla |

`internal/tray`, `main.go`, `internal/compose`, `internal/db` e
`internal/settings` non cambiano. `main.go` crea già il `Notifier` prima della
finestra, quindi l'ordine del ticket è quello attuale.

### Mittente

`notifyArgs(title, body)` è l'elenco degli argomenti di `Notify`, nell'ordine
dell'interfaccia. Alla posizione 5 c'è l'azione: `[]string{"default", ""}`.
L'etichetta è vuota, quindi il desktop non disegna un pulsante, e il click cade
sul corpo della notifica. Il resto è quello di prima: nome applicazione, id di
sostituzione 0, icona vuota, titolo, testo, hint vuoti e timeout -1.

`DBus.Send` resta non bloccante: apre una goroutine per ogni notifica. Dentro
la goroutine `notify` apre il bus di sessione, avvia l'ascolto, fa la chiamata e
tiene l'id che il servizio restituisce:

- l'ascolto parte **prima** della chiamata, così un click immediato trova il
  canale già aperto;
- `listen` registra la regola di match
  `interface='org.freedesktop.Notifications', member='ActionInvoked'` e apre una
  goroutine che legge i segnali. Un campo `listening` sotto mutex lo fa partire
  una volta sola, quindi la seconda notifica non aggiunge né una regola né una
  goroutine;
- un `AddMatchSignal` rifiutato non ferma l'invio: il mittente scrive la riga di
  log e manda la notifica lo stesso. Perde solo il click;
- un bus di sessione assente o un servizio che non risponde danno una riga di
  log e l'app prosegue, come prima.

La regola degli id sta in `pending`, un tipo puro senza bus e senza Qt.
`add(id)` registra un id in attesa, `take(id)` lo restituisce una volta sola.
`handle(signal)` applica la regola: accetta solo
`org.freedesktop.Notifications.ActionInvoked` con due argomenti, un `uint32` e
la stringa `default`. Poi chiama `take(id)`: un id mai inviato, o già consumato,
non produce niente. La callback gira fuori dal lock, quindi una callback lenta
non ferma la lettura dei segnali.

Il set degli id è limitato a `pendingLimit`, 64. Una notifica che l'utente
ignora non esce mai dal set, che altrimenti crescerebbe per tutta la sessione.
A set pieno esce l'id più vecchio: un click su quell'id non fa niente, mentre
il tetto tiene la memoria costante e la regola sui 64 id resta una sola.

La callback si registra con `DBus.SetOnClick(func())`. `Notifier.OnClick` la
inoltra quando il mittente implementa `notify.Clickable`; con un mittente che
non sa riportare il click la chiamata non fa niente e la notifica compare
comunque.

### Click

`qtapp.go` chiama `notifier.OnClick` subito dopo `newShell`, prima di `startTray`:

```go
notifier.OnClick(func() { mainthread.Start(a.present) })
```

Il segnale arriva sulla goroutine D-Bus, quindi la finestra si tocca solo sul
thread Qt. `a.present` è lo stesso metodo del click sinistro della tray:
`ShowNormal`, `Raise`, `ActivateWindow`. Senza tray la registrazione avviene lo
stesso, perché non dipende da `a.tray`.

### Test

- `internal/notify/notify_test.go`, `TestPending`: un id mai inviato non si
  consuma, un id inviato si consuma, un id consumato non si consuma di nuovo.
- `internal/notify/notify_test.go`, `TestDBus_Handle`: tabella sul gestore del
  segnale. Un click `default` su un id inviato produce un click; un id di
  un'altra applicazione, un secondo click sullo stesso id, un'azione diversa,
  un altro segnale, un segnale con un solo argomento e un id di tipo `int32`
  non producono niente.
- `internal/notify/notify_test.go`, `TestNotifyArgs`: gli argomenti sono otto,
  l'azione è `[]string{"default", ""}`, titolo e testo sono quelli passati.
- `internal/gui/notify_test.go`, `TestNotifier_OnClick`: la callback arriva a un
  mittente che sa riportare il click. `TestNotifier_OnClickWithoutAClickableSender`
  controlla che un mittente che non sa riportarlo non rompa niente.
- I test di `internal/tray` e `TestNotifier` restano invariati e verdi.

### Comandi

```
gofmt -l .                      # pulito sui file tracciati (restano tre file
                                # già sporchi sotto .agents/tmp/scaf, non toccati)
go vet ./...
make check
make test-ui
```

### Prove

KDE Plasma su Wayland, app avviata con `QT_QPA_PLATFORM=xcb`, quindi la finestra
è un client XWayland e `xprop` la vede. `dbus-monitor --session` aperto su tutto
il bus.

1. Click su "Up" del Progetto `vakt` nel menu della tray (il suo File compose
   manca, quindi l'Attività fallisce): `com.canonical.dbusmenu.Event(131,
   "clicked")`.
2. La chiamata sul bus:

```
method call ... interface=org.freedesktop.Notifications; member=Notify
   string "Docker Controller"
   uint32 0
   string ""
   string "Docker Controller"
   string "Up di vakt fallito: open /home/cristiano/Personal/Dockers/vakt/compose.yaml: no such file or directory"
   array [
      string "default"
      string ""
   ]
   ...
method return ... reply_serial=4
   uint32 395
```

3. `ActionInvoked` con un id inventato, finestra coperta da una finestra
   `xmessage`:

```
gdbus emit --session --object-path /org/freedesktop/Notifications \
  --signal org.freedesktop.Notifications.ActionInvoked "uint32 999999" "default"
```

   Ordine di impilamento prima e dopo: `0x1c0000f, 0x1a0000e, 0x1200020` in
   entrambi i casi, quindi la finestra resta dietro. Nessun click.

4. `ActionInvoked` con l'id vero:

```
gdbus emit --session --object-path /org/freedesktop/Notifications \
  --signal org.freedesktop.Notifications.ActionInvoked "uint32 395" "default"
```

   Prima: `_NET_CLIENT_LIST_STACKING = 0x1c0000f, 0x1a0000e, 0x1200020`
   (finestra davanti: `xmessage`). Dopo:
   `0x1c0000f, 0x1200020, 0x1a0000e` e
   `_NET_ACTIVE_WINDOW = 0x1a0000e`. La finestra passa sopra e prende il focus
   (`WM_STATE: Normal`). Screenshot: `/tmp/opencode/notify-proof/prima-del-click.png`
   e `/tmp/opencode/notify-proof/dopo-il-click.png`.

5. Secondo `ActionInvoked` con lo stesso id 395: impilamento invariato, la
   finestra resta dietro.

6. `ActionInvoked` con 999999 di nuovo: impilamento invariato.

Senza tray: una sessione privata con `dbus-run-session`, un servizio di
notifiche finto e il mittente vero (`internal/notify`), senza applicazione e
senza tray:

```
server: Notify app="Docker Controller" actions=["default" ""] title="Docker Controller" body="prova senza tray" timeout=-1
ok: an id never sent reports no click
ok: the click on the notification of this application arrived
```

Il servizio finto risponde con l'id 4242 e manda `ActionInvoked(4242,
"default")`; la callback scatta. Un `ActionInvoked(99, "default")` prima della
notifica non fa scattare niente.

### Casi limite

- `gdbus emit` senza tipo manda un `int32` per il numero, mentre l'interfaccia
  dichiara `uint32`. Per questo le prove usano `"uint32 395"` come argomento:
  il comando del ticket con il numero nudo non supera il controllo di tipo di
  `handle`. Il servizio vero manda sempre `uint32`.
- Un id resta in `pending` anche quando l'utente chiude la notifica senza
  cliccare. Il ticket chiede solo `ActionInvoked`, quindi `NotificationClosed`
  non si ascolta e il set cresce di un `uint32` per ogni notifica non
  cliccata, fino al tetto di 64: da lì esce l'id più vecchio, che perde il
  click. Una notifica parte solo alla fine di un'Attività, quindi la crescita
  è lenta; togliere l'id su `NotificationClosed` sarebbe una decisione nuova.
- Un click arrivato tra il ritorno di `Notify` e `pending.add` andrebbe perso.
  L'ascolto parte prima della chiamata, quindi il segnale è già in coda, ma
  l'id non è ancora noto. In pratica il click arriva dopo che l'utente vede la
  notifica.
- `ActionInvoked` si accetta da qualunque client del bus di sessione che conosca
  l'id: non c'è un controllo sul mittente. Un altro processo può alzare la
  finestra. È lo stesso modello di fiducia del servizio di notifiche.
- Su Wayland nativo, senza un token di attivazione vero, KWin può rifiutare
  `ActivateWindow` (lo ha misurato il ticket 11). La prova qui è su xcb.
