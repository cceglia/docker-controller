# 02: Shell Qt, lista Progetti e uscita di GTK

**What to build:** L'app si avvia come applicazione Qt 6 Widgets, con la finestra 1120x760 e il titolo "Docker Controller". La colonna di sinistra mostra i Progetti salvati, raggruppati per Gruppo, con il Nome visualizzato quando esiste, il simbolo di stato e le tacche dei Container. La lista si aggiorna a ogni snapshot del Manager. Il dettaglio a destra per ora invita a scegliere un Progetto. I file di vista GTK escono dal repository qui, insieme alla loro logica pura: la build non conosce più GTK. Tray e notifiche mancano fino ai ticket 11 e 12, e la chiusura esce dall'app fino al ticket 13.

**Blocked by:** 01

**Status:** resolved

- [x] `make check` verde: `go vet`, `go test` e build con cgo contro Qt 6
- [x] La finestra si apre a 1120x760 con il titolo "Docker Controller"
- [x] La finestra usa `QMainWindow` con `QSplitter` fra lista e dettaglio
- [x] La lista mostra ogni Progetto salvato con il Nome visualizzato, o il nome compose quando manca
- [x] I Progetti con Gruppo compaiono sotto il loro Gruppo, quelli senza Gruppo in fondo nella sezione Senza Gruppo
- [x] Ogni riga mostra il simbolo di stato: Attivo, Parziale, Fermo, Rimosso
- [x] Le tacche mostrano fino a 12 Container e riassumono il resto in "+n"
- [x] Un Progetto aggiunto o rimosso compare senza riavviare l'app
- [x] La selezione di un Progetto apre il pannello di destra, anche vuoto
- [x] La colonna ha i pulsanti "Aggiungi progetto" e "Impostazioni", visibili anche con la lista vuota
- [x] Sotto i 900 px la finestra mostra una pagina alla volta con il pulsante indietro
- [x] Le funzioni pure che vivono nei file GTK si spostano in file senza Qt prima della cancellazione: tacche, pannello Attività, comportamento di chiusura, preferenze e utilizzo
- [x] I test di quelle funzioni restano verdi, e il ticket elenca quali test restano, quali si spostano e quali spariscono
- [x] I test di compose, database, impostazioni e tray restano identici
- [x] Nessun import di `gotk4`, `gotk4-adwaita` o `fyne.io/systray` resta nel repository
- [x] L'applicazione dichiara l'ID `io.github.cceglia.DockerController` come nome applicazione, e la classe di finestra coincide con `StartupWMClass`
- [x] Liste e campi hanno un nome accessibile e un ordine di tabulazione esplicito
- [x] Verifica manuale su KDE: la finestra segue lo stile Breeze del desktop
- [ ] L'icona nel menu è quella giusta: serve la voce desktop installata, quindi la prova passa al ticket 14

## Comments

Primo ticket di codice del branch `miqt`. L'app perde tray, notifiche e comportamento di chiusura fino ai ticket successivi: è accettato, perché GTK e Qt non convivono in un solo processo. La verifica senza display e i token di design arrivano nel ticket 03.

## Esito

Il ticket è implementato sul checkout del branch `miqt`. La verifica manuale su KDE è passata: la finestra si apre, segue lo stile Breeze e mostra la lista come descritto. L'icona nel menu resta da provare con la voce desktop installata, nel ticket 14.

### Come è fatta la shell

- `internal/gui` è un package Qt 6 Widgets. `qtapp.go` crea la `QApplication`, dichiara nome applicazione, `desktopFileName` e nome visualizzato, e avvia il loop.
- `qtshell.go` costruisce la `QMainWindow`: titolo "Docker Controller", 1120x760, minimo 560x420. Un `QSplitter` divide la colonna della lista dal pannello di destra. Sotto i 900 px un `QStackedWidget` mostra una pagina alla volta, con il pulsante "Indietro".
- `qtprojectlist.go` è la lista: una `QListWidget` con una riga per Progetto, le intestazioni dei Gruppi e la sezione "Senza Gruppo" in fondo. La riga mostra il simbolo di stato, il Nome visualizzato, il nome compose quando è diverso, il conteggio e le tacche.
- `qttack.go` disegna le tacche con un `paintEvent` e i colori della palette.
- `qtstyle.go` è l'unico foglio di stile: ruoli di stato, intestazioni e scala tipografica. Il resto è il tema del desktop.
- Il watch del Manager gira in una goroutine e ogni snapshot passa sul thread Qt con `mainthread.Start`. Il thread Qt non aspetta nulla.
- La chiusura della finestra esce dall'app: `QGuiApplication_SetQuitOnLastWindowClosed(true)`.

### Verifica eseguita

Xvfb `:77`, 1280x900, senza window manager, con `QT_QPA_PLATFORM=xcb` e una HOME temporanea.

- `xwininfo -root -tree`: `0x20000c "Docker Controller" ... 1120x760`, con la classe `("dockercontroller" "io.github.cceglia.DockerController")`.
- `xprop -id 0x20000c WM_CLASS`: `WM_CLASS(STRING) = "dockercontroller", "io.github.cceglia.DockerController"`. Il secondo campo, la classe, è uguale a `StartupWMClass`. `_NET_WM_NAME` è "Docker Controller" e `WM_NORMAL_HINTS` riporta `user specified size: 1120 by 760`.
- Con 7 Progetti veri scoperti da Docker: i Gruppi "Casa" e "Lavoro" compaiono in cima, "Senza Gruppo" in fondo, il Nome visualizzato "VAKT" mostra sotto il nome compose "vakt".
- Una modifica al database durante l'esecuzione sposta `sunlinkvpn` sotto "Casa" al refresh successivo, senza riavviare l'app.
- Con Docker non raggiungibile e nessun Progetto salvato la lista mostra "Nessun Progetto." e i due pulsanti restano visibili.
- A 800x620 la finestra mostra la lista da sola. Un click su una riga apre la pagina del dettaglio con "Indietro", e "Indietro" torna alla lista con la riga ancora selezionata. Un secondo click sulla stessa riga riapre il dettaglio.
- Un `WM_DELETE_WINDOW` chiude la finestra e termina il processo.
- Le tacche oltre 12 e lo stato Parziale sono stati provati con un grab offscreen di una riga da 14 Container: 12 tacche, 9 piene e 3 vuote, più "+2". Il test temporaneo è stato cancellato.

### Test che restano invariati

I 20 file di test di `internal/gui`, con 124 funzioni, sono gli stessi di prima: `actions_test.go`, `activitylog_test.go`, `addproject_test.go`, `bottompanel_test.go`, `close_test.go`, `detail_test.go`, `edit_test.go`, `format_test.go`, `logstreams_test.go`, `logview_test.go`, `notify_test.go`, `open_test.go`, `preferences_test.go`, `remove_test.go`, `sections_test.go`, `status_test.go`, `tacks_test.go`, `usagebar_test.go`, `usagestream_test.go`, `user_errors_test.go`.

Nessun test si sposta: le funzioni pure passano in file nuovi dello stesso package, con lo stesso nome, la stessa firma e la stessa semantica.

I test di `compose`, `db`, `settings` e `tray` non cambiano di una riga.

### Funzioni pure spostate

| Funzione o costante | Da | A |
| --- | --- | --- |
| `closeAction`, `closeHide`, `closeQuit`, `closeAsk` | `close.go` | `closebehavior.go` |
| `closeActionFor` | `close.go` | `closebehavior.go` |
| `askCloseHeading`, `askCloseBody`, `rememberCloseText`, `closeHideResponse`, `closeQuitResponse`, `closeCancelResponse` | `close.go` | `closebehavior.go` |
| `closeAnswer`, `askCloseAnswers`, `closeAnswerFor` | `close.go` | `closebehavior.go` |
| `closeToSave`, `saveCloseBehavior` | `close.go` | `closebehavior.go` |
| `preferencesTitle`, `saveSettingsText`, `settingsSavedText` | `preferences.go` | `preferencesmodel.go` |
| `closeBehaviorGroupTitle`, `closeBehaviorGroupHint`, `terminalGroupTitle`, `terminalGroupHint`, `terminalRowTitle` | `preferences.go` | `preferencesmodel.go` |
| `closeBehaviorChoice`, `closeBehaviorChoices` | `preferences.go` | `preferencesmodel.go` |
| `closeBehaviorIndex`, `closeBehaviorAt`, `preferencesSaveOutcome` | `preferences.go` | `preferencesmodel.go` |
| `waitingUsageText`, `unreachableUsageText`, `usageTitle` | `usagebar.go` | `usagestate.go` |
| `UsageState`, `UsageFigures`, `UsageWaiting`, `UsageNone`, `UsageStateOf` | `usagebar.go` | `usagestate.go` |
| `ServiceRuns` | `usagebar.go` | `usagestate.go` |
| `StripState`, `StripWaiting`, `StripUnreachable`, `StripFigures`, `StripStateOf` | `usagebar.go` | `usagestate.go` |
| `panelActivity` | `bottompanel.go` | `panelmodel.go` |
| `panelCloseText`, `activityPageTitle`, `logPageTitle` | `bottompanel.go` | `panelmodel.go` |
| `activityTooltip`, `emptyActivityText` | `activitypanel.go` | `activitytext.go` |
| `appID`, `windowTitle`, `windowWidth`, `windowHeight`, `breakpointWidth`, `emptyListText` | `gui.go` | `windowconsts.go` |
| `groupLabels` | `editsection.go` | `edit.go` |

Non si spostano, perché descrivono solo GTK: `preferencesWidth`, `preferencesHeight`, `panelHeight`, `historyWidth`, i metri di `usagebar.go`, `toastTimeout`, `showOutputText`, `dismissed`, `filterList`, `rgbaText`, i costruttori di widget e le funzioni che li usano. Il ticket che riapre quell'area decide la controparte Qt.

### Test che spariscono

- `internal/trayadapter/tree_test.go`: sparisce con `internal/trayadapter`, che il ticket cancella. Provava la costruzione dell'albero del menu a partire dallo snapshot, cioè il pezzo che il ticket 11 rifà su `QMenu`. `internal/tray` con i suoi test resta, perché è puro.

Nessun altro test sparisce: nessun file `_test.go` del package `gui` costruiva widget GTK.

### Comandi di validazione

| Comando | Esito |
| --- | --- |
| `gofmt -l .` | vuoto sui file versionati. Segnala solo tre file sotto `.agents/tmp/scaf/`, una copia di scaffold non tracciata e precedente al ticket. |
| `gofmt -l $(git ls-files '*.go')` | vuoto |
| `go vet ./...` | nessun messaggio |
| `go test ./...` | `ok` su `dockercontroller`, `compose`, `db`, `gui`, `settings`, `tray`; nessun test rosso |
| `CGO_ENABLED=1 go build -o bin/dockercontroller .` | binario prodotto |
| `make check` | verde: vet, test e build |

Il primo build con miqt in questo checkout è durato pochi secondi perché la cache Go della macchina era già calda per `miqt v0.15.0` dal prototipo del ticket 01.

### Pulsanti disabilitati

"Aggiungi progetto" e "Impostazioni" restano presenti e disabilitati. La scelta è questa: i ticket 09 e 13 costruiscono i dialoghi, quindi un pulsante attivo senza dialogo sarebbe un comando che non fa nulla. Disabilitato, il pulsante dice la verità e la posizione nella colonna è già quella definitiva. Non hanno un tooltip: il testo parlerebbe di ticket all'utente finale.

### Note per i ticket successivi

- Il pannello di destra è visibile da subito con l'invito a scegliere un Progetto. Sotto i 900 px la selezione apre la pagina del dettaglio. In largo il pannello non si nasconde, perché l'invito è la sua ragione di esistere finché il dettaglio non arriva.
- `internal/tray/menu.go` cita ancora `internal/trayadapter` in un commento: il ticket 11 lo aggiorna.
- L'intestazione del `Makefile` dice ancora "GTK4 via gotk4": è un commento, e il ticket 14 riscrive gli script.
- Le stringhe "GTK" e "Adwaita" nei commenti dei file puri restano dove descrivono il comportamento vecchio. Il ticket che riscrive quell'area le corregge.
