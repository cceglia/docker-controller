# 09: Modifica Progetto

**What to build:** Dal dettaglio l'utente apre un form per Nome visualizzato, Gruppo e Profili attivi. Il form usa i valori salvati, propone i Gruppi già in uso e salva dallo stesso pannello. Un salvataggio non valido mostra l'errore sotto il campo. Uno snapshot che arriva durante la modifica non cancella quello che l'utente ha scritto.

**Blocked by:** 04

**Status:** resolved

- [x] Il pulsante di modifica apre il form con i valori salvati
- [x] Il form modifica Nome visualizzato, Gruppo e Profili attivi
- [x] I Gruppi già in uso sono proposti come scelta
- [x] Il salvataggio aggiorna la lista e il dettaglio senza riavviare l'app
- [x] Annullare riporta i valori salvati
- [x] Un Nome visualizzato vuoto torna al nome compose
- [x] Un campo non valido mostra il messaggio sotto il campo e il nome accessibile lo ripete
- [x] Uno snapshot durante la modifica non sovrascrive il testo digitato
- [x] Lo snapshot aggiorna il Progetto dopo il salvataggio
- [x] Test da tavolo per la normalizzazione di Nome visualizzato, Gruppo e Profili
- [ ] Verifica manuale su KDE: il form si usa solo con la tastiera

## Comments

La normalizzazione dei valori resta pura e già coperta dai test.

## Esito

### Struttura

- `internal/gui/qtedit.go` è il form della modifica. `detailPanel` tiene un
  `editForm`: i valori digitati, il messaggio di un salvataggio rifiutato e i
  widget dell'ultimo disegno. I valori vivono nel `ProjectEdit` del form e non nei
  widget, quindi un nuovo disegno ricostruisce il form senza perdere il testo.
- Il blocco `Dati del Progetto` ora porta il titolo e il comando "Modifica" in
  testa. Il form prende il posto del blocco dei dati. L'intestazione, le Azioni,
  i Comandi e i Servizi restano dove sono.
- Campi: "Nome visualizzato" (`QLineEdit`, con il nome compose come segnaposto),
  "Gruppo" (`QComboBox` modificabile) e "Profili attivi" (una casella di spunta
  per Profilo definito dai File compose). Ogni campo ha la didascalia sopra, poi
  il campo, l'errore e il suggerimento.
- Il Gruppo propone `GroupSuggestions(GroupChoices(...))`: i Gruppi in uso, senza
  la voce vuota, che resta il segnaposto "Nessun Gruppo". Un Gruppo nuovo si
  scrive nella casella.
- "Salva" e "Annulla" chiudono il form. Annulla non tocca il backend: il form
  tiene solo valori in attesa, quindi i valori salvati tornano da soli.

### Salvataggio

- `detailPanel.saveEdit` legge `ProjectEdit.Settings()` e chiama `onSave`. La
  shell lo collega a `Manager.UpdateSettings` in una goroutine. Il risultato torna
  sul thread Qt con `mainthread.Start` e arriva a `saveDone`.
- Un rifiuto lascia il form aperto e scrive `UserError(err).Error()` sotto il
  campo "Nome visualizzato". Un rifiuto del salvataggio riguarda il Progetto
  intero, non un valore: il messaggio va sotto il primo campo del form. Il nome
  accessibile del campo ripete il messaggio (`fieldAccessibleName`) e il focus
  torna sul campo, così la tastiera e il lettore di schermo lo trovano.
- Un salvataggio riuscito chiude il form e ridà il focus a "Modifica". Lo snapshot
  che il Manager pubblica porta i valori nuovi alla lista e al dettaglio, senza
  riavviare l'app.

### Snapshot durante la modifica

- `detailPanel.set` conserva i dati nuovi e attende: finché il form è aperto non
  disegna. Lo stesso vale per `settle`, che chiude una domanda ma non tocca il
  form. Il testo digitato, il Gruppo e le caselle restano.
- Alla chiusura il pannello disegna l'ultimo snapshot: dopo il salvataggio il
  Progetto è aggiornato, e un Annulla mostra i valori salvati.

### Tastiera

- L'ordine di tabulazione è quello di lettura: Nome visualizzato, Gruppo,
  Profili, Annulla, Salva. Il form entra nella catena dei controlli del dettaglio
  come gli altri blocchi.
- All'apertura il focus va sul primo campo. Alla chiusura torna a "Modifica".
- Due `QShortcut` con contesto `WidgetWithChildrenShortcut` sul blocco del form:
  Invio (e Invio del tastierino) salva, Escape annulla. Con il popup del Gruppo
  aperto i due tasti restano del popup: Invio sceglie la voce, Escape chiude il
  popup e il form resta aperto.

### Test

- `edit_test.go`: casi nuovi per `SetProfile` (spazi intorno al nome, nome
  vuoto), per `Settings` che conserva un Profilo non più definito, per
  `GroupChoices` e `GroupIndex` (spazi intorno al Gruppo), per `ProfileChoices`
  (spazi) e per `GroupSuggestions`.
- `qtedit_test.go`: `fieldAccessibleName` ripete il messaggio del campo.
- `qtshell_test.go`, passo `checkEditForm` dello smoke test `ui`: il comando apre
  il form con i valori salvati, il Gruppo proposto, l'ordine di tabulazione, il
  focus, uno snapshot che non cancella il testo, un rifiuto con il messaggio sotto
  il campo, il salvataggio che aggiorna lista e dettaglio, il nome vuoto che torna
  al nome compose, Annulla, Invio ed Escape.
- `captureScreens` scrive il form con l'errore in tema chiaro e in tema scuro.

### Comandi eseguiti

```
gofmt -l .                      # solo file già presenti sotto .agents/tmp
go vet ./...
make check                      # vet, test, build
make test-ui                    # smoke test offscreen e token
CGO_ENABLED=1 QT_QPA_PLATFORM=offscreen go test -race -tags ui -run TestShellOffscreen ./internal/gui
SHOT_DIR=/tmp/opencode/shots QT_QPA_PLATFORM=offscreen go test -tags ui -run TestShellOffscreen ./internal/gui
```

### Screenshot

- `18-form-light.png` e `19-form-dark.png`: il form con i valori salvati, il
  Gruppo "Casa" nella casella, i Profili debug e metrics accesi e trace spento, e
  «Progetto sconosciuto.» sotto il campo del nome. Il rosso legge su entrambe le
  superfici.
- Corretto dopo la revisione: la prima versione usava un `QFormLayout` a due
  colonne e il suggerimento del Gruppo veniva tagliato. Ora ogni campo è una
  colonna con didascalia, campo, errore e suggerimento, e il testo va a capo anche
  a 800 px.
- La verifica manuale su KDE resta aperta: qui non c'è un desktop.

