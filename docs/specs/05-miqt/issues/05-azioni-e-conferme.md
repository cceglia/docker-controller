# 05: Azioni e conferme

**What to build:** La riga delle Azioni lancia Up, Stop, Restart, Pull, Build, Down e Down con volumi sul Progetto, e Start, Stop, Restart e Down su ogni Servizio. Ogni Azione distruttiva chiede conferma in un dialogo modale, con il testo di oggi e il pulsante affermativo come predefinito. Un'Azione rifiutata dal Manager mostra l'errore utente in un avviso in sovrimpressione. Il doppio click non lancia due volte la stessa Azione.

**Blocked by:** 04

**Status:** resolved

- [x] Il Progetto offre Up, Stop, Restart, Pull, Build, Down e Down con volumi
- [x] Ogni Servizio offre Start, Stop, Restart e Down
- [x] Un'Azione non ammessa è disabilitata, con il motivo visibile
- [x] Down, Down di un Servizio e Down con volumi chiedono conferma con il testo di oggi
- [x] Annullare la conferma non lancia niente
- [x] Il pulsante affermativo del dialogo è predefinito, e Escape annulla
- [x] Il dialogo restituisce il focus al controllo che lo ha aperto
- [x] Un'Azione rifiutata mostra il messaggio utente in un avviso in sovrimpressione
- [x] Durante l'Azione il comando che l'ha lanciata è disabilitato
- [x] Il doppio click non produce due Attività
- [x] Test da tavolo per l'elenco delle Azioni ammesse e per i testi di conferma
- [x] Verifica manuale su KDE: ogni Azione arriva a Docker

## Comments

La decisione su quali Azioni sono ammesse resta nelle funzioni pure già coperte: `ProjectActions`, `ServiceActions`, `ActionDisabledReason`.

## Esito

Le Azioni del Progetto e dei Servizi arrivano al Manager. Un'Azione distruttiva
apre la sua domanda, e un rifiuto del Manager diventa il testo italiano di un
avviso in sovrimpressione. Il doppio click non produce due Attività.

**File toccati**:

| File | Ruolo |
| --- | --- |
| `internal/gui/qtdetail.go` | Riga delle Azioni del Progetto, riga per Servizio, scelta, lancio, guardia, focus di ritorno |
| `internal/gui/qtconfirm.go` | Dialogo di conferma costruito, modal, non bloccante |
| `internal/gui/qttoast.go` | Avviso in sovrimpressione della shell |
| `internal/gui/startguard.go` | Guardia del doppio avvio, senza Qt |
| `internal/gui/qtshell.go` | `do` lancia il Manager fuori dal thread Qt, `showNotice` scrive l'avviso |
| `internal/gui/qtstyle.go` | Ruolo `dialogTitle`, riquadro `toast`, testo `toastText` |
| `internal/gui/qtshell_test.go` | Smoke test delle Azioni, della domanda e dell'avviso |

### Azioni

`ProjectActions` costruisce sette voci per il Progetto, con Down e Down con
volumi in fondo e la conferma allegata. `ServiceActions` costruisce Start, Stop,
Restart e Down per ogni Servizio. Il pannello legge `view.Enabled` e
`view.Confirmation` e non decide niente da solo.

- Un comando ammesso è premuto e chiama `detailPanel.choose`.
- Un comando non ammesso resta disabilitato. `actionReasonText` scrive il motivo
  sotto la riga del Progetto, e `ActionDisabledReason` lo ripete come
  descrizione accessibile e tooltip di ogni comando del Progetto e dei Servizi.
- `choose` apre la domanda quando `Confirmation` non è nil, altrimenti lancia.
- `launch` chiude la guardia, disabilita il comando che ha lanciato e chiama
  `onAction`, che la shell collega a `do`.
- `shell.do` esegue `manager.Do` in una goroutine. Il risultato torna sul thread
  Qt con `mainthread.Start`: un errore passa da `UserError` all'avviso, poi il
  pannello apre di nuovo la guardia.

### Conferme

`newConfirmDialog` costruisce il `QDialog` prima di mostrarlo, quindi lo smoke
test lo controlla senza display. I testi vengono da `Confirmation`: titolo,
messaggio, etichetta affermativa.

- Il dialogo è modale e si apre con `Open`, quindi non blocca il chiamante.
- Il pulsante affermativo è `SetDefault(true)` e `SetAutoDefault(true)`. Invio
  conferma, Escape rifiuta con il comportamento standard di `QDialog`.
- Con `TypeToConfirm` il dialogo aggiunge il campo e il suggerimento, tiene il
  pulsante affermativo spento e lo accende solo quando il testo coincide.
- `onRun` gira solo su `accepted`. Un rifiuto, Escape e la chiusura non lanciano
  niente.
- `onClosed` gira su `finished`: il pannello cancella il puntatore, disegna lo
  snapshot che ha aspettato e dà il focus al comando che ha aperto la domanda,
  trovato per chiave. Se quel comando ora è disabilitato, prende il focus il
  primo comando disponibile.
- Mentre la domanda è aperta il pannello non si ricostruisce: lo snapshot
  aspetta, così il comando che ha aperto il dialogo non viene distrutto.

### Avviso in sovrimpressione

`toast` vive nella shell, non nel pannello, così il ticket 06 lo usa per
un'Attività fallita. `shell.showNotice(testo)` è l'unico punto d'ingresso.

- Il riquadro è figlio di `pages`, in alto a destra, sopra il contenuto e fuori
  dalla riga dei comandi in basso e dal pulsante "Indietro" in alto a sinistra.
- Non prende il focus (`NoFocus`) e lascia passare i click
  (`WA_TransparentForMouseEvents`), quindi non toglie un comando all'utente.
- Ha il nome accessibile "Avviso" e la descrizione accessibile con il messaggio.
- Resta sei secondi, la durata del vecchio avviso, poi `QTimer` lo nasconde. Un
  messaggio nuovo sostituisce il precedente e riavvia il tempo.
- `place` gira a ogni ridimensionamento della finestra.

### Doppio avvio

Due strati, più il Manager come autorità finale.

- `startGuard` è la regola pura: `begin` chiude la guardia al primo click e
  rifiuta il secondo, `release` la riapre quando il risultato torna, `snapshot`
  la riapre quando lo snapshot riporta l'Attività.
- Il comando cliccato si disabilita subito, prima di lasciare il thread Qt.
- `manager.Do` rifiuta comunque una seconda Attività sullo stesso Progetto con
  `ErrActivityRunning`, che diventa il testo italiano dell'avviso.

### Test

- `startguard_test.go`: test da tavolo delle sei sequenze (click singolo, doppio
  click, risultato, snapshot con Attività, snapshot senza Attività, lancio
  rifiutato).
- `actions_test.go`: resta invariato e verde. Copre l'elenco delle Azioni
  ammesse, l'ordine, l'abilitazione e i testi di conferma.
- `qtshell_test.go`: `checkActions` controlla i sette comandi del Progetto e i
  quattro di ogni Servizio, l'abilitazione, il motivo scritto durante
  un'Attività, l'apertura della domanda di Down e il rifiuto che non lancia.
  `checkConfirmDialog` controlla modalità, titolo, pulsante predefinito, testo da
  digitare, rifiuto che non lancia e conferma che lancia. Lo screenshot
  `04-notice` mostra l'avviso sopra il contenuto.

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
| `03-detail.png` | Le Azioni del Progetto e di ogni Servizio, abilitate |
| `04-notice.png` | L'avviso in alto a destra, sopra il contenuto |
| `05-detail-missing.png` | Il File compose mancante, con le Azioni ancora presenti |
| `06-detail-unreachable.png` | Docker non raggiungibile: comandi spenti e motivo scritto |

### Casi limite

- Il comando di un Servizio è chiavato "servizio/Azione", quindi il "Down" di un
  Servizio e quello del Progetto non si confondono quando il focus torna.
- Il backend offscreen non ha un widget con il focus: il ritorno del focus dopo
  la domanda si controlla solo a mano.
- La voce "Rimuovi dalla lista" e i comandi Cartella, Terminale e Porta restano
  disabilitati: appartengono al ticket 10.
- L'avviso non si ferma quando la finestra va nella tray: la tray è del ticket
  11 e il ciclo di vita della chiusura del ticket 13.

### Difetti visti nella prova manuale, e corretti

La prova dell'utente su KDE ha mostrato due difetti.

1. **I Container sembravano appartenere al Servizio successivo.** Il blocco di
   un Servizio metteva le Azioni prima dei Container, e lo spazio riservato
   all'Utilizzo risorse creava un vuoto in mezzo al blocco. Ora il blocco va dal
   nome ai comandi: nome e stato, Container, Azioni. Una linea del tema separa
   due Servizi. La linea è un `QFrame` con ombra `Sunken`: un `QWidget` semplice
   non dipinge lo sfondo del foglio di stile, e `palette(mid)` non si vedeva.
2. **I dialoghi tagliavano il testo.** Un messaggio con `SetWordWrap` dà al
   dialogo la larghezza di una parola. Ora il dialogo ha larghezza minima 420 px
   e massima 560 px, e `AdjustSize` misura il contenuto prima di aprirlo.

L'utente ha anche chiesto di togliere lo spazio vuoto riservato all'indicatore
di utilizzo: non serve e confonde. Il segnaposto `serviceUsage` sparisce. Il
ticket 08 costruirà l'indicatore dove sta meglio, senza un buco preventivo.

La revisione è stata fatta sugli screenshot offscreen in tema chiaro e in tema
scuro, più una cattura della finestra vera sotto `xcb` con `import`: il tema
scuro reale coincide con la palette di prova.

**Screenshot**: `03-detail`, `09-confirm`, `10-confirm-type`, `11-detail-dark`,
`12-confirm-dark` in `/tmp/opencode/shots`.

