# 10: Aggiunta e aperture

**What to build:** Dalla finestra l'utente aggiunge un Progetto scegliendo i File compose con il selettore di file di Qt, rimuove un Progetto dalla lista, apre la cartella, il terminale e la porta nel browser. La rimozione dalla lista chiede conferma. L'aggiunta rifiutata mostra il motivo con il testo di oggi.

**Blocked by:** 04, 05

**Status:** resolved

- [x] "Aggiungi progetto" apre il selettore di file di Qt
- [x] Il selettore accetta più File compose e li mostra nell'ordine scelto
- [x] Un nome compose già presente mostra il rifiuto con il testo di oggi
- [x] Un insieme di file che compose non accetta mostra il rifiuto con il testo di oggi
- [x] Il Progetto aggiunto compare nella lista senza riavviare l'app, perché il pannello segue `Manager.Watch`
- [x] "Rimuovi dalla lista" chiede conferma e non tocca Docker
- [x] "Apri cartella" e "Apri terminale" usano `xdg-open` e il comando configurato
- [x] "Apri porta" apre l'indirizzo nel browser
- [x] I comandi restano disabilitati quando il File compose manca
- [x] Test da tavolo per l'ordine dei File compose e per i testi di rifiuto
- [ ] Verifica manuale su KDE: il selettore di file mostra i file nascosti quando serve

## Comments

La validazione dell'aggiunta resta nel backend, con i test di oggi.

## Esito

La finestra aggiunge un Progetto, ne toglie uno dalla lista e apre cartella,
terminale e porta. Ogni rifiuto mostra il testo italiano di oggi.

### Struttura

| File | Ruolo |
| --- | --- |
| `internal/gui/qtadd.go` | Il selettore di file, il form del nuovo Progetto, l'ordine dei File compose, il nome e i rifiuti |
| `internal/gui/qtshellactions.go` | `showAddDialog`, `addProject`, `selectProject`, `openFolder`, `openTerminal`, `openPort`, `removeFromList` |
| `internal/gui/qtdetail.go` | I tre comandi collegati e il comando di rimozione dell'avviso File compose mancante |
| `internal/gui/addproject.go` | `AddRefusal`: i rifiuti che la finestra può dire prima del Manager |
| `internal/gui/open.go` | `FirstPublishedPort`, `CommandViews` e i motivi dei comandi spenti |
| `internal/gui/remove.go` | `removeTooltip` del comando che lascia il Progetto in lista |
| `internal/gui/detailstate.go` | La regola di ricostruzione legge anche la cartella, perché i comandi la disegnano |
| `internal/gui/qtstyle.go` | Un ruolo per il nome del File compose nelle righe del form |

### Aggiunta

- "Aggiungi progetto" apre `QFileDialog` con `QFileDialog__ExistingFiles`, quindi
  più file in una scelta. Il filtro è `File compose (*.yml *.yaml)`, con
  `Tutti i file (*)` per un nome diverso.
- Il selettore è quello a widget di Qt (`DontUseNativeDialog`). Quel selettore
  porta l'azione "Show hidden files", quindi un File compose in una cartella
  nascosta è raggiungibile quando serve.
- Il form si apre solo dopo una scelta vera. Mostra i File compose nell'ordine
  restituito dal selettore, con il ruolo "principale" sulla prima riga. Ogni
  riga ha il comando "Rimuovi". Il comando "Aggiungi un altro File compose…"
  accoda altri file in fondo, quindi l'ordine resta quello della scelta.
- Il nome compose arriva da `compose.ProposeName` sul File compose principale.
  Il campo resta disabilitato finché non c'è un File compose. Un nome digitato
  non viene più sostituito.
- `AddRefusal` rifiuta un nome già salvato, un nome vuoto e l'assenza di File
  compose, con i testi dei test di oggi:
  "Esiste già un Progetto con questo nome.",
  "Nome non valido: usa lettere minuscole, cifre, - e _.",
  "Serve almeno un File compose.".
  Il resto lo rifiuta il Manager, e `UserError` porta il suo testo nel form.
- "Aggiungi" chiama `Manager.AddProject` in una goroutine. Un rifiuto lascia il
  form aperto con il motivo sotto il campo del nome e il nome accessibile del
  campo ripete il messaggio. Un salvataggio chiude il form e chiama
  `selectProject`, che apre il dettaglio del Progetto nuovo.
- Lo snapshot che il Manager pubblica porta la riga nuova alla lista. Il
  cablaggio è quello di `Watch`: non c'è un riavvio, ma la corsa completa con
  Docker non l'ho eseguita (vedi i casi limite).

### Rimozione dalla lista

- Il comando vive nell'avviso "File compose mancante". È attivo solo quando
  `CanRemoveFromList` lo ammette, e `RemoveDisabledReason` scrive il motivo nel
  tooltip e nella descrizione accessibile.
- Il comando apre `RemoveConfirmation` con `confirmDialog`. Un no, Escape e la
  chiusura non fanno niente. Un sì chiama `Manager.RemoveFromList` in una
  goroutine: il Manager rilegge `docker compose ps` e non ferma niente. Un
  rifiuto diventa l'avviso in sovrimpressione.

### Aperture

- "Cartella" apre la cartella del Progetto con `xdg-open`. "Terminale" apre il
  comando salvato nelle Impostazioni, con `{dir}` sostituito dalla cartella e
  `cmd.Dir` sulla cartella. "Porta" apre `http://localhost:<porta>` della prima
  porta pubblicata del Progetto.
- `FirstPublishedPort` legge i Servizi, i Container e le porte nell'ordine dello
  snapshot e salta un numero che non è una porta.
- `CommandViews` decide i tre comandi: un File compose mancante li spegne tutti,
  una cartella vuota spegne "Cartella" e "Terminale", nessuna porta pubblicata
  spegne "Porta". Il motivo è `MissingComposeText`, `NoFolderCommandText` o
  `NoPortCommandText`.
- Un errore di apertura arriva sulla shell e diventa il testo utente
  dell'avviso in sovrimpressione. Un click su un comando spento non fa niente.

### Test

- `addproject_test.go`: casi nuovi per l'ordine dei File compose fra due scelte
  e una rimozione, e per i testi di rifiuto di `AddRefusal`.
- `open_test.go`: casi nuovi per `FirstPublishedPort` (ordine, porta non valida,
  nessuna porta) e per `CommandViews` con i tre motivi.
- `detailstate_test.go`: la cartella ora ricostruisce il pannello, perché i
  comandi la disegnano.
- `qtshell_test.go`, passi nuovi dello smoke test `ui`:
  - `checkPicker`: il selettore è modale, ha il titolo di oggi, prende più file
    esistenti, offre il filtro dei File compose ed è quello a widget di Qt.
  - `checkAddForm`: l'ordine delle righe, il ruolo della prima, il nome proposto,
    i due testi di rifiuto, il salvataggio che chiude e nomina il Progetto, il
    form vuoto con il campo spento, la caduta del File compose principale.
  - `checkCommands`: i tre comandi attivi, il click che chiama la shell, i
    comandi spenti senza cartella, senza porta e con il File compose mancante,
    la conferma di rimozione con un no e con un sì.
- `internal/gui/qtshell_test.go` aggiorna la verifica del vecchio comando spento
  nell'avviso: ora è spento perché il Progetto ha Container, e il motivo è
  `ContainerRefusalText`.

### Comandi eseguiti

```
gofmt -l .                      # solo i tre file già presenti sotto .agents/tmp
go vet ./...
make check                      # vet, test, build
make test-ui                    # smoke test offscreen e token
SHOT_DIR=/tmp/opencode/shots QT_QPA_PLATFORM=offscreen go test -tags ui -run TestShellOffscreen ./internal/gui
```

### Screenshot

- `03-detail.png`: il dettaglio con "Cartella", "Terminale" e "Porta" attivi.
- `05-detail-missing.png`: l'avviso del File compose mancante con il comando
  spento, perché il Progetto ha ancora Container.
- `05b-detail-missing-remove.png`: lo stesso avviso con il comando "Rimuovi
  dalla lista" attivo e i tre comandi spenti.
- `20-add-form.png`: il form con due File compose nell'ordine scelto, il ruolo
  "principale" sulla prima riga e il nome proposto.
- `21-confirm-remove.png`: la domanda di rimozione con il testo di oggi.
- `08-narrow-detail.png` e `11-detail-dark.png`: i comandi restano su una riga
  stretta e leggono nel tema scuro.
- Corretto dopo la revisione: il nome del File compose nella riga del form non
  si distingueva dal ruolo. Ora è in grassetto con il ruolo `fileComposeName`.

### Casi limite

- La rimozione dalla lista ha un solo comando, quello dell'avviso File compose
  mancante, come vuole la struttura del ticket 04. Un Progetto Rimosso i cui File
  compose esistono ancora non ha quindi un comando per lasciare la lista. È una
  decisione nuova: o il comando entra nella riga dei comandi, o la lista cambia.
- La verifica manuale su KDE del selettore con i file nascosti resta aperta: qui
  c'è una sessione Wayland, ma nessun desktop con cui provare il selettore.
  Lo smoke test controlla la scelta del selettore a widget, che è quella che
  porta l'azione "Show hidden files".
- La corsa completa dell'aggiunta con Docker non l'ho eseguita, per non scrivere
  nel database dell'utente. Il Manager e la lista hanno i loro test; il
  cablaggio passa da `Watch`.

## Nota dopo la revisione

La rimozione aveva un solo comando, dentro l'avviso del File compose mancante.
Un Progetto Rimosso con i File presenti non poteva quindi lasciare la lista.
Ora il comando sta nella riga dei comandi. Se l'avviso del File compose mancante
è presente, il comando resta solo lì: una schermata mostra un comando solo.
