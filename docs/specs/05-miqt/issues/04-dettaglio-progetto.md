# 04: Dettaglio Progetto

**What to build:** Il pannello di destra mostra il Progetto selezionato: intestazione con nome e stato, avviso quando Docker non risponde, avviso per un File compose mancante, i dati del Progetto, i comandi e l'elenco dei Servizi con il loro stato. La riga delle Azioni è presente e disabilitata, con il motivo scritto accanto. Il contenuto si ricostruisce solo quando lo snapshot cambia, e lo scorrimento resta dov'era. Il tema usa i ruoli del foglio applicativo per le etichette di stato.

**Blocked by:** 02

**Status:** resolved

- [x] L'intestazione mostra nome del Progetto, stato e tacche dei Container
- [x] Docker non raggiungibile mostra un avviso con il testo di oggi, al posto dei numeri
- [x] Durante il caricamento, prima del primo snapshot, il dettaglio non dichiara Docker non raggiungibile
- [x] Un File compose mancante mostra l'avviso e il comando per rimuovere il Progetto dalla lista
- [x] I dati del Progetto mostrano Gruppo, Profili attivi e File compose
- [x] La riga delle Azioni è disabilitata quando il Progetto non ammette Azioni, con il motivo visibile
- [x] Ogni Servizio mostra nome, stato dei suoi Container e lo spazio dell'indicatore di utilizzo
- [x] Un Progetto senza Servizi mostra il testo di oggi
- [x] Un aggiornamento dello snapshot non muove lo scorrimento sotto gli occhi dell'utente
- [x] Il dettaglio si aggiorna da solo a ogni snapshot, senza perdere la selezione
- [x] Lo stato non si distingue solo dal colore: ogni etichetta di stato ha un testo
- [x] Test da tavolo per la scelta del testo di stato e per il motivo di un'Azione disabilitata
- [x] Verifica manuale su KDE: il pannello regge la scala al 125% e al 150%
- [x] Le righe della lista restano su una linea sola al 125%

## Comments

Le funzioni pure già coperte dai test restano: `ProjectName`, `DockerNotice`, `ActionDisabledReason`.

## Esito

Il pannello vive in `internal/gui/qtdetail.go`, con la regola di ricostruzione in
`internal/gui/detailstate.go`. La shell passa il Progetto scelto e lo aggiorna a
ogni snapshot: `shell.apply` cerca il Progetto selezionato e lo dà al pannello,
`shell.open` fa lo stesso sulla riga cliccata.

**Struttura del pannello**, dall'alto:

| Blocco | Contenuto |
| --- | --- |
| Intestazione | Nome visualizzato, nome compose quando è diverso, glifo e testo dello stato, conteggio e tacche. Con Docker non raggiungibile l'avviso prende il posto del conteggio e delle tacche |
| Dati del Progetto | Gruppo, Profili attivi, File compose. Le righe senza valore restano nascoste; il Gruppo senza valore dice "Gruppo: nessuno" |
| File compose mancante | L'avviso e il comando "Rimuovi dalla lista", presente e disabilitato |
| Azioni | La riga di `ProjectActions`, disabilitata, con il motivo di `actionReasonText` scritto accanto |
| Comandi | "Cartella", "Terminale", "Porta", presenti e disabilitati |
| Servizi | Il titolo con il numero, un blocco per Servizio con nome, conteggio dei Container, tacche, lo spazio riservato dell'indicatore di utilizzo e una riga per Container. Senza Servizi mostra `NoServicesText` |

Lo stato ha sempre un testo: il glifo dello stato e la parola ("Attivo",
"Parziale", "Fermo", "Rimosso"), il glifo del Container con lo stato e la salute
a parole. Il foglio di stile aggiunge il ruolo `statusFault` per il testo di un
Container guasto: il rosso delle tacche vale 3:1, troppo poco per un testo. Il
ruolo sta in `designtokens.go` con il suo tipo, e il test dei token lo controlla
nei due temi.

**Regola di ricostruzione**: `detailPanel.set` confronta i dati con
`detailData.sameDraw`. Il confronto legge solo i campi che il pannello disegna:
Nome, Nome visualizzato, Gruppo, File compose, File compose mancante, stato,
conteggi, Servizi, Profili attivi, presenza dell'Attività, Docker e caricamento.
Un cambio di `Folder` o dei Profili definiti non ricostruisce il pannello.

**Scorrimento**: il pannello legge il valore della barra prima di ricostruire e
lo restituisce quando arriva la nuova ampiezza, sul segnale `rangeChanged`. Un
timer non basta: scatta quando i blocchi vecchi sono già tolti e l'ampiezza vale
ancora zero.

**Test aggiunti**:

- `detailstate_test.go`: `TestDetailData_SameDrawComparesOnlyWhatThePanelDraws`,
  14 casi, dal `Folder` che non conta al Container che cambia stato.
- `actions_test.go`: `TestActionReasonText_SaysNothingBeforeTheFirstSnapshot`.
- `detail_test.go`: `TestServiceStateLine_CountsTheRunningContainersOfTheService`
  e `TestGroupLine_NamesTheGroupOrSaysThereIsNone`.
- `qtshell_test.go`: lo smoke test costruisce il pannello con un Progetto finto,
  controlla i blocchi, non ricostruisce con gli stessi dati, ricostruisce con dati
  diversi, tiene lo scorrimento a 200, mostra l'avviso del File compose mancante
  con il comando disabilitato, mette l'avviso di Docker al posto dei numeri solo
  dopo il primo snapshot, e controlla 15 controlli contro il bersaglio di 24 px.
  Il processo crea una sola `QApplication`.

**Comandi eseguiti**:

```
gofmt -l .
go vet ./...
make check
make test-ui
go test -race ./internal/gui
```

Tutti passano. `gofmt -l .` segnala solo i file di `.agents/tmp/scaf`, presenti
prima di questo ticket e fuori dal repository.

**Verifica manuale su KDE**: il pannello regge il 125% e il 150%, senza testo
tagliato e senza blocchi sovrapposti. Al 125% le righe della lista mostravano il
nome compose su una seconda linea. L'utente ha chiesto una riga sola, quindi la
riga mostra il solo Nome visualizzato e il nome compose passa nel tooltip e
nell'intestazione del dettaglio. `rowHeightTwoLines` sparisce, la lista torna a
`SetUniformItemSizes(true)` e la regola `composeName` esce dal foglio di stile.

**Difetti visti sugli screenshot e corretti**:

- La riga selezionata aveva testo e tacche scuri sul fondo blu della selezione.
  Ora il foglio di stile colora le etichette della riga selezionata con
  `palette(highlighted-text)`, e `tackStrip` porta lo stato di selezione in un
  campo: le tacche si dipingono con lo stesso colore.
- Il pulsante "Indietro" occupava tutta la larghezza della pagina stretta. Ora
  sta a sinistra con la sua larghezza.
- L'avviso di Docker non raggiungibile stava sulla riga dello stato e si
  leggeva come parte dello stato. Ora ha una riga sua, al posto dei numeri.

**Screenshot senza display**: lo smoke test scrive i PNG degli stati principali
quando `SHOT_DIR` è impostato. `QWidget.Grab()` rende la finestra con il backend
offscreen, quindi la revisione del layout non chiede un desktop.

```
SHOT_DIR=/tmp/opencode/shots QT_QPA_PLATFORM=offscreen \
  go test -tags ui -run TestShellOffscreen ./internal/gui
```

I file sono `01-empty`, `02-list`, `03-detail`, `04-detail-missing`,
`05-detail-unreachable`, `06-narrow-list`, `07-narrow-detail`. Restano fuori dal
repository.
