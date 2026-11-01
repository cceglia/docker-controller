# 08: Utilizzo risorse

**What to build:** Una barra in alto nel dettaglio mostra CPU, memoria, rete e disco della macchina. Ogni Servizio mostra l'utilizzo dei suoi Container con un indicatore e i valori numerici. Una macchina non raggiungibile mostra l'avviso al posto dei numeri. Le cifre si fermano quando la finestra si nasconde nella tray e ripartono quando torna visibile.

**Blocked by:** 04

**Status:** resolved

- [x] La barra globale mostra CPU, memoria, rete e disco della macchina
- [x] Ogni Servizio mostra l'utilizzo dei suoi Container con indicatore e valori
- [x] Gli indicatori sono `QProgressBar` con il valore numerico accanto
- [x] Un campione arrivato dopo lo stop non riporta cifre vecchie sullo schermo
- [x] Docker non raggiungibile mostra un avviso invece di numeri falsi
- [x] Un Servizio senza Container non mostra indicatore
- [x] Il numero di Container non cambia l'altezza delle righe
- [x] Test da tavolo per la formattazione di byte, percentuali, core e rate
- [ ] Il fermo e il riavvio alla visibilità della finestra si verificano nel ticket 13

## Comments

Le funzioni di formattazione restano pure e già coperte dai test. Il segnale di visibilità della finestra arriva nel ticket 13, insieme al comportamento di chiusura.

## Esito

### Struttura

- `internal/gui/qtusage.go` è la vista dell'Utilizzo risorse. `usageStrip` è la
  barra della macchina: una riga di quattro celle, CPU, Memoria, Rete e Disco,
  con due numeri per cella. La prima riga di una cella porta la didascalia, la
  barra quando la misura è una percentuale e il valore. La seconda porta ciò che
  il valore da solo non dice: i core della CPU, la memoria della macchina in cui
  sta quella usata, la direzione opposta di un rate.
- `usageRow` e `usageCell` sono le stesse celle per la barra e per un Servizio,
  così le due righe si leggono allo stesso modo.
- `serviceUsageRow` è l'indicatore di un Servizio, dentro il suo blocco sotto le
  righe dei Container e sopra le Azioni. Parte nascosto e non prenota spazio: un
  widget nascosto in un `QVBoxLayout` non riceve geometria.
- `detailPanel` tiene la barra fuori dai blocchi, sopra lo `QStackedWidget`,
  così resta al suo posto mentre il dettaglio scorre. Una revisione del
  Progetto non la ricostruisce.
- `usageTextsOf` in `internal/gui/usagetext.go` è pura: trasforma misure e rate
  nei testi della riga, con i suoi test da tavolo.
- `internal/gui/usagefeed.go` possiede lo stream. `Follow(project,
  dockerReachable)` avvia `docker stats` solo quando il dettaglio mostra un
  Progetto con un Container in esecuzione e Docker risponde. Un cambio di
  Progetto ferma il processo vecchio e ne avvia uno nuovo: `UsageStreamer`
  aspetta la fine del precedente, quindi non ne restano due aperti. Chiudere il
  Progetto, un Docker che non risponde e un Progetto senza Container in
  esecuzione fermano lo stream. Il fermo e il riavvio alla visibilità della
  finestra restano al ticket 13.

### Campioni vecchi

- `usageFeed` legge la generazione dello stream quando la sorgente parte, non
  quando il campione arriva. Ogni campione si porta dietro quella generazione e
  arriva al pannello con `mainthread.Start`.
- Al momento di disegnare, il pannello accetta il campione solo se la sua
  generazione è ancora quella corrente. Un campione in coda prima di uno stop
  non arriva sullo schermo, e con lui non torna nessuna cifra vecchia.
- Lo stop consegna un campione `nil`: la barra torna alla riga di attesa e gli
  indicatori dei Servizi lasciano la pagina.
- La generazione si alza solo su un cambio vero (avvio, fermo, cambio di
  Progetto): una revisione dello stesso Progetto non scarta i campioni in volo.

### Docker non raggiungibile

- Un campione con `DockerReachable` falso mostra «Docker non raggiungibile» nel
  ruolo `statusFault`, con l'errore nel tooltip, al posto dei numeri.
- Lo dice anche lo snapshot, prima che arrivi un campione: un `dockerReachable`
  falso toglie le cifre e gli indicatori subito, senza aspettare lo stream.
- Prima del primo campione la barra mostra «In attesa della prima misura…» e
  nessun numero. Prima del primo snapshot non parla di Docker.

### Altezza delle righe

- I valori della riga di un Servizio non vanno a capo: l'altezza dipende solo
  dal contenuto, non dalla larghezza della finestra né dal numero di Container.
  I valori a larghezza fissa sono le sole barre, che non portano testo.
- La barra della macchina invece va a capo sotto i 700 px circa, così nessuna
  cifra viene tagliata: resta una riga di celle fin dove lo spazio lo consente.

### Test

- `format_test.go`: casi nuovi per byte (un confine di 1024, il riporto
  dell'ultimo KiB, 1024 byte in unità decimali, poco sotto un MB), percentuali
  (un centesimo che arrotonda, una percentuale negativa), core (un quinto, cento
  core) e rate (poco sotto e a un kB/s).
- `usagetext_test.go`: i quattro testi di una riga, per la macchina, per un
  Servizio senza dimensione della macchina e per una riga a riposo.
- `usagebar_test.go`: `meterValue` porta una percentuale nei passi della barra,
  con il taglio a 0 e a 100 e il caso che non è un numero.
- `usagefeed_test.go`: avvio solo con Docker e un Container in esecuzione,
  consegna dei campioni, una sola stream per lo stesso Progetto, nuova stream
  dopo un cambio di Progetto, stop che toglie le cifre, campione in coda prima
  dello stop che non arriva, feed senza sorgente.
- `qtshell_test.go`, passo `checkUsage` dello smoke test `ui`: barra prima del
  primo campione, cifre dopo, valore della barra e testo del Servizio senza
  Container, altezza della riga a 940 px e con un Container in più, stop,
  avviso di Docker non raggiungibile dal campione e dallo snapshot, barra
  assente senza Progetto. `captureScreens` e `captureLogs` scrivono gli
  screenshot con i campioni.

### Comandi eseguiti

```
gofmt -l .                      # solo file già presenti sotto .agents/tmp
go vet ./...
make check                      # vet, test, build
make test-ui                    # smoke test offscreen e token
CGO_ENABLED=1 go test -race ./internal/gui -run TestUsage
SHOT_DIR=/tmp/opencode/shots QT_QPA_PLATFORM=offscreen go test -tags ui -run TestShellOffscreen ./internal/gui
```

### Screenshot

- `03-detail.png` e `11-detail-dark.png`: barra con CPU 18,7% e 1,50 core,
  memoria 6,39 GiB su 16,0 GiB, rete e disco; indicatore del Servizio «web».
- `03a-waiting.png`: la barra prima del primo campione, con la sola riga di
  attesa.
- `06-detail-unreachable.png`: l'avviso al posto dei numeri e nessun
  indicatore.
- `08-narrow-detail.png`: barra e indicatore a 800 px, ancora su una riga.
- Corretto dopo la revisione: l'avviso in sovrimpressione copriva la cella del
  disco, quindi ora parte sotto la barra; la barra va a capo sotto i 700 px
  invece di tagliare l'ultima cella.
- La verifica manuale su KDE resta aperta: qui non c'è un desktop.
