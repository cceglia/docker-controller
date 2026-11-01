# 07: Log

**What to build:** Il tab Log mostra i log di un Progetto o di un Servizio in un `QPlainTextEdit` in sola lettura, con le righe colorate per Servizio e gli errori distinti. Il pannello ha ricerca, pausa dello scorrimento, mostra o nascondi i timestamp, tiene al massimo 5.000 righe e permette di selezionare e copiare il testo. Chiudere il pannello termina il processo `docker compose logs`.

**Blocked by:** 06

**Status:** resolved

- [x] Il tab Log mostra i log del Progetto selezionato
- [x] Un Servizio del dettaglio apre i log del solo Servizio
- [x] Ogni Servizio ha un colore stabile, preso dalla palette dei ruoli
- [x] Le righe di stderr si distinguono da quelle di stdout
- [x] La ricerca evidenzia le righe che corrispondono e conta i risultati
- [x] La pausa ferma lo scorrimento e riprendere torna all'ultima riga
- [x] Il pulsante dei timestamp mostra e nasconde l'ora di ogni riga
- [x] Il pannello tiene al massimo 5.000 righe e scarta le più vecchie
- [x] Il testo si seleziona e si copia dal widget
- [x] Chiudere il pannello o cambiare Progetto termina il processo dei log
- [x] I colori dei Servizi entrano nel test dei token del ticket 03
- [x] Test da tavolo per la costruzione dei tag del log e per il tetto delle righe
- [x] Colori leggibili in tema chiaro e scuro, verificati sugli screenshot dei due temi

## Comments

La macchina degli stream resta nel package della GUI senza Qt, con i test di oggi. Il fermo degli stream alla chiusura della finestra arriva nel ticket 13, quando esiste il nascondi.

## Esito

### Struttura

- `internal/gui/qtlog.go` è la pagina del tab Log. Tiene il `QPlainTextEdit` in
  sola lettura, la barra dei comandi sopra il testo e la riga di stato sotto.
  Possiede lo stream: `Open` avvia `docker compose logs`, `Stop` termina il
  processo, `Clear` ferma il flusso e dimentica le righe.
- La barra ha il nome del bersaglio, l'etichetta «Cerca», la ricerca che prende
  lo spazio libero, la casella «Timestamp», il pulsante «Pausa» e il conteggio.
  Sta su una riga sola: nessuna colonna resta vuota. L'etichetta «Cerca» è
  visibile perché il colore del segnaposto di Qt non arriva a 4.5:1 sul tema
  scuro.
- `internal/gui/qtpanel.go` collega la pagina al pannello. Il tab Log apre il
  Progetto selezionato, `ShowLogs` apre il Progetto intero o un Servizio,
  `setProject` segue la selezione della shell. `Close` ferma lo stream.
- Il dettaglio ha due vie ai log: «Mostra log» sotto l'intestazione del
  Progetto, e «Log» su ogni Servizio. Il secondo apre il solo Servizio.

### Colori dei Servizi

- Le sedici tinte vivono in `designtokens.go`, come ruoli `service0`…`service15`
  con contrasto di testo. `ServiceColor(hue, theme)` legge il ruolo, quindi il
  colore di un Servizio ha una sola sorgente per i due temi.
- `designtokens_test.go` controlla ogni tinta nei due temi: colore `#rrggbb`,
  tipo di contrasto di testo, rapporto almeno 4.5:1 sulla superficie, ruolo
  presente, tinte distinte e diverse fra i due temi.
- L'ora usa il ruolo nuovo `logTimestamp`, con contrasto di testo nei due temi.

### stderr

- Una riga di stderr porta il prefisso `! `, il peso grassetto e il corsivo,
  oltre al colore di errore del tema. Il prefisso e il corsivo distinguono la
  riga anche in bianco e nero.

### Ricerca, pausa, timestamp, tetto

- La ricerca filtra le righe che contengono il testo, senza badare alle
  maiuscole. Il conteggio dice «N righe su M (massimo 5000)». La ricerca non
  guarda il nome del Servizio.
- La pausa ferma solo lo scorrimento automatico. Il ritorno riporta la vista
  all'ultima riga.
- La casella «Timestamp» mostra o nasconde l'ora di ogni riga. Lo stato si vede
  dalla spunta, non dal solo colore.
- Il tetto è `LogLineLimit`, 5.000 righe. La pagina scarta le più vecchie con
  `AppendLogLines`. Il widget ha `SetMaximumBlockCount(LogLineLimit+1)` come
  rete di sicurezza: una riga tiene due blocchi, il testo e il ritorno a capo.
- La selezione e la copia arrivano dal widget in sola lettura di Qt.
- La riga di stato mostra «In attesa di righe…», «Nessuna riga corrisponde alla
  ricerca.», «Il flusso dei log è terminato.» oppure l'errore di compose nel
  colore di guasto.

### Test

- `logview_test.go`: casi nuovi per `ServiceTag` (tag stabile, tag distinti, mai
  uguali a quelli dell'ora e di compose), per `serviceHueOf` (andata e ritorno,
  tag estranei), per il tetto con il limite vero del pannello, e per
  `LogTargetText`.
- `designtokens_test.go`: `TestServiceColoursComeFromTheRoles`.
- `qtshell_test.go`, passo `checkLogs` dello smoke test `ui`: tab del Progetto,
  colori dei ruoli, prefisso di stderr, ricerca con conteggio, ricerca senza
  esito, timestamp, pausa e ripresa, tetto di 5.000 righe, log di un Servizio,
  chiusura del pannello, cambio di Progetto. Ogni fine dello stream passa da un
  canale del finto `LogSource`.
- `captureLogs` scrive gli screenshot del tab nei due temi.

### Comandi eseguiti

```
gofmt -l .                      # solo file già presenti sotto .agents/tmp
go vet ./...
make check                      # vet, test, build
make test-ui                    # smoke test offscreen e token
QT_QPA_PLATFORM=offscreen go test -race -tags ui -run TestShellOffscreen ./internal/gui
SHOT_DIR=/tmp/opencode/shots QT_QPA_PLATFORM=offscreen go test -tags ui -run TestShellOffscreen ./internal/gui
```

### Screenshot

- `16-log-light.png` e `17-log-dark.png`, con tre Servizi e una riga di errore.
- Corretto dopo la revisione: il prefisso `! ` va davanti al testo di compose e
  non davanti all'ora. La casella dei timestamp mostra la spunta, quindi lo
  stato non dipende dal colore. L'etichetta «Cerca» sostituisce il segnaposto,
  che sul tema scuro non si leggeva. I contrasti delle tinte e dell'ora restano
  leggibili nei due temi.
- La verifica manuale su KDE resta aperta: qui non c'è un desktop.
