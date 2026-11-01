# 11: Tray Qt

**What to build:** L'icona di sistema passa a `QSystemTrayIcon`. Il menu ha le voci per Gruppo, Progetto e Servizio, si aggiorna a ogni snapshot, cambia icona quando Docker non risponde e mostra un tooltip che dice "Docker Controller" e lo stato di Docker. Il click sinistro mostra la finestra, il tasto destro apre il menu. In fondo ci sono "Mostra finestra" ed "Esci", e "Esci" chiude l'applicazione. Quando la tray non è disponibile l'app funziona lo stesso e la finestra offre "Esci" e le Impostazioni.

**Blocked by:** 05

**Status:** resolved

- [x] L'app mostra un'icona di sistema con `QSystemTrayIcon`
- [x] Il menu ha le voci per Gruppo, Progetto e Servizio, con le stesse gerarchie di oggi
- [x] Il menu si ricostruisce solo quando l'albero delle voci cambia
- [ ] Il click sinistro mostra la finestra e le porta il focus
- [ ] Il tasto destro apre il menu
- [x] Su Wayland il menu lo disegna il desktop tramite `com.canonical.dbusmenu`: il popup Qt non serve
- [x] L'icona cambia quando Docker non risponde
- [x] Il tooltip dice "Docker Controller" e ripete lo stato di Docker a parole
- [x] Le Azioni della tray funzionano, comprese le conferme di un'Azione distruttiva
- [x] "Mostra finestra" chiude il menu e mostra la finestra
- [x] "Esci" chiude l'applicazione con `QApplication_Quit`
- [x] Senza tray l'app parte lo stesso, e la finestra offre "Esci" e le Impostazioni
- [x] L'avvio controlla la disponibilità della tray prima di crearla
- [x] Test da tavolo per l'albero delle voci della tray e per il confronto fra due alberi
- [ ] Verifica manuale su KDE, su X11 e su Wayland: icona, menu, i due tasti e un menu aperto durante uno snapshot

## Comments

Il package della tray resta puro: la descrizione del menu non cambia. Cambia solo l'adapter che la applica. Il fermo degli stream all'uscita arriva nel ticket 13.

## Esito

### Struttura

`internal/tray` non cambia: `BuildMenu`, `IconFor` e i due PNG dell'icona
restano quelli. Il package della GUI prende due file.

- `internal/gui/traytree.go` è puro, senza Qt. Contiene `trayNode`, la mappa
  `trayTree([]tray.Item) []trayNode`, il confronto `sameTrayTree`, il tooltip
  `trayTooltip(compose.Snapshot)` e la domanda `trayQuestion(trayNode)`. Il
  confronto è la condizione che decide se ricostruire il menu.
- `internal/gui/qttray.go` è l'unico file che conosce `QSystemTrayIcon`. Contiene
  l'adapter (`newTrayAdapter`, `startTray`, `apply`, `rebuild`, `add`, `choose`)
  e le azioni dell'applicazione che l'adapter chiama: `present`, `quit`,
  `chooseTray`, `askTray`, `launchTray`.

`qtapp.go` crea la tray prima del primo snapshot e prima di mostrare la
finestra, e passa ogni snapshot a `a.tray.apply`. `qtshell.go` prende il metodo
`present` e il pulsante "Esci".

Il pulsante "Esci" sta in una riga sua sotto "Aggiungi progetto" e
"Impostazioni": tre etichette in una riga sola allargano la colonna oltre i
300 px e la lista perde spazio. Il pulsante entra nell'ordine di tabulazione
dopo "Impostazioni", e `tabEntry` restituisce lui.

### Icona

L'icona viene dal package puro: `tray.IconFor(snapshot)` sceglie la variante e
`tray.Icon.PNG()` dà i byte che `QPixmap.LoadFromDataWithData` carica.
`QIcon_FromTheme` non si usa, quindi l'icona c'è anche in una build non
installata.

Perché non `go:embed` di `build/linux/icons`: `go:embed` non legge fuori dal
package, quindi servirebbe una copia del PNG dentro `internal/gui` e una seconda
variante per Docker spento. I due PNG della tray sono già l'arte dell'app, già
versionati, già con la variante offline, e sono quelli che l'utente vedeva con
`systray`. Una copia in più sarebbe una seconda fonte di verità per la stessa
immagine. Il disegno con `QPainter` resta l'alternativa se un giorno l'arte
cambia forma, ma non serve per un PNG già pronto.

### Menu

`apply` aggiorna sempre tooltip e icona, e ricostruisce il menu solo quando
`sameTrayTree` dice che l'albero è cambiato. `rebuild` fa `Clear` sul
`QMenu` del contesto, cancella i sottomenu del giro prima (dal più profondo) e
riempie di nuovo. Il menu è lo stesso oggetto per tutta la vita dell'app, quindi
`SetContextMenu` gira una volta sola. Su Wayland il desktop chiede il layout a
`com.canonical.dbusmenu`: nessun popup Qt.

Il tooltip è `Docker Controller · <stato>` con lo stato "Controllo di
Docker…", "Docker risponde" o "Docker non risponde".

### Azioni e uscita

Un click su una voce arriva a `trayAdapter.choose` sul thread Qt. "Mostra
finestra" chiama `present`, cioè `ShowNormal`, `Raise` e `ActivateWindow`.
"Esci" chiama `QCoreApplication_Quit` (miqt espone `QApplication_Quit` come
`QCoreApplication_Quit`). Un'Azione con `Confirm` apre `newConfirmDialog` con la
`Confirmation` di `trayQuestion`; `Down con volumi` chiede anche il nome del
Progetto, anche se il menu della tray oggi non offre quella voce. L'Azione parte
su una goroutine, con `Notifier.launch` quando c'è il notifier, e un rifiuto
arriva all'avviso della finestra.

### Senza tray

`startTray` chiama `QSystemTrayIcon_IsSystemTrayAvailable` e restituisce `nil`
quando la risposta è falsa. `Run` va avanti lo stesso, la finestra ha "Esci" e
"Impostazioni". "Impostazioni" resta disabilitata: la finestra delle
Impostazioni arriva nel ticket 13.

### Test

- `internal/gui/traytree_test.go`: tabelle per `trayTree` (separatore, avviso,
  Azione con domanda, Command, sottomenu, tre livelli), per `sameTrayTree`
  (etichetta, stato, domanda, Command, Azione, Servizio, separatore, figlio
  mancante, operandi scambiati), per `trayTooltip` e per `trayQuestion` (Down di
  Progetto, Down di Servizio, Down con volumi). Un caso parte da
  `tray.BuildMenu` vero e controlla la gerarchia Gruppo → Progetto → Servizi →
  Azione. Un caso controlla che l'albero non condivida memoria con la
  descrizione.
- `internal/gui/qtshell_test.go`, passo `checkTray` dello smoke test `ui`:
  costruisce l'adapter con `newTrayAdapter`, senza mostrare l'icona. Controlla
  che `startTray` risponda `nil` con il backend offscreen, il tooltip, l'icona
  non nulla, le quattro voci in cima (Progetto, separatore, "Mostra finestra",
  "Esci"), i due `Trigger`, la scelta di una voce, che uno snapshot uguale non
  ricostruisca il menu e che uno diverso lo ricostruisca, cambi icona e svuoti i
  sottomenu.
- I test di `internal/tray` restano invariati e verdi.

### Comandi

```
gofmt -l .
go vet ./...
make check
make test-ui
SHOT_DIR=/tmp/opencode/shots QT_QPA_PLATFORM=offscreen go test -tags ui -run TestShellOffscreen ./internal/gui
```

### Prove

Wayland (KDE, Qt 6.10.2), app avviata da `bin/dockercontroller`:

- voce StatusNotifierItem nella lista di `org.kde.StatusNotifierWatcher`;
- icona bianca visibile nel pannello (la balena, prima del cubo di Docker);
- `Title` "Docker Controller", `ToolTip` "Docker Controller · Docker
  risponde";
- `GetLayout` su `/MenuBar` restituisce Gruppo, Progetto, Azioni, "Servizi",
  Servizio, "Mostra finestra" ed "Esci", con i separatori al loro posto;
- `Event("clicked")` su "Down" apre il dialogo "Docker Controller" con il testo
  "Fermare e rimuovere i Container di sunlink-shared? I volumi restano." e i due
  pulsanti "Down" e "Annulla" (screenshot
  `/tmp/opencode/shots/tray-confirm-wayland.png`);
- nessuna Azione è stata confermata: il dialogo è rimasto senza risposta e
  l'app è uscita da "Esci". La strada dal menu al `Manager.Do` è la stessa del
  pannello, già coperta dai suoi test;
- `Event("clicked")` su "Esci" chiude l'app con codice 0;
- con `DOCKER_HOST` su un socket inesistente il tooltip diventa "Docker
  Controller · Docker non risponde", il menu comincia con "Docker non
  raggiungibile" e l'icona diventa la variante grigia con il punto rosso
  (screenshot `/tmp/opencode/shots/panel-icons-offline.png`).

xcb (XWayland), app avviata con `QT_QPA_PLATFORM=xcb`:

- la tray c'è anche qui (Qt parla StatusNotifierItem via D-Bus), con lo stesso
  tooltip e lo stesso menu;
- `Activate` porta la finestra davanti (screenshot
  `/tmp/opencode/shots/window-xcb-front.png`), che mostra "Esci" e
  "Impostazioni";
- `Esci` chiude l'app con codice 0.

Senza tray: `dbus-run-session -- ./bin/dockercontroller` non trova il watcher,
quindi `startTray` restituisce `nil`. L'app parte lo stesso, la finestra esiste
(KWin la elenca) e il processo resta vivo. Lo smoke test offscreen controlla lo
stesso caso e la finestra con "Esci" e "Impostazioni" (screenshot
`/tmp/opencode/shots/02-list.png`).

### Casi limite

- Il click fisico del mouse e il menu aperto durante uno snapshot non sono
  riproducibili su questa macchina: non c'è `xdotool`, `ydotool` né `wtype`.
  Il percorso è stato provato con i metodi D-Bus che il desktop chiama
  (`Activate`, `Event("clicked")`), non con il puntatore.
- Su Wayland un `Activate` sintetico non riporta davanti una finestra ridotta a
  icona dal compositore: la chiamata arriva (`present` è stato visto girare con
  una stampa temporanea), ma senza il token di attivazione di un input vero KWin
  rifiuta. Su xcb la stessa chiamata porta la finestra davanti.
- "Impostazioni" resta disabilitata finché il ticket 13 non costruisce la
  finestra. Il pulsante c'è, in entrambi i casi.
- Ogni `OnTriggered` crea un handle cgo che miqt non libera quando l'azione
  viene distrutta. Le ricostruzioni sono rare (due all'avvio in una prova di 25
  secondi: una per lo stato di caricamento, una per il primo snapshot), ma un
  menu che cambia spesso lascia handle per il processo.
- Un `GetLayout` da un client D-Bus usa e getta (gdbus) fa comparire su stderr
  `DBusMenuExporterDBus::GetLayout ... Condition failed: menu`. La stessa riga
  compare con altri programmi Qt (VLC). Con il client del desktop il menu si
  legge e i sottomenu rispondono.
- Il primo snapshot va sull'icona prima di `Show`, altrimenti Qt stampa
  `QSystemTrayIcon::setVisible: No Icon set`.
