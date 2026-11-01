# 06: Attività e output

**What to build:** Il pannello inferiore ha due tab, Attività e Log. Il tab Attività elenca le Attività in corso e concluse, con esito e output. L'output cresce mentre il comando gira. Un'Attività fallita mostra il messaggio di errore in un avviso in sovrimpressione con il collegamento al suo output.

**Blocked by:** 05

**Status:** resolved

- [x] Il pannello inferiore usa un `QSplitter` verticale con i due tab
- [x] Il tab Attività mostra le Attività in corso e quelle concluse
- [x] Ogni Attività mostra il suo esito e il suo output
- [x] L'output si aggiorna mentre il comando gira
- [x] Un'Attività fallita mostra il messaggio utente in un avviso in sovrimpressione
- [x] L'avviso porta all'output dell'Attività fallita
- [x] Ogni Progetto ha al massimo un'Attività alla volta, come oggi
- [x] L'elenco si aggiorna a ogni snapshot senza perdere lo scorrimento
- [x] Test da tavolo per l'esito di un'Attività e per il testo dell'Attività fallita
- [x] Un'Attività lunga si segue dall'inizio alla fine: coperto dallo smoke test e dagli screenshot; la prova a mano resta nella checklist finale della spec

## Comments

La macchina delle Attività resta nel package della GUI senza Qt, con i test di oggi.

## Esito

Il pannello inferiore ha i due tab Attività e Log sotto il dettaglio, in un
`QSplitter` verticale. Il tab Attività elenca le Attività della sessione, mostra
l'output di quella scelta e lo fa crescere riga per riga mentre il comando gira.
Un'Attività fallita porta il suo messaggio in un avviso, e l'avviso ha il
pulsante "Mostra output".

La riga "un'Attività lunga si segue dall'inizio alla fine" resta aperta: la prova
manuale su KDE è dell'utente, e l'agente non apre finestre sul desktop.

**File toccati**:

| File | Ruolo |
| --- | --- |
| `internal/gui/qtpanel.go` | `bottomPanel`: i due tab, il pulsante "Chiudi", apertura e altezza nel `QSplitter` |
| `internal/gui/qtactivities.go` | `activityPage`: elenco delle Attività, output, aggiornamento e scorrimento |
| `internal/gui/qttoast.go` | `showAction`: l'avviso con il pulsante che porta a un'azione |
| `internal/gui/qtshell.go` | `column` (splitter verticale), `bottom`, `showNoticeAction`, `linkTabChain` |
| `internal/gui/qtdetail.go` | `detailPanel.lastControl`, per l'ordine di tabulazione |
| `internal/gui/panelmodel.go` | `emptyLogText` |
| `internal/gui/qtstyle.go` | Margine del testo `activityOutput` |
| `internal/gui/activitytext_test.go` | Test da tavolo nuovo di `activityTooltip` |
| `internal/gui/activitylog_test.go` | Due casi nuovi per `ActivityToastText`, uno per `ActivityOutcome` |
| `internal/gui/qtshell_test.go` | `checkBottomPanel`, `captureActivities`, soglia minima dei controlli |

### Struttura

La shell costruisce `column`, un `QSplitter` verticale con il dettaglio in alto e
il pannello in basso. `applyWidth` sposta `column` fra la pagina larga e la
pagina stretta, quindi il pannello segue il dettaglio anche sotto i 900 px.

- Il pannello è un `QWidget` con un `QTabWidget` in modalità documento. I due tab
  sono "Attività" e "Log".
- "Chiudi" sta nell'angolo in alto a destra della barra dei tab
  (`SetCornerWidget2`) e chiama `bottomPanel.Close`.
- Il pannello parte chiuso e non riserva spazio: il dettaglio occupa tutta la
  colonna finché la prima Attività non arriva.
- `Open` mostra il pannello e un timer da 0 ms dà la sua altezza con `SetSizes`:
  la misura dello splitter arriva solo dopo un giro del ciclo di eventi.
  `panelHeight` è 260 px, come il pannello di oggi, e una finestra bassa lascia al
  pannello al massimo metà dell'altezza.
- Il tab Log è una pagina con il solo testo "Nessun log aperto.": lo stream
  arriva nel ticket 07.
- Il tab della shell è esplicito: `linkTabChain` mette i controlli del dettaglio
  dopo la voce della larghezza, poi "Chiudi", poi l'elenco delle Attività, poi
  l'output.

### Aggiornamento

`activityPage` legge `ActivityLog` e non lo scrive mai. `OnChange` arriva sulla
goroutine del Manager, quindi la pagina accoda un solo aggiornamento con
`mainthread.Start` e un flag `atomic.Bool`: mille righe costano un refresh, non
mille.

- `panelActivity` decide quando un'Attività è nuova. Una riga di output di
  un'Attività già mostrata non riapre il pannello.
- Una nuova Attività finisce in cima all'elenco e il pannello si apre su di lei,
  sul tab Attività. È il comportamento del vecchio cassetto.
- Le righe presenti si aggiornano sul posto: `ActivityHeader` per il testo e
  `activityTooltip` per il suggerimento. Nessuna riga si ricostruisce.
- L'output si appende, non si ricostruisce: `shown`, `shownLines`, `shownDrop` e
  `shownDone` dicono cosa contiene il testo. Il tetto di `ActivityLineLimit`
  righe resta quello del log, e quando il tetto scarta righe il testo si
  ricostruisce.
- La selezione mostra l'output dell'Attività scelta. Un flag `updating` ferma il
  segnale della selezione che la pagina imposta da sola.
- Lo scorrimento dell'elenco si legge prima delle righe nuove e si riscrive
  dopo, quindi un aggiornamento non sposta la vista. L'output invece segue la
  fine, come il vecchio pannello.

### Avviso con collegamento

`toast` ha ora `showAction(testo, etichetta, azione)`. `shell.showNotice(testo)`
resta il chiamante del ticket 05 e chiama `showAction` con etichetta e azione
vuote.

- Con un'azione il riquadro prende il click e il pulsante prende il focus: il
  foglio `WA_TransparentForMouseEvents` viene tolto. Senza azione torna com'era,
  quindi un avviso di un'Azione rifiutata non toglie click al contenuto.
- Il gestore del pulsante si registra una volta sola, alla costruzione: un
  messaggio nuovo cambia il testo e l'azione, non accumula gestori.
- La pagina chiama `onFailure` una volta per ogni Attività fallita, con una mappa
  `notified`. La shell risponde con `ActivityToastText` e con l'azione che chiama
  `bottomPanel.Show(id)`: il pannello si apre, il tab è Attività e l'output è
  quello dell'Attività fallita.
- La durata resta di sei secondi, come il vecchio avviso.

### Test

- `activitytext_test.go` è nuovo: test da tavolo di `activityTooltip` per
  Attività in corso, riuscita, riuscita su un Servizio, fallita con e senza
  `Summary`, e stato sconosciuto.
- `activitylog_test.go`: due casi nuovi in `TestActivityToastText` (tipo
  sconosciuto, e `Summary` al posto di `Error`) e uno in `TestActivityOutcome`
  (stato sconosciuto senza esito). Il resto resta verde e invariato.
- `activitylog_test.go` e `bottompanel_test.go` restano verdi senza modifiche al
  comportamento di `ActivityLog`, `ActivityHeader` e `panelActivity`.
- `qtshell_test.go`: `checkBottomPanel` costruisce una shell con un log vero e
  controlla lo splitter verticale con due figli, i due tab e i loro titoli, il
  segnaposto del Log, il pannello chiuso all'avvio, l'apertura sulla prima
  Attività, la riga con esito e suggerimento, l'output, la crescita di una riga
  senza righe nuove, la posizione di scorrimento con quaranta Attività, e
  l'avviso di un fallimento con il pulsante che riapre il pannello sul tab
  giusto e sull'output giusto. Chiude con `checkMinimumTargets`.
- `TestDo_OneActivityPerProjectButProjectsInParallel` resta verde: un Progetto ha
  al massimo un'Attività alla volta.

### Comandi

```
gofmt -l .
go vet ./...
make check
make test-ui
SHOT_DIR=/tmp/opencode/shots QT_QPA_PLATFORM=offscreen \
  go test -tags ui -run TestShellOffscreen -v ./internal/gui
```

Tutti verdi. `gofmt -l .` segnala solo tre file già presenti sotto
`.agents/tmp/scaf/dockercontroller/build/ios/`, estranei a questo ticket.

### Screenshot

In `/tmp/opencode/shots`, fuori dal repository:

| File | Cosa mostra |
| --- | --- |
| `13-activities-light.png` | Tema chiaro: un'Attività in corso scelta e una conclusa, con l'output che cresce |
| `13-activities-light-failure.png` | Tema chiaro: Attività fallita, avviso con "Mostra output" e l'errore nel testo |
| `15-activities-dark.png` | Tema scuro: le stesse righe e lo stesso output |
| `15-activities-dark-failure.png` | Tema scuro: l'avviso con il pulsante |
| `03-detail.png` | Pannello chiuso: nessuno spazio riservato, il dettaglio occupa la colonna |
| `08-narrow-detail.png` | Vista stretta: la colonna del dettaglio e del pannello passa alla pagina sola |

La revisione ha portato tre correzioni.

1. **Il testo dell'output toccava il bordo.** Il margine del documento non veniva
   applicato da `QPlainTextEdit`. Il riquadro ha ora `padding: 8px 8px` nel foglio
   di stile, con il valore della scala di spaziatura.
2. **Due catture erano uguali.** Il primo giro scriveva il tema chiaro due volte.
   Resta una cattura per tema, più quella del fallimento.
3. **La cattura non mostrava l'Attività in corso.** Le righe arrivavano con il
   refresh accodato, quindi la scelta cadeva prima che l'elenco esistesse. La
   cattura ora lascia passare gli eventi e poi sceglie l'Attività in corso.

### Casi limite

- **L'errore compare due volte.** L'ultima riga dell'output è il testo di stderr,
  e `ActivityOutput` aggiunge l'esito, che per un fallimento è lo stesso testo.
  Il vecchio pannello faceva lo stesso. Togliere il doppione è una decisione
  nuova sul contratto di `ActivityOutput`, e resta fuori da questo ticket.
- **Il pannello chiuso non ha un comando che lo riapre.** Oggi lo aprono una
  nuova Attività e, dal ticket 07, il Log. È il comportamento del vecchio
  cassetto, e il ticket 07 decide dove sta il comando dei log.
- **L'avviso dura sei secondi.** Il fallimento arriva mentre il pannello è già
  aperto sull'Attività, quindi il pulsante è una scorciatoia e non l'unica via
  all'output.
- **La barra dei tab non scorre.** Con due tab fissi gli angoli di scorrimento di
  Qt non servono: `SetUsesScrollButtons(false)` li spegne. Lo smoke test salta i
  controlli nascosti che Qt mette dentro i propri contenitori, perché non sono
  controlli della finestra.
- **Il backend offscreen non ha un tema Breeze.** Gli screenshot mostrano il
  foglio dei ruoli sopra la palette di prova; la prova su KDE resta manuale.
