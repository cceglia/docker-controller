# 13: Impostazioni e chiusura

**What to build:** La finestra delle Impostazioni permette di scegliere cosa fa la chiusura della finestra: nascondi nella tray, chiedi ogni volta, esci. Il dialogo "Chiedi ogni volta" ha l'opzione di ricordare la scelta. Il comando del terminale si modifica e si valida. "Esci" chiude l'app e ferma gli stream. La chiusura nasconde la finestra e non la distrugge, e la riapertura ritrova il Progetto aperto e la cronologia delle Attività. Senza tray la chiusura esce, perché l'utente non potrebbe altrimenti riaprire la finestra. Le impostazioni restano nel database esistente.

**Blocked by:** 07, 08, 11

**Status:** resolved

- [x] Le Impostazioni offrono le tre scelte di chiusura
- [x] La scelta salvata vale al riavvio
- [x] "Chiedi ogni volta" apre un dialogo con l'opzione di ricordare la scelta
- [x] Il comando del terminale si valida e un comando non valido non si salva
- [x] Chiudere la finestra la nasconde e non la distrugge, quando la scelta è nascondi
- [x] Riaprire la finestra ritrova il Progetto aperto e la cronologia delle Attività
- [x] Nascondere la finestra ferma il pannello Log e l'Utilizzo risorse
- [x] L'output di un'Attività lanciata dalla tray continua con la finestra nascosta
- [x] "Esci" chiude l'app e ferma ogni processo esterno
- [x] Le impostazioni salvate restano leggibili dopo l'aggiornamento
- [x] Senza tray la chiusura esce, e l'opzione nascondi non è disponibile
- [x] I testi delle Impostazioni e del dialogo di chiusura non promettono più la perdita della cronologia
- [x] Il test della macchina di chiusura copre il testo nuovo e il caso senza tray
- [x] Test da tavolo per la validazione del comando del terminale
- [x] Le tre scelte di chiusura si comportano come il testo dice, verificate dallo smoke test offscreen; la prova a mano resta nella checklist finale

## Comments

Il comportamento di chiusura cambia rispetto all'ADR 0003: la finestra non si distrugge più. Due testi utente e un test promettono ancora la perdita della cronologia delle Attività, e vanno riscritti qui.

## Esito

### Struttura

| File | Ruolo |
| --- | --- |
| `internal/gui/preferencesmodel.go` | Testi e funzioni pure delle Impostazioni: scelte, disponibilità, scelta mostrata, validazione del comando, valore salvato, esito del salvataggio. |
| `internal/gui/qtpreferences.go` | Il `QDialog` delle Impostazioni: due gruppi, `Salva` e `Annulla`, messaggio della validazione sotto il campo. |
| `internal/gui/closebehavior.go` | Macchina di chiusura pura: `closeActionFor` con il tray, risposte disponibili, testo della domanda per desktop con e senza tray, testi. |
| `internal/gui/qtclose.go` | Il dialogo "Chiedi ogni volta" e il percorso di chiusura dell'`application`: `closeRequested`, `performClose`, `askClose`, `answerClose`. |
| `internal/gui/qtshell.go` | Evento di chiusura della finestra, `hideWindow`, `stopStreams` e `startStreams`, pulsante "Impostazioni" collegato. |
| `internal/gui/qtapp.go` | Lettura delle impostazioni prima della finestra, `QuitOnLastWindowClosed(false)`, i campi nuovi dell'`application`. |
| `internal/gui/qtpanel.go` | `Suspend` e `Resume` del pannello: il flusso dei log si ferma con la finestra. |
| `internal/gui/usagefeed.go` | `SetHidden` del feed: `docker stats` gira solo con la finestra a schermo. |
| `internal/gui/qttray.go` | `quit` ferma gli stream prima di `QCoreApplication_Quit` e non si ripete. |

`compose`, `db`, `tray` e `notify` non cambiano. In `internal/settings` cambia un solo commento: quello di `CloseHide` diceva che la finestra si distruggeva.

### Comportamento

- **Impostazioni.** Il pulsante della finestra apre un `QDialog` modale con due gruppi: le tre scelte di chiusura e il comando del terminale. La scelta salvata è quella selezionata. Senza tray "Nascondi" è spento e una riga dice perché. `Salva` valida il comando sul posto: un comando non valido mostra il messaggio italiano sotto il campo, resta nel dialogo e non tocca il database. `Annulla` e Escape chiudono senza scrivere. Un salvataggio che riesce chiude il dialogo, scrive nel `Store` e conferma con "Impostazioni salvate." nella finestra.
- **Scelta salvata.** `Run` legge le impostazioni prima della finestra, quindi la chiusura risponde senza una query. Dopo ogni salvataggio l'`application` tiene la copia che il `Store` ha scritto: un comando vuoto vale il predefinito, come nel database.
- **Chiusura.** La finestra non si distrugge più: l'evento di chiusura è rifiutato e la scelta decide. `Nascondi` ferma i due stream e nasconde la finestra, `Chiedi` apre la domanda, `Esci` chiude l'app. Senza tray ogni scelta esce, e la domanda non offre "Nascondi".
- **Ricorda la scelta.** La casella salva la risposta come impostazione tramite `saveCloseBehavior`, che legge e riscrive le impostazioni esistenti. "Annulla" non salva e non chiude.
- **Stream.** `hideWindow` ferma il pannello Log e `docker stats`. `present` e il ritorno dal tray li riavviano. L'output di un'Attività continua: il `Manager` lo spinge nel log della sessione, che nessuna finestra possiede. Il ritorno ritrova il Progetto aperto, la lista e la cronologia delle Attività.
- **Esci.** Il pulsante della finestra e la voce della tray passano da `application.quit`, che ferma gli stream e poi chiama `QCoreApplication_Quit`.

### Test

| Test | Copertura |
| --- | --- |
| `TestCloseActionFor` | Tabella con il tray: nascondi, chiedi ed esci con tray; ogni scelta senza tray esce. |
| `TestAskCloseText_DoesNotPromiseTheLossOfTheHistory` | Testo nuovo: nomina la tray e la cronologia, e non contiene "persa", "perdi" o "va perso". |
| `TestAskCloseAnswers_AreHideQuitAndCancel` | Le tre risposte con tray, nell'ordine. |
| `TestAvailableCloseAnswers_WithoutATrayLeaveNoHide` | Senza tray restano "Esci" e "Annulla". |
| `TestCloseQuestionBody_WithoutATrayDoesNotOfferHiding` | Il corpo della domanda senza tray non offre di nascondere. |
| `TestCloseBehaviorAvailable` e `TestCloseBehaviorShown` | La scelta disponibile e la scelta mostrata: "hide" senza tray legge come "quit". |
| `TestValidateTerminalCommand` | Tabella: vuoto, spazi, comando semplice, virgolette doppie e singole, virgoletta aperta, backslash finale. |
| `TestStoredSettings` | Il valore che il `Store` tiene: spazi tolti, comando vuoto sostituito dal predefinito. |
| `TestUsageFeed_AHiddenWindowStopsTheStreamAndTheReturnStartsIt` | Il feed ferma `docker stats` con la finestra nascosta e lo riavvia al ritorno. |
| `checkPreferences` (smoke `ui`) | Le tre scelte, il comando non valido con il messaggio sotto il campo, il salvataggio che chiude, il caso senza tray. |
| `checkCloseDialog` (smoke `ui`) | Modale, tre risposte, "Annulla" che non risponde, "Nascondi" ricordato, due risposte senza tray. |
| `checkHiddenWindow` (smoke `ui`) | La chiusura vera nasconde la finestra e non la distrugge, il flusso dei log si ferma, l'output dell'Attività continua da nascosto, il ritorno ritrova selezione e cronologia. |
| `checkQuit` (smoke `ui`) | `quit` segna l'uscita, ferma il flusso dei log e non si ripete. |

`close_test.go` e `preferences_test.go` restano verdi. Il test del testo vecchio è stato riscritto: ora controlla che la cronologia non venga più promessa come persa.

### Comandi

```
gofmt -l internal main.go          # nessun file
go vet ./...                       # pulito
make check                         # vet, test e build con cgo
make test-ui                       # smoke offscreen e token di design
SHOT_DIR=/tmp/opencode/shots QT_QPA_PLATFORM=offscreen go test -tags ui -run TestShellOffscreen ./internal/gui
```

`gofmt -l .` segnala tre file sotto `.agents/tmp/scaf/`, che non sono del repository e non sono stati toccati.

### Screenshot

Le immagini stanno in `/tmp/opencode/shots` e non entrano nel repository.

| Immagine | Cosa mostra |
| --- | --- |
| `22-preferences-light.png` | Le Impostazioni in tema chiaro: le tre scelte, il comando e i due pulsanti. |
| `22b-preferences-error.png` | Lo stesso dialogo con il messaggio del comando non valido sotto il campo. |
| `23-close-light.png` | La domanda di chiusura con tray: "Nascondi", "Esci", "Annulla" e la casella. |
| `24-preferences-dark.png` | Le Impostazioni su Breeze Dark, senza tray: "Nascondi" spento e la sua riga. |
| `25-close-dark.png` | La domanda di chiusura su Breeze Dark, senza tray: solo "Esci" e "Annulla". |

Cosa ho corretto dopo aver guardato le immagini:

- La domanda di chiusura senza tray diceva ancora che la finestra può restare nel tray, mentre il desktop non ha l'icona. Ho aggiunto `askCloseBodyNoTray` e la funzione pura `closeQuestionBody`, con il suo test.
- Il testo del gruppo della chiusura parlava ancora di cronologia persa: riscritto prima della cattura.

Il resto tiene: i due gruppi hanno margini e spazi del ritmo 4-8-12-16, i pulsanti stanno a destra con "Salva" evidenziato, il messaggio di errore è rosso e leggibile su entrambi i temi, e la scelta spenta resta visibile con la sua spiegazione.

### Casi limite

- **Scelta "nascondi" salvata su un desktop senza tray.** La finestra non potrebbe tornare, quindi la chiusura esce e il dialogo mostra "Esci" al posto di "Nascondi". Un salvataggio scrive "quit": il dialogo mostra ciò che la chiusura fa davvero.
- **"Chiedi ogni volta" senza tray.** La scelta resta valida, ma la domanda offre solo "Esci" e "Annulla" e un corpo che non propone di nascondere.
- **Ritorno dal nascosto e log.** Il flusso dei log riparte dalle ultime 500 righe, quindi il testo riparte da capo. La cronologia delle Attività resta intatta, perché vive in memoria.
- **Attività nuova con la finestra nascosta.** Il pannello passa alla scheda Attività, quindi il flusso dei log riparte solo quando l'utente torna alla scheda Log.
- **Impostazioni illeggibili all'avvio.** `Run` applica i valori predefiniti e prosegue: l'utente non resta chiuso dentro la finestra.
