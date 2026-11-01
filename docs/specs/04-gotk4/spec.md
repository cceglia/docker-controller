# GUI GTK4 con gotk4

**Status:** superseded by `05-miqt`

## Problem Statement

L'app oggi è Wails v3 più una webview WebKit più un frontend Svelte. Ogni istanza occupa circa 185 MB di PSS in tre processi, misurati sul binario installato. La build richiede Node, Vite, il CLI `wails3` e la rigenerazione dei binding TypeScript. La vista è 3.659 righe di componenti Svelte più 227 righe di CSS, e vive in un altro linguaggio rispetto al backend.

Il backend non dipende dal framework. I package che parlano con Docker, con il database e con le impostazioni sono 2.544 righe di Go con 3.309 righe di test. Con la tray, il fake usato dai test e la shell il totale arriva a 4.009 righe di Go e 3.898 di test. Comunicano con Docker solo attraverso la CLI.

## Solution

Sostituire Wails, WebKit e Svelte con una GUI GTK4 scritta in Go: `gotk4` per i widget, `libadwaita` per il tema e i pattern di GNOME, `fyne.io/systray` per la tray.

Il backend resta intatto. La vista si riscrive come package di GUI. La shell diventa una `AdwApplication` con una finestra sola, il tray si attacca al loop GTK con `systray.Register()` e le notifiche passano a D-Bus. Alla fine del lavoro escono dal repository il frontend, Node, Vite, il CLI `wails3` e le dipendenze Wails.

L'app resta su Linux, usa il tema Adwaita e i font di sistema, e mantiene il comportamento attuale: stati dei Progetti, Azioni, Attività, Log, Utilizzo risorse, tray con menu annidato, chiusura verso la tray e notifiche.

## User Stories

### Finestra e navigazione

1. Come utente, voglio una sola finestra con la lista dei Progetti a sinistra e il dettaglio a destra, così da riconoscere l'app di prima.
2. Come utente, voglio il pannello Attività e Log in basso nel dettaglio, così da leggere l'output senza cambiare schermata.
3. Come utente, voglio che sotto i 900 px la finestra mostri una pagina alla volta con un pulsante indietro, così da usarla anche stretta.
4. Come utente, voglio che la finestra si apra a 1120x760, così da avere lo stesso spazio di oggi.
5. Come utente, voglio che il titolo della finestra sia "Docker Controller", così da riconoscerla nella barra delle applicazioni.
6. Come utente, voglio che la finestra si riapra dal tray con un click sull'icona, così come succede ora.

### Progetti e Servizi

7. Come utente, voglio vedere i Progetti raggruppati per Gruppo, con il Nome visualizzato quando esiste, così da ritrovare la stessa lista.
8. Come utente, voglio il colpo d'occhio sullo stato dei Container, con le tacche dei Container accanto al nome, così da capire "3 su 5" senza aprire il dettaglio.
9. Come utente, voglio le Azioni Up, Stop, Restart, Pull, Build, Down e Down con volumi sul Progetto, e Start, Stop, Restart e Down sul Servizio, così da non perdere nessuna azione.
10. Come utente, voglio le stesse conferme sulle Azioni distruttive, così da non cancellare dati per sbaglio.
11. Come utente, voglio modificare Nome visualizzato, Gruppo e Profili attivi dalla finestra, così da non dover passare dal menu del Progetto.
12. Come utente, voglio aggiungere un Progetto scegliendo i File compose dal selettore di file, rimuoverlo dalla lista, aprire la cartella, il terminale e la porta nel browser, così da fare tutto dalla UI.

### Attività

13. Come utente, voglio vedere le Attività in corso e concluse, con output e esito, così da sapere cosa sta succedendo.
14. Come utente, voglio che ogni Progetto abbia al massimo un'Attività alla volta, così come oggi.
15. Come utente, voglio che un'Attività fallita mostri il messaggio d'errore in un toast, così da accorgermene subito.
16. Come utente, voglio che l'output dell'Attività si aggiorni mentre il comando gira, così da seguirlo in tempo reale.

### Log

17. Come utente, voglio leggere i log di un Progetto o di un Servizio, con le righe colorate per Servizio e gli errori distinti, così da orientarmi.
18. Come utente, voglio cercare dentro i log, mettere in pausa lo scorrimento e mostrare o nascondere i timestamp, così come oggi.
19. Come utente, voglio che il pannello segua le righe nuove finché non lo metto in pausa, così da vedere l'ultima riga.
20. Come utente, voglio selezionare e copiare il testo dei log, così da incollarlo altrove.
21. Come utente, voglio che il pannello tenga al massimo le ultime 5.000 righe, così da non far crescere la memoria.
22. Come utente, voglio che chiudere il pannello termini il processo `docker compose logs`, così da non lasciare processi appesi.

### Utilizzo risorse

23. Come utente, voglio la barra globale con CPU, memoria, rete e disco della macchina, così da vedere il carico a colpo d'occhio.
24. Come utente, voglio l'utilizzo di ogni Servizio accanto ai suoi Container, con meter e valori, così da trovare il Servizio pesante.
25. Come utente, voglio che le cifre si fermino quando la finestra si nasconde nella tray, così da non mostrare numeri vecchi.
26. Come utente, voglio che una macchina non raggiungibile mostri un avviso invece di numeri falsi, così da capire cosa manca.

### Tray

27. Come utente, voglio l'icona nella tray con il menu per Gruppo, Progetto e Servizio, così da comandare tutto senza aprire la finestra.
28. Come utente, voglio che l'icona cambi quando Docker non risponde, così da vedere lo stato guasto.
29. Come utente, voglio che il click sinistro riporti la finestra in primo piano e che il tasto destro apra il menu, così come oggi su KDE.
30. Come utente, voglio che il menu si aggiorni a ogni snapshot, così da vedere lo stato corrente.
31. Come utente, voglio che le Azioni della tray funzionino, comprese le conferme di Down, così da non perdere niente.
32. Come utente, voglio l'hover sull'icona che dice "Docker Controller", così da riconoscerla.
33. Come utente, voglio le voci "Mostra finestra" ed "Esci" in fondo al menu, così da chiudere l'app davvero.

### Notifiche

34. Come utente, voglio una notifica desktop alla fine di ogni Attività lanciata dalla tray e a ogni Attività fallita, così da sapere com'è andata.
35. Come utente, voglio lo stesso titolo "Docker Controller" e lo stesso testo nelle notifiche, così da riconoscerle.
36. Come utente, voglio che una notifica non compaia due volte per la stessa Attività, così come oggi.

### Impostazioni e chiusura

37. Come utente, voglio scegliere cosa fa la chiusura della finestra: nascondi nella tray, chiedi ogni volta, esci.
38. Come utente, voglio il dialogo "Chiedi ogni volta" con l'opzione di ricordare la scelta, così come oggi.
39. Come utente, voglio impostare il comando del terminale e vederlo validato, così da aprire il terminale che uso.
40. Come utente, voglio che "Esci" dal tray chiuda davvero l'app e fermi gli stream, così da non lasciare processi docker appesi.
41. Come utente, voglio che le impostazioni salvate restino nel database esistente, così da non riconfigurare nulla.

### Build e manutenzione

42. Come manutentore, voglio che il backend di compose, database e impostazioni resti invariato, così da non rimettere in discussione il comportamento già coperto dai test.
43. Come manutentore, voglio che `make test` esegua gli stessi test Go di oggi, così da avere la stessa rete di sicurezza.
44. Come manutentore, voglio una build senza Node, npm e `wails3`, così da avere solo Go e le librerie di sistema.
45. Come manutentore, voglio che `make install` installi lo stesso binario, la stessa voce di menu e le stesse icone, così da non cambiare il modo di installare l'app.
46. Come manutentore, voglio i prerequisiti di build aggiornati nella documentazione, così che una macchina nuova si prepari in un passo.
47. Come manutentore, voglio che il frontend Svelte sparisca dal repository alla fine del port, così da non mantenere due viste.
48. Come manutentore, voglio un ADR che spieghi la scelta di gotk4, della libreria della tray e del vincolo AGPL, così che nessuno le rimetta in discussione senza contesto.

## Implementation Decisions

- **Framework**: `gotk4` `v0.4.1`, branch `4` (baseline GTK 4.14), con `CGO_ENABLED=1`. Sul computer c'è GTK 4.22.4: la baseline 4.14 è compatibile ABI e copre tutti i widget usati. Fra i prerequisiti entra `libgtk-4-dev`.
- **Tema e widget di GNOME**: `gotk4-adwaita`, ramo `adw-1.9`, con `libadwaita` 1.9.1 già presente come runtime. Fornisce `AdwApplication`, `AdwHeaderBar`, `AdwNavigationSplitView`, `AdwToastOverlay`, `AdwMessageDialog` e `AdwPreferencesWindow`.
- **App ID**: `io.github.cceglia.DockerController`, dal remote `cceglia/docker-controller`. La voce desktop prende lo stesso nome, perché GApplication associa nome, icona e notifiche tramite l'ID.
- **Struttura della finestra**: `AdwApplicationWindow` con `AdwNavigationSplitView`. A sinistra la lista Progetti in un `GtkListBox`, a destra il dettaglio in uno `GtkScrolledWindow`. Il pannello inferiore è uno `GtkPaned` verticale con `GtkStackSwitcher` e `GtkStack` per Attività e Log. Il collasso non è automatico: l'app aggiunge un `AdwBreakpoint` con condizione a 900 sp, cioè 900 px logici, che imposta `collapsed`, e alla selezione di un Progetto mostra la pagina di destra.
- **Moduli**: nuovo package per la GUI, con shell, viste e view model. Il `main` resta il cablaggio: database, impostazioni, `compose.Manager`, GUI e tray. Gli adapter Wails dei servizi spariscono; la loro logica non di presentazione si sposta nel package della GUI con i test. Il package della tray resta puro, senza dipendenze da GTK e dal framework.
- **Nessun livello eventi**: la UI legge lo snapshot del Manager e riceve le callback di Attività, Log e Utilizzo direttamente. Spariscono gli eventi, i payload JSON e i binding TypeScript.
- **Chiusura verso la tray**: alla chiusura la finestra si distrugge e si ricostruisce dallo snapshot alla riapertura, come oggi, così la cronologia delle Attività resta perduta e il testo del dialogo resta vero. Prima di distruggerla l'app ferma Utilizzo risorse e ogni stream di log. Un'app nascosta non deve lasciare processi `docker compose logs` aperti.
- **Thread**: ogni chiamata GTK avviene sul loop principale. Le callback che arrivano da goroutine (snapshot, log, `docker stats`, click della tray da D-Bus) si accodano con `glib.IdleAdd`.
- **Tray**: `fyne.io/systray` (Apache-2.0), avviata con `Register()` dentro il loop GTK. Su Linux usa D-Bus StatusNotifierItem via `godbus`, senza cgo e senza un secondo loop. L'adapter mappa la descrizione del menu della tray su voci di `systray`, ricostruisce il menu a ogni snapshot, usa il PNG dell'icona già generato e `SetOnTapped` per il click sinistro.
- **Notifiche**: nuovo package interno con un `Sender` e un trasporto D-Bus su `org.freedesktop.Notifications`. Il `Notifier` esistente continua a decidere cosa notificare, senza modifiche. Niente `GNotification`, che richiederebbe l'ID applicazione nel desktop file senza aggiungere nulla.
- **Elementi di dati**: i meter diventano `GtkLevelBar`, le pill di stato etichette con classi CSS di Adwaita. Le tacche dei Container restano un `GtkDrawingArea` disegnato con cairo, quaranta righe circa, perché sono informazione a colpo d'occhio.
- **Log**: `GtkTextView` in sola lettura con `GtkTextBuffer` e tag per Servizio e stderr, tetto di 5.000 righe, ricerca, pausa e timestamp. La selezione e la copia arrivano dal widget.
- **Azioni dalla tray**: il click arriva su una goroutine D-Bus; la conferma di Down si apre come `AdwMessageDialog` sul loop principale. Le tre chiamate del Notifier restano invariate.
- **Aperture esterne**: cartella, terminale e porta usano ancora `xdg-open` e il comando del terminale, con la stessa validazione di oggi.
- **Errori utente**: la mappatura attuale si sposta nel package della GUI e continua a produrre i testi italiani mostrati nei toast e nei dialoghi.
- **Build**: gli script di build perdono il CLI `wails3`, `npm` e il passo frontend. Restano `go vet`, `go test`, build e install. La verifica non rigenera più binding.
- **Prerequisiti**: `build-essential`, `pkg-config`, `libgtk-4-dev`, `libadwaita-1-dev`. Spariscono Node, npm, `wails3` e le librerie WebKit.
- **Font**: nessun bundle di Geist. Adwaita usa i font di sistema; il log usa il mono di sistema. Nessun cgo per fontconfig.

## Testing Decisions

- Un buon test verifica il comportamento osservabile, non i dettagli implementativi. Qui il comportamento osservabile è ciò che il backend espone a Docker e ciò che la vista calcola dai dati, non la forma dei widget.
- Le 3.653 righe di test di compose, database, impostazioni e tray non si toccano. Le 245 righe dei test dei servizi nel package main si spostano nel package della GUI con la stessa copertura: la macchina degli stream di log e la mappatura degli errori.
- Nuove funzioni pure con test da tavolo: formattazione di byte, percentuali, core e rate, stato di un Container, mappatura della descrizione del menu della tray sulle voci di `systray`, costruzione dei tag del log. Sono le stesse regole che oggi vivono in TypeScript, quindi i casi si portano uno a uno.
- Prior art: i test da tavolo di parsing e statistica nel package di compose, e i test del menu della tray. Stesso stile, stesse fixture dove esistono.
- Nessun test automatico dell'interfaccia: GTK4 non gira headless su questa macchina e un test che apre una finestra verifica poco.
- La seam più alta resta la pipeline di verifica: `go vet`, `go test` e la build con cgo, senza Wails e senza Node. È una seam sola per tutto il port.
- Checklist manuale su KDE prima di chiudere il branch: lista e raggruppamento, tacche dei Container, ogni Azione, conferma di Down, aggiunta e rimozione di un Progetto, apertura di cartella, terminale e porta, log con ricerca e pausa, uso della CPU, tray con click sinistro e menu, notifiche, tre comportamenti di chiusura, installazione con `make install`.

## Out of Scope

- Windows e macOS.
- Riportare i token CSS del frontend o il font Geist.
- Nuove funzioni non presenti oggi: temi personalizzati, notifiche interattive, scorciatoie da tastiera non esistenti.
- Test automatici dell'interfaccia.
- Modifiche al contratto con la CLI `docker compose` e ai package di compose, database e impostazioni.
- Riscrivere la logica del menu della tray e delle notifiche.

## Further Notes

- **AGPL**: `gotk4-adwaita` è AGPL-3.0. Il codice generato che l'app importa da `gotk4/pkg` è MPL-2.0, e l'AGPL di `gotk4` copre solo il generatore `gir`, che l'app non usa. Il vincolo nasce quindi da `gotk4-adwaita`: l'app resta personale, ma un binario distribuito diventa opera combinata AGPL. L'ADR 0004 lo registra.
- **Memoria attesa**: un probe GTK4 con lo stesso carico (30 Progetti, 20 Servizi, 5.000 righe di log) misura 68 MB di PSS con Python e circa 58-60 MB in una build Rust. Con `gotk4` ci si aspetta 60-70 MB, contro i 185 MB attuali.
- **Modalità a pagine**: sotto i 900 px la finestra mostra una pagina sola con il pulsante indietro. Oggi la finestra resta a due colonne e stringe la colonna laterale. Il comportamento nuovo arriva con `AdwNavigationSplitView` e con il suo `AdwBreakpoint`: è una scelta del port, non una parità.
- **Perché una libreria per la tray**: GTK4 non ha un'icona di stato. `fyne.io/systray` è l'unica opzione pura Go che non porta un loop GTK proprio: con Wails era esclusa per quel motivo (ADR 0001), con `gotk4` `Register()` è pensato per questo caso.
- **Ordine di lavoro suggerito**: shell e finestra, lista e dettaglio, Azioni, Attività, Log, Utilizzo, tray, notifiche, impostazioni e chiusura, packaging, rimozione del frontend. Ogni passo lascia l'app compilabile finché la build non passa sulla nuova shell; da lì in poi il branch non torna a Wails.
