# GUI Qt 6 con miqt

**Status:** ready-for-agent

**Supersedes:** `04-gotk4`

## Problem Statement

L'app oggi usa GTK4 con `gotk4` e `libadwaita`. Sul desktop KDE la finestra non
sembra parte del sistema: il tema Adwaita e i pattern di GNOME stanno accanto a
Breeze senza accordarsi. La build chiede `libgtk-4-dev`, `libadwaita-1-dev` e
`gobject-introspection`, quindi una macchina nuova deve conoscere la toolchain
GTK. `gotk4-adwaita` è AGPL-3.0: finché l'app resta personale non cambia nulla,
ma un binario distribuito diventa opera combinata AGPL.

La tray usa una seconda libreria, `fyne.io/systray`, che parla solo
StatusNotifierItem su D-Bus. Così l'app ha due toolkit per una finestra e
un'icona, e due vocabolari di widget da mantenere.

I test oggi misurano il comportamento del backend, ma nessuno misura la qualità
delle asserzioni. Un test che passa sempre non si distingue da un test utile.

## Solution

Sostituire `gotk4`, `libadwaita` e `fyne.io/systray` con miqt: Qt 6 Widgets
scritto in Go, con licenza MIT. Una sola libreria copre finestra, dialoghi, menu
e tray. L'aspetto è quello nativo di KDE, con Breeze come stile e con QSS solo
dove serve un ruolo semantico, come i colori dei Servizi nel log.

Il backend resta intatto: compose, database e impostazioni non cambiano
contratto. La vista si riscrive come package Qt. La tray passa a
`QSystemTrayIcon` e la chiusura nasconde la finestra invece di distruggerla.
Le notifiche restano su D-Bus, con l'azione `default` e l'ascolto di
`ActionInvoked`.

Alla fine del lavoro escono dal repository `gotk4`, `libadwaita` e
`fyne.io/systray`. Entrano miqt e Qt 6. Il mittente `internal/notify` resta,
perché è la via che riporta il click sulla notifica.
Entra anche la mutation testing con gremlins sui test unitari, con un target
`make mutate` e un file di configurazione per package sotto `.gremlins/`,
versionato.

## User Stories

### Finestra e navigazione

1. Come utente, voglio una sola finestra con la lista dei Progetti a sinistra e il dettaglio a destra, così da riconoscere l'app di prima.
2. Come utente, voglio il pannello Attività e Log in basso nel dettaglio, così da leggere l'output senza cambiare schermata.
3. Come utente, voglio che sotto i 900 px la finestra mostri una pagina alla volta con un pulsante indietro, così da usarla anche stretta.
4. Come utente, voglio che la finestra si apra a 1120x760, così da avere lo stesso spazio di oggi.
5. Come utente, voglio che il titolo della finestra sia "Docker Controller", così da riconoscerla nella barra delle applicazioni.
6. Come utente, voglio che la finestra torni dal tray con un click sull'icona, così come succede ora.
7. Come utente, voglio che l'app segua lo stile del desktop, così da non sembrare estranea a KDE.
8. Come utente, voglio che la finestra conservi il Progetto aperto e la cronologia delle Attività quando la nascondo nel tray, così da ritrovare lo stato al ritorno.

### Progetti e Servizi

9. Come utente, voglio vedere i Progetti raggruppati per Gruppo, con il Nome visualizzato quando esiste, così da ritrovare la stessa lista.
10. Come utente, voglio il colpo d'occhio sullo stato dei Container, con le tacche accanto al nome, così da capire "3 su 5" senza aprire il dettaglio.
11. Come utente, voglio le Azioni Up, Stop, Restart, Pull, Build, Down e Down con volumi sul Progetto, e Start, Stop, Restart e Down di un Servizio, così da non perdere nessuna azione.
12. Come utente, voglio le stesse conferme sulle Azioni distruttive, così da non cancellare dati per sbaglio.
13. Come utente, voglio modificare Nome visualizzato, Gruppo e Profili attivi dalla finestra, così da non dover passare dal menu del Progetto.
14. Come utente, voglio aggiungere un Progetto scegliendo i File compose dal selettore di file, rimuoverlo dalla lista, aprire la cartella, il terminale e la porta nel browser, così da fare tutto dalla UI.

### Attività

15. Come utente, voglio vedere le Attività in corso e concluse, con output e esito, così da sapere cosa sta succedendo.
16. Come utente, voglio che ogni Progetto abbia al massimo un'Attività alla volta, così come oggi.
17. Come utente, voglio che un'Attività fallita mostri il messaggio d'errore in un avviso in sovrimpressione, così da accorgermene subito.
18. Come utente, voglio che l'output dell'Attività si aggiorni mentre il comando gira, così da seguirlo in tempo reale.

### Log

19. Come utente, voglio leggere i log di un Progetto o di un Servizio, con le righe colorate per Servizio e gli errori distinti, così da orientarmi.
20. Come utente, voglio cercare dentro i log, mettere in pausa lo scorrimento e mostrare o nascondere i timestamp, così come oggi.
21. Come utente, voglio che il pannello segua le righe nuove finché non lo metto in pausa, così da vedere l'ultima riga.
22. Come utente, voglio selezionare e copiare il testo dei log, così da incollarlo altrove.
23. Come utente, voglio che il pannello tenga al massimo le ultime 5.000 righe, così da non far crescere la memoria.
24. Come utente, voglio che chiudere il pannello termini il processo `docker compose logs`, così da non lasciare processi appesi.

### Utilizzo risorse

25. Come utente, voglio la barra globale con CPU, memoria, rete e disco della macchina, così da vedere il carico a colpo d'occhio.
26. Come utente, voglio l'utilizzo di ogni Servizio accanto ai suoi Container, con un indicatore e i valori, così da trovare il Servizio pesante.
27. Come utente, voglio che le cifre dell'Utilizzo risorse si fermino quando la finestra si nasconde nella tray e ripartano al ritorno, così da non mostrare numeri vecchi.
28. Come utente, voglio che una macchina non raggiungibile mostri un avviso invece di numeri falsi, così da capire cosa manca.

### Tray

29. Come utente, voglio l'icona nella tray con il menu per Gruppo, Progetto e Servizio, così da comandare tutto senza aprire la finestra.
30. Come utente, voglio che l'icona cambi quando Docker non risponde, così da vedere lo stato guasto.
31. Come utente, voglio che il click sinistro riporti la finestra in primo piano e che il tasto destro apra il menu, così come oggi su KDE.
32. Come utente, voglio che il menu si aggiorni a ogni snapshot, così da vedere lo stato corrente.
33. Come utente, voglio che le Azioni della tray funzionino, comprese le conferme di Down, così da non perdere niente.
34. Come utente, voglio l'hover sull'icona che dice "Docker Controller", così da riconoscerla.
35. Come utente, voglio le voci "Mostra finestra" ed "Esci" in fondo al menu, così da chiudere l'app davvero.

### Notifiche

36. Come utente, voglio una notifica desktop alla fine di ogni Attività lanciata dalla tray e a ogni Attività fallita, così da sapere com'è andata.
37. Come utente, voglio lo stesso titolo "Docker Controller" e lo stesso testo nelle notifiche, così da riconoscerle.
38. Come utente, voglio che una notifica non compaia due volte per la stessa Attività, così come oggi.
39. Come utente, voglio un click sulla notifica che riporti la finestra in primo piano, così da arrivare al Progetto.

### Impostazioni e chiusura

40. Come utente, voglio scegliere cosa fa la chiusura della finestra: nascondi nella tray, chiedi ogni volta, esci.
41. Come utente, voglio il dialogo "Chiedi ogni volta" con l'opzione di ricordare la scelta, così come oggi.
42. Come utente, voglio impostare il comando del terminale e vederlo validato, così da aprire il terminale che uso.
43. Come utente, voglio che "Esci" dal tray chiuda davvero l'app e fermi gli stream, così da non lasciare processi docker appesi.
44. Come utente, voglio che le impostazioni salvate restino nel database esistente, così da non riconfigurare nulla.
45. Come utente senza tray, voglio che la chiusura esca dall'app e che le Impostazioni restino nella finestra, così da non restare bloccato.

### Build e manutenzione

46. Come manutentore, voglio che il backend di compose, database e impostazioni resti invariato, così da non rimettere in discussione il comportamento già coperto dai test.
47. Come manutentore, voglio che `make test` esegua ancora tutti i test di compose, database, impostazioni e tray, così da avere la stessa rete di sicurezza.
48. Come manutentore, voglio una build senza Node, senza npm e senza CLI di terze parti, così da avere solo Go, un compilatore C++ e le librerie di sistema.
49. Come manutentore, voglio che `make install` installi lo stesso binario, la stessa voce di menu e le stesse icone, così da non cambiare il modo di installare l'app.
50. Come manutentore, voglio i prerequisiti di build aggiornati nella documentazione, così che una macchina nuova si prepari in un passo.
51. Come manutentore, voglio che gli ultimi resti di GTK, Adwaita e systray spariscano dal repository, così da non mantenere due toolkit.
52. Come manutentore, voglio un ADR che spieghi la scelta di miqt, della tray e della licenza, così che nessuno la rimetta in discussione senza contesto.
53. Come manutentore, voglio che ogni modifica alla UI segua la skill locale di miqt, così che finestre, focus e QSS restino coerenti.
54. Come manutentore, voglio uno smoke test che costruisca la finestra senza display, così da scoprire subito un errore di costruttore o di cablaggio.
55. Come manutentore, voglio un test che controlli i token di design, così che la scala tipografica e i contrasti restino onesti.
56. Come manutentore, voglio la mutation testing sui test unitari, così da scoprire le asserzioni deboli.
57. Come manutentore, voglio che la mutation testing fallisca sotto una soglia fissa, così che una regressione blocchi il cambiamento.

## Implementation Decisions

- **Framework**: miqt `v0.15.0`, binding Go dei Qt 6 Widgets, con licenza MIT e `CGO_ENABLED=1`. Il repository importa il package `qt6` e il package `mainthread`. Qt 6.10.2 è la versione di questa macchina. La versione minima dichiarata è Qt 6.4, quella del README di miqt.
- **Un solo toolkit**: finestra, dialoghi, menu e tray usano Qt. `gotk4`, `gotk4-adwaita`, `libadwaita` e `fyne.io/systray` escono dal `go.mod`.
- **App ID**: resta `io.github.cceglia.DockerController`. L'applicazione lo dichiara come nome applicazione, perché Qt ricava da lì la classe della finestra, e la voce desktop dichiara la stessa classe in `StartupWMClass`. Senza questo allineamento KDE mostra l'icona sbagliata e non raggruppa la finestra sotto la voce di menu.
- **Finestra**: `QMainWindow` con `QSplitter` fra la lista a sinistra e il dettaglio a destra. Il pannello inferiore è un secondo `QSplitter` verticale con i due tab Attività e Log. Sotto i 900 px un `resizeEvent` porta la vista a un `QStackedWidget` con una pagina sola e il pulsante indietro. La finestra parte a 1120x760.
- **Ciclo di vita**: la chiusura nasconde la finestra, non la distrugge. `QApplication` gira con l'uscita sull'ultima finestra disattivata. La voce "Esci" della tray chiude l'applicazione e ferma gli stream. Senza tray la chiusura esce, e l'opzione "nascondi" non è disponibile: altrimenti l'utente non avrebbe modo di riaprire la finestra.
- **Moduli**: il package della GUI contiene shell, viste e view model. Il `main` resta il cablaggio: database, impostazioni, `compose.Manager`, GUI e tray. Il package della tray resta puro e senza Qt. Solo l'adapter della tray conosce `QSystemTrayIcon`.
- **Nessun livello eventi**: la UI legge lo snapshot del Manager e riceve le callback di Attività, Log e Utilizzo direttamente. Nessun payload JSON e nessun binding generato.
- **Thread**: ogni chiamata Qt avviene sul thread principale. Le callback che arrivano da goroutine si accodano con `mainthread.Start`. Il thread principale resta libero: nessuna operazione oltre 100 ms vi gira dentro. `mainthread.Wait` non si usa dentro uno slot.
- **Tray**: `QSystemTrayIcon` con `QMenu` aggiornato a ogni snapshot. Il menu si ricostruisce solo quando l'albero cambia. L'icona ha una variante per Docker non raggiungibile, e il tooltip dice "Docker Controller" più lo stato di Docker a parole. Il click sinistro mostra la finestra e il tasto destro apre il menu. L'avvio controlla `QSystemTrayIcon_IsSystemTrayAvailable`; senza tray l'app funziona lo stesso, la chiusura esce e le Impostazioni restano nella finestra. Su Wayland l'icona è uno StatusNotifierItem: KDE disegna il menu tramite `com.canonical.dbusmenu`, quindi il popup Qt non serve.
- **Notifiche**: D-Bus tramite `internal/notify` e `github.com/godbus/dbus/v5`, con titolo "Docker Controller" e il testo di oggi. `org.freedesktop.Notifications.Notify` porta l'azione `default`: `[]string{"default", ""}`. Il mittente ascolta `org.freedesktop.Notifications.ActionInvoked` per l'id della notifica inviata, e sul click mostra e alza la finestra. Le notifiche arrivano anche senza tray, perché passano dal bus di sessione. L'invio gira su una goroutine e non blocca il thread Qt. La decisione su cosa notificare resta nel package puro della tray, con le sue tre chiamate: `Launching`, `Aborted` e `Observe`. `internal/notify` non esce dal repository e `godbus` resta in `go.mod`.
- **Finestra nascosta e stream**: nascondere la finestra ferma il pannello Log e l'Utilizzo risorse, perché `docker compose logs` e `docker stats` sono processi esterni. L'output di un'Attività lanciata dalla tray continua, perché non dipende dalla finestra. Mostrare la finestra riavvia i due stream.
- **Tema**: stile nativo del desktop, Breeze su KDE. Niente foglio di stile globale che sostituisca il tema. Un QSS applicativo limitato definisce i ruoli che il tema non copre: i colori dei Servizi nel log, le etichette di stato e le tacche. I ruoli hanno un nome semantico e vivono in una sola costante Go. I valori vengono dalla palette del sistema dove la palette li espone, con un valore di ripiego dichiarato accanto. Il test dei token legge la stessa costante che il codice applica.
- **Tema chiaro e scuro**: i ruoli hanno due set, uno per la superficie chiara e uno per Breeze Dark, perché una tinta che passa su bianco di solito non passa su nero. Il set si sceglie dalla luminosità di `QPalette::Window`, con soglia 128. È il ripiego dichiarato quando la palette del sistema non espone il ruolo: la selezione e il pulsante principale usano ancora `palette(highlight)`, mentre un ruolo semantico assente dal sistema prende il valore del set, e il test dei token controlla ogni ruolo nei due set.
- **Azione suggerita**: la prima Azione che lo stato del Progetto suggerisce porta il nome oggetto `primaryAction`. Il QSS le dà `palette(highlight)` e il peso 600, e nessuna Azione ha un colore proprio. Lo stesso ruolo segna il comando che conferma il form di un nuovo Progetto e quello che salva le Impostazioni.
- **Riga selezionata e avviso**: la selezione della lista e l'avviso in sovrimpressione stanno nel QSS applicativo, perché il tema non li copre. La riga selezionata prende `palette(highlight)` e `palette(highlighted-text)`, anche senza fuoco. L'avviso è un `QFrame#toast` con `palette(base)`, il bordo `palette(mid)` e il raggio dei token.
- **Uscita dalla finestra**: il pulsante "Esci" non è nella finestra. L'uscita passa dalla voce "Esci" della tray e dal comportamento scelto per la chiusura; senza tray la chiusura esce dall'app, quindi una tray assente non lascia l'app senza uscita.
- **Tipi e spaziature**: scala tipografica con rapporto 1.125 e corpo a 16 px. Ritmo di spaziatura 4, 8, 12, 16, 24. Margini di pagina 16-24 px. Nessun testo utente in un controllo a larghezza fissa. Bersaglio minimo di 24x24 px per ogni controllo.
- **Accessibilità**: nome accessibile su ogni campo e su ogni lista. Ordine di tabulazione esplicito. Anello di focus visibile su ogni elemento interattivo. Lo stato non si distingue solo dal colore: c'è sempre un testo o una forma.
- **Elementi di dati**: gli indicatori di utilizzo sono `QProgressBar` con il valore numerico accanto. Le etichette di stato sono `QLabel` con un nome oggetto. Le tacche dei Container sono un `QWidget` con `paintEvent` dedicato, perché sono informazione a colpo d'occhio.
- **Log**: `QPlainTextEdit` in sola lettura, con formati di carattere per Servizio e per stderr, tetto di 5.000 righe, ricerca, pausa e timestamp. La selezione e la copia arrivano dal widget.
- **Azioni dalla tray**: il click arriva sul thread Qt. La conferma di un'Azione distruttiva si apre come dialogo modale. Le tre chiamate del Notifier restano invariate.
- **Aperture esterne**: cartella, terminale e porta usano ancora `xdg-open` e il comando del terminale, con la stessa validazione di oggi.
- **Errori utente**: la mappatura attuale resta nel package della GUI e continua a produrre i testi italiani degli avvisi e dei dialoghi.
- **Lingua**: stringhe italiane in Go, in una sola sorgente. Nessun `QTranslator` e nessun catalogo esterno, perché l'app ha una lingua sola.
- **Build**: gli script perdono ogni riferimento a GTK. Restano `go vet`, `go test`, build con cgo e install. Il primo build compila lo shim C++ di miqt e richiede 256,53 s con la cache Go vuota, circa quattro minuti. Con la cache calda dura 7,3 s. Anche `go test ./...` compila cgo, quindi chiede Qt e un compilatore C++.
- **Prerequisiti**: `build-essential`, `pkg-config`, `qt6-base-dev`. Spariscono `libgtk-4-dev`, `libadwaita-1-dev` e `gobject-introspection`.
- **Skill di UI**: ogni modifica alla vista carica la skill locale `miqt-ui-design` prima di toccare i widget. Le revisioni di layout e navigazione usano `qt-ui-design`. La regola vive in `AGENTS.md` e vale anche per i subagent.
- **Verifica della UI**: `make test-ui` esegue con `QT_QPA_PLATFORM=offscreen` lo smoke test della finestra e il test dei token di design. Il passaggio sugli screenshot resta manuale, con `QWidget.Grab()`, e le immagini non entrano nel repository.
- **Mutation testing**: gremlins `v0.6.0`, installato da `make tools`, configurato da un file per package in `.gremlins/` con `workers: 0` e `timeout-coefficient: 20`. Il target `make mutate` resta fuori da `make check`. Un primo ticket misura soli numeri, per ogni package, e li scrive nel ticket. Un secondo ticket fissa nei file `.gremlins/<package>.yaml` le soglie `--threshold-efficacy` e `--threshold-mcover` al valore misurato meno cinque punti, e fa fallire il target sotto soglia.
- **Ordine di lavoro**: una sola passata sul branch `miqt`. I primi tre ticket sono il prototipo su Qt 6.10.2, la shell Qt con la lista e l'uscita di GTK, e la verifica senza display. Da lì in poi ogni ticket riporta un'area e l'app resta compilabile. Gli ultimi tre coprono il packaging e i due giri di mutation testing.

## Testing Decisions

- Un buon test verifica il comportamento osservabile, non la forma dei widget. Qui il comportamento osservabile è ciò che il backend espone a Docker e ciò che la vista calcola dai dati.
- Le 3.653 righe di test di compose, database, impostazioni e tray restano identiche.
- I test della GUI si dividono. Le funzioni pure che oggi vivono nei file GTK si spostano in file senza Qt, con i loro test da tavolo: tacche, pannello Attività, comportamento di chiusura, preferenze, utilizzo, formattazione, stato di un Container, descrizione del menu della tray, tag del log, errori utente, macchina delle Attività e stream di log e di utilizzo. I test che costruiscono widget GTK spariscono: il loro comportamento passa allo smoke test del ticket 03 e alla checklist manuale.
- Da questo port `go test ./...` compila cgo, quindi richiede `qt6-base-dev` e un compilatore C++. `make test` resta lo stesso comando, con prerequisiti nuovi.
- I test nuovi seguono lo stile dei test da tavolo di compose e del menu della tray: casi in tabella, fixture esistenti, nomi che descrivono il comportamento.
- La mutation testing misura la qualità delle asserzioni sui package `internal/tray` e `internal/gui`, poi su `internal/compose`. Il primo giro misura e riporta, il secondo fissa le soglie nei file `.gremlins/<package>.yaml`. Restano esclusi `internal/db`, il fake di compose e i file generati.
- Il test dei token legge il foglio di stile e la scala tipografica, controlla il corpo minimo di 16 px e i rapporti di contrasto 4.5:1 e 3:1. Fallisce quando un ruolo nuovo salta il controllo.
- Lo smoke test costruisce la finestra con `QT_QPA_PLATFORM=offscreen` e la distrugge, senza mostrarla. Scopre errori di costruttore e di cablaggio senza display.
- La seam più alta resta la pipeline di verifica: `go vet`, `go test`, build con cgo, `make test-ui` e `make mutate`. Per il port è l'unica rete che copre tutto.
- Checklist manuale su KDE prima di chiudere il branch: lista e raggruppamento, tacche, ogni Azione, conferma di Down, aggiunta e rimozione di un Progetto, apertura di cartella, terminale e porta, log con ricerca e pausa, utilizzo della CPU, tray con click sinistro e menu, notifiche e click sulla notifica, tre comportamenti di chiusura, finestra nascosta e riaperta, scala al 125% e al 150%, installazione con `make install`.

## Out of Scope

- Windows e macOS.
- Una UI in QML.
- Temi personalizzati, dark mode disegnata a mano, notifiche con pulsanti oltre all'azione `default`.
- Catalogo di traduzioni e `QTranslator`.
- Un ripiego per la tray quando il desktop non offre uno StatusNotifierItem. Le notifiche restano fuori da questo caso, perché passano dal bus di sessione.
- Modifiche al contratto con la CLI `docker compose` e ai package di compose, database e impostazioni.
- Nuove funzioni di dominio non presenti oggi. Fanno eccezione la vista a pagine sotto i 900 px, la cronologia delle Attività conservata e gli strumenti di verifica dei test.
- Una pipeline di integrazione continua: il repository non ne ha una.
- La riscrittura della logica del menu della tray e della decisione su cosa notificare. Cambia solo il mittente D-Bus delle notifiche.

## Further Notes

- **Licenza**: miqt è MIT e Qt 6 è LGPL-3.0 con linking dinamico. Il vincolo AGPL di `gotk4-adwaita` sparisce. L'ADR 0005 lo registra.
- **Rischio tray**: chiuso dal prototipo. Su Wayland con KDE `QSystemTrayIcon_IsSystemTrayAvailable()` risponde `true`, `org.kde.StatusNotifierWatcher` è presente e l'oggetto `/StatusNotifierItem` è attivo. Il menu lo disegna il desktop tramite `com.canonical.dbusmenu`: il popup Qt non serve.
- **Rischio versione**: chiuso dal prototipo. miqt `v0.15.0` compila e gira contro Qt 6.10.2, con Go 1.27.3 e g++ 15.2.0. Il binding resta fissato a `v0.15.0`.
- **Rischio classe di finestra**: Qt ricava la classe della finestra dal nome applicazione. Se il nome non coincide con `StartupWMClass`, KDE mostra l'icona sbagliata. Il ticket 02 lo imposta, il ticket 14 lo controlla con un test.
- **Rischio build**: il primo build con cgo richiede 256,53 s con la cache Go vuota, circa quattro minuti, e una toolchain C++. Con la cache calda dura 7,3 s. Anche `go test ./...` compila cgo, quindi chiede Qt. Una macchina senza `qt6-base-dev` non esegue nemmeno i test.
- **Memoria misurata**: 27,0 MB di PSS a riposo su Wayland con la finestra aperta, 136,7 MB di RSS e 11 thread. Restano i 60-70 MB attesi da GTK4 e i 185 MB della versione Wails per il confronto.
- **Perché non QML**: la vista è fatta di liste, form e pannelli, quindi i Widgets coprono il caso senza un livello dichiarativo e senza il tooling QML.
- **Perché D-Bus per le notifiche**: `QSystemTrayIcon.ShowMessage` invia `Notify` con l'elenco `actions` vuoto, quindi KDE non può segnalare il click. D-Bus costa un trasporto da mantenere, ma il trasporto esiste già in `internal/notify`. In cambio il click sulla notifica funziona, e la notifica arriva anche senza tray.
- **ADR**: l'ADR 0005 sostituisce l'ADR 0003 per il ciclo di vita della finestra e l'ADR 0004 per il framework e la libreria della tray.

## Stato della verifica manuale

Verificato il 2 novembre 2026:

| Voce | Esito |
| --- | --- |
| Lista, raggruppamento, tacche, dettaglio, Servizi | Screenshot offscreen e finestra vera su Wayland e xcb |
| Azioni, conferme, aggiunta, modifica | Prova dell'utente su KDE e test offscreen |
| Log, Utilizzo risorse, Attività | Test offscreen e screenshot nei due temi |
| Tray: icona, tooltip, menu, "Esci", Azioni | Chiamate `Activate` ed `Event` del desktop su Wayland e xcb |
| Notifica e click | Notifica vera e `ActionInvoked` con l'id vero |
| Chiusura, nascondi, riapertura | Test offscreen |
| Scala al 125% e al 150% | `QT_SCALE_FACTOR` e screenshot |
| Installazione | `make install` in `~/.local`: binario, voce desktop valida, icone, `WM_CLASS` uguale a `StartupWMClass`, `kbuildsycoca6` |
| Mutation testing | Giro completo con le soglie, uscita 0 |

Restano da provare con le mani: i click fisici sulla tray, il giro completo con
la sola tastiera, i file nascosti nel selettore e la voce nel menu di KDE.
