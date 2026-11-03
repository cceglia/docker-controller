# Sottopackage di compose e gui

**Status:** ready-for-agent

**Amends:** `docs/adr/0006-sottopackage-compose-e-gui.md`

## Problem Statement

`internal/compose` e `internal/gui` sono due package piatti e troppo grandi.
Compose ha 27 file e circa 5.200 righe; gui ne ha 75 e circa 16.700, di cui
circa 10.000 di adattatori Qt. Chi apre il repository non capisce dove cercare:
il parsing dell'output Docker sta accanto alla persistenza e alla porta verso la
CLI, e la logica di presentazione sta accanto ai widget.

Per il manutentore il costo è concreto: un file di 900 righe si legge tutto per
cambiare una riga, e un test che prova una funzione pura deve convivere con il
build cgo di Qt. Per l'utente il costo è indiretto: un refactor più lento, un
giro di test più fragile e un rischio maggiore di regressione su Progetti,
Attività, Log e Utilizzo risorse.

## Solution

Dividere i due package in sottopackage dentro `internal/`, senza cambiare il
comportamento dell'app.

Compose si divide in `model` (tipi di dominio, helper puri, sentinelle),
`docker` (porta e adattatore della CLI, con il doppio di test sotto `fake`),
`store` (persistenza SQLite) e `parse` (parsing dell'output della CLI). Il
package radice tiene `Manager`, che resta l'unica interfaccia del modulo, e
riesporta con alias i tipi usati da gui, tray e main.

Gui si divide in `design`, `view`, `project`, `forms`, `activity`, `logs`,
`usage` e `traymenu`, più `guitest` per gli helper condivisi dai test. Le viste
Qt restano in `internal/gui`: sono adattatori, e spostarle richiederebbe di
esportare l'intera superficie dei widget usata dalla shell. I sottopackage
contengono solo codice senza Qt, già testabile senza display. `view` è il
package foglia: nessun package di feature lo importa al contrario.

Ogni nuovo package entra in Gremlins con il proprio file di configurazione. Le
soglie di compose e gui vengono rimisurate dopo lo split, e i package che oggi
sono coperti solo attraverso `Manager` ricevono test di package.

## User Stories

### Comportamento dell'app

1. Come utente, voglio che Progetti, Servizi, Container, Attività, Log e Utilizzo risorse funzionino come prima, così da non accorgermi del refactor.
2. Come utente, voglio che le Azioni e le conferme restino identiche, così da non rischiare un Down per sbaglio.
3. Come utente, voglio che Nome visualizzato, Gruppo e Profili attivi restino nel database esistente, così da non riconfigurare nulla.
4. Come utente, voglio che la tray e le notifiche continuino a funzionare, così da comandare l'app senza aprire la finestra.
5. Come utente, voglio che la finestra si apra, si nasconda e si chiuda come prima, così da ritrovare le stesse abitudini.
6. Come utente, voglio che l'Utilizzo risorse si fermi quando la finestra si nasconde, così come oggi.
7. Come utente, voglio che un Docker irraggiungibile mostri lo stesso avviso, così da capire cosa manca.
8. Come utente, voglio che i Log tengano le ultime 5.000 righe e si fermino alla chiusura del pannello, così da non lasciare processi appesi.
9. Come utente, voglio che un'Attività fallita mostri lo stesso messaggio in sovrimpressione, così da accorgermene subito.

### Package di compose

10. Come manutentore, voglio che i tipi di dominio di compose vivano in `model`, così da sapere dove cercare Progetto, Servizio, Container e Snapshot.
11. Come manutentore, voglio che la porta verso la CLI Docker e il suo adattatore stiano in `docker`, così da isolare l'unico punto che esegue processi.
12. Come manutentore, voglio che il doppio di test stia sotto `docker/fake`, così da non mescolarlo al codice di produzione.
13. Come manutentore, voglio che la persistenza stia in `store`, così da leggere le query SQL in un solo posto.
14. Come manutentore, voglio che il parsing dell'output CLI stia in `parse`, così da provarlo senza un Docker vero.
15. Come manutentore, voglio che `Manager` resti l'unica interfaccia di compose, così da non cambiare i chiamanti.
16. Come manutentore, voglio che i tipi usati da gui, tray e main restino raggiungibili con alias, così da non toccare quegli import.
17. Come manutentore, voglio che `Do`, `Logs` e `Usage` restino metodi su `Manager`, così da non introdurre componenti pass-through.
18. Come manutentore, voglio che `Config`, `New` e le sentinelle mantengano lo stesso contratto, così da non rompere l'avvio.
19. Come manutentore, voglio che la direzione delle dipendenze sia `model` ← `store`/`parse` ← radice, con `docker` indipendente, così da non creare cicli.
20. Come manutentore, voglio che `model` contenga anche gli helper puri (`StatusOf`, `Folder`), così da non lasciare logica sparsa.
21. Come manutentore, voglio che i test di `exec` seguano l'adattatore in `docker`, così da provare il codice dove vive.
22. Come manutentore, voglio che le sentinelle usate dai sottopackage stiano in `model`, così che `store` e `parse` non importino la radice.

### Package di gui

23. Come manutentore, voglio che i token di design stiano in `design`, così da cambiare la scala tipografica in un posto.
24. Come manutentore, voglio che formattazione, testi condivisi e apertura di cartelle, terminali e URL stiano in `view`, così da non duplicarli tra le viste.
25. Come manutentore, voglio che azioni, rimozione dalla lista e stato del dettaglio stiano in `project`, così da isolare la presentazione del Progetto.
26. Come manutentore, voglio che i modelli dei form di aggiunta, modifica, impostazioni e chiusura stiano in `forms`, così da provarli senza Qt.
27. Come manutentore, voglio che il log delle Attività stia in `activity`, così da provarlo senza display.
28. Come manutentore, voglio che stream e vista dei Log stiano in `logs`, così da isolare il ciclo di vita del processo.
29. Come manutentore, voglio che feed e stato dell'Utilizzo risorse stiano in `usage`, così da provare la generazione degli stream.
30. Come manutentore, voglio che l'albero del menu della tray stia in `traymenu`, così da provarlo senza Qt.
31. Come manutentore, voglio che le viste Qt restino in `internal/gui`, così da non esportare la superficie dei widget.
32. Come manutentore, voglio che nessun sottopackage di gui importi la radice, così da evitare cicli.
33. Come manutentore, voglio che `gui.Run` resti il punto di avvio della UI, così da avere un solo seam dell'app.
34. Come manutentore, voglio che `main.go` cambi solo il costruttore del log delle Attività, così da non toccare il wiring.
35. Come manutentore, voglio che `notify`, `startguard` e le costanti di finestra restino in radice, così da non creare package da un file.
36. Come manutentore, voglio che i test non-Qt si spostino con il codice che provano, così da mantenere la copertura.
37. Come manutentore, voglio che il test della shell con tag `ui` resti in radice, così da non spostare il seam Qt.

### Test e mutation testing

38. Come manutentore, voglio che `parse` abbia test table-driven sui `testdata` esistenti, così da provare i formati reali di Docker.
39. Come manutentore, voglio che `store` abbia test su un database SQLite temporaneo, così da provare le query senza un database di produzione.
40. Come manutentore, voglio che `model` abbia test sugli helper puri, così da avere mutanti da misurare.
41. Come manutentore, voglio che ogni package abbia il suo file Gremlins, così da misurare il codice dove vive.
42. Come manutentore, voglio che i file Gremlins si chiamino dal path (`compose-model.yaml`, `gui-view.yaml`), così da non avere ambiguità.
43. Come manutentore, voglio che `make mutate` giri tutti i package e che `PKG=` ne misuri uno, così da scegliere fra giro completo e mirato.
44. Come manutentore, voglio che le soglie siano misurate dopo lo split e pari al misurato meno 5 punti, così da non bloccare i cambiamenti.
45. Come manutentore, voglio che un package senza mutanti sia annotato invece di avere una soglia inventata, così da non fingere una misura.
46. Come manutentore, voglio che il blocco condiviso dei file Gremlins sia verificato da un test, così da non divergere.
47. Come manutentore, voglio che `docker/fake` e `guitest` restino esclusi dalla mutation testing, così da non misurare i doppi di test e gli helper.
48. Come manutentore, voglio che `make test-ui` usi `./internal/gui/...`, così da coprire anche i nuovi package se un giorno ospiteranno test UI.

### Build, documentazione, manutenzione

49. Come manutentore, voglio che `go build ./...`, `go test ./...` e `go vet ./...` passino dopo ogni gruppo di spostamenti, così da isolare gli errori.
50. Come manutentore, voglio che `make check` resti la porta unica di validazione, così da non cambiare il flusso.
51. Come manutentore, voglio che l'ADR 0006 registri la scelta, così che un lettore futuro capisca perché le viste Qt restano in radice.
52. Come manutentore, voglio che il README aggiorni tabella delle soglie e percorsi, così da non leggere numeri vecchi.
53. Come manutentore, voglio che ogni package abbia un commento di package con responsabilità e regola di dipendenza, così da orientarmi senza leggere tutto.
54. Come manutentore, voglio un commit per gruppo di package spostati, così da poter tornare indietro su un passo.
55. Come manutentore, voglio che il comportamento dell'app resti invariato, così da non dover rifare la verifica manuale della UI.

### Coerenza dei package

56. Come manutentore, voglio che `view` non importi nessun package di feature, così da rompere i cicli `view` ↔ `logs`, `view` ↔ `forms` e `view` ↔ `project`.
57. Come manutentore, voglio che `ProjectName`, `DisplayNameOf` e `MissingComposeText` vivano in `view`, così che i package di feature li usino senza importarsi a vicenda.
58. Come manutentore, voglio che i messaggi degli errori di fine stream stiano in `logs`, così che `view` non conosca `ErrLogOpen`.
59. Come manutentore, voglio che il foglio di stile e `ServiceColor` stiano in `design`, così che il test dei token provi il foglio senza importare la radice.
60. Come manutentore, voglio che `serviceHueOf` stia in `logs`, così che il test della vista Log non dipenda dalla radice.
61. Come manutentore, voglio che `MeterValue` stia in `usage`, così che il test della barra si sposti con il calcolo.
62. Come manutentore, voglio un package `guitest` per gli helper condivisi, così che i test di `activity`, `view`, `forms`, `usage` e radice non duplichino lo stesso codice.
63. Come manutentore, voglio che `qtedit_test` resti in radice, così da provare `fieldAccessibleName` dove vive.
64. Come manutentore, voglio che `model` contenga `LogRequest` e `LogLine`, così che `parse` non importi la radice.
65. Come manutentore, voglio che `parse` contenga `Machine` e `StatsRow`, così che il parsing non crei un ciclo con `usage`.
66. Come manutentore, voglio che i test di `parse` leggano `../testdata`, così da riusare le fixture senza duplicarle.

## Implementation Decisions

**Compose.** Il package radice resta il punto di composizione e conserva `Manager`, `Config`, `New`, `Snapshot`, `Watch`, `Refresh`, `Run`, `Do`, `Logs`, `Usage`, `AddProject`, `RemoveFromList`, `UpdateSettings` e la cache dei Profili. I metodi non diventano componenti separati: sono metodi su `*Manager`, e spostarli richiederebbe un ridisegno del modulo senza un guadagno di profondità.

- `model`: tipi di dominio (`Project`, `Service`, `Container`, `Port`, `Snapshot`, `Status`, `ActionKind`, `ActivityID`, `ActivityRef`, `Settings`, `Discovered`, `LogRequest`, `LogLine`, `Measures`), helper puri (`StatusOf`, `Folder`) e sentinelle usate dai sottopackage.
- `docker`: la porta `Docker`, l'adattatore `ExecDocker` e il doppio di test sotto `fake`.
- `store`: la persistenza, con `Store` e `New` esportati, i metodi `UpsertDiscovered`, `Add`, `UpdateSettings`, `List` e `Remove`, e il database ricevuto dal chiamante.
- `parse`: le funzioni pure che leggono l'output della CLI (elenco Progetti, stato Container, Log, Utilizzo risorse, Profili, dimensione macchina), con i tipi di appoggio del parsing (`Machine`, `StatsRow`, `ContainerRow`) e il raggruppamento delle righe di `docker stats` accanto a loro.

La radice riesporta con alias solo la superficie usata fuori da compose (gui, tray, main) e nelle firme di `Manager`. Dentro compose i sottopackage parlano direttamente `model.*`. La direzione delle dipendenze è `model` ← `store`/`parse` ← radice, con `docker` indipendente da `model`; nessun sottopackage importa la radice.

**Gui.** Il package radice resta la composizione Qt: applicazione, shell, pannello inferiore, widget condivisi, `notify`, `startguard` e costanti di finestra. Le viste Qt restano qui.

- `design`: token di design, foglio di stile e colore di una voce della palette dei Servizi.
- `view`: package foglia con formattazione numerica, stato e tacche del Progetto, sezioni della lista, `ProjectName`, `DisplayNameOf`, `MissingComposeText`, testi di errore e apertura di cartelle, terminali e URL.
- `project`: azioni e relative etichette, rimozione dalla lista, formattazione del dettaglio e stato del disegno.
- `forms`: modelli di aggiunta, modifica, impostazioni e comportamento di chiusura.
- `activity`: log delle Attività, testi e modello del pannello.
- `logs`: ciclo di vita degli stream, modello della vista Log, `serviceHueOf` e i messaggi degli errori di fine stream.
- `usage`: feed, stream, stato dell'Utilizzo risorse e `MeterValue`.
- `traymenu`: albero del menu della tray.
- `guitest`: helper condivisi dai test (eventi di Attività, asserzioni su slice, attesa di un segnale); non è codice di produzione.

La direzione delle dipendenze è `design`/`view` ← package di feature ← radice. `view` non importa nessun package di feature: i simboli che creavano cicli (`ProjectName`, `DisplayNameOf`, `MissingComposeText`) vivono lì, mentre i messaggi dei Log stanno in `logs` e `MeterValue` in `usage`. Nessun sottopackage importa `internal/gui`.

**Avvio e API.** `gui.Run` resta il punto di avvio della UI; cambia solo il tipo del log delle Attività, che passa al package `activity`. `main.go` aggiorna il costruttore del log. `compose.Manager` e le sentinelle mantengono la firma per gui, tray e main.

**Gremlins.** Ogni package di codice di produzione entra in `MUTATE_PACKAGES` con un file di configurazione il cui nome deriva dal path, con `/` sostituito da `-` e `internal/` tolto (`compose-model.yaml`, `compose-docker.yaml`, `gui-view.yaml`, `gui-activity.yaml`, …). `make mutate` esegue tutti i package; `make mutate PKG=…` ne esegue uno. La regola della soglia resta quella del README: valore misurato meno 5 punti. Un package senza mutanti viene annotato nel proprio file e nel README, senza soglia inventata. `docker/fake` e `guitest` non entrano in `MUTATE_PACKAGES` e `exclude-files` li esclude, perché sono codice di supporto ai test.

**Documentazione.** Il README aggiorna la tabella delle soglie e il percorso di `make test-ui`. Ogni nuovo package ha un commento di package con responsabilità e regola di dipendenza. L'ADR 0006 registra la decisione.

**Nessun cambiamento di comportamento.** Stessi comandi Docker, stesso schema del database, stessa UI, stesse notifiche. Nessuna migrazione.

## Testing Decisions

Un buon test prova il comportamento esterno del modulo, non la sua implementazione: per compose il comportamento resta quello osservabile attraverso `Manager`; per i sottopackage è quello delle funzioni esportate; per gui è quello delle API esportate dai package di feature e, per la UI, quello della shell.

**Seam pubblici, invariati.** `compose.Manager` resta il seam di compose: i test esistenti di Manager, Attività, Log, Utilizzo risorse, Profili, rimozione e impostazioni non cambiano. `gui.Run` e la shell restano il seam dell'app: `TestShellOffscreen` (tag `ui`, offscreen) resta l'unico test che attraversa Qt.

**Seam interni, per package.** Servono perché Gremlins misura il singolo package:

- `parse`: test table-driven sui `testdata` già presenti, con i formati reali di `compose ls`, `compose ps`, Log, Utilizzo risorse e Profili, letti da `../testdata` perché la directory di lavoro di un test è quella del package.
- `store`: test su un database SQLite temporaneo aperto con le migrazioni esistenti, con `db.Open` su un percorso dentro `t.TempDir()`; stesso stile dei test di `Manager`, che già costruiscono un database usa-e-getta.
- `model`: test degli helper puri (`StatusOf`, `Folder`).
- `docker`: il test dell'adattatore si sposta con l'adattatore.
- Gui: i test non-Qt si spostano con il codice che provano. `designtokens_test` va in `design` insieme al foglio di stile, e porta con sé gli helper di contrasto che oggi stanno in `logview_test`, perché solo quel test li usa; `logview_test` e i test degli stream vanno in `logs`; `usagebar_test` va in `usage`; `open_test` e la parte generica di `user_errors_test` vanno in `view`; la parte dei Log di `user_errors_test` va in `logs`; `qtedit_test` resta in radice perché prova `fieldAccessibleName` della vista. `notify_test` e `qtshell_test` restano in radice e usano `guitest`. Nessun nuovo test Qt.
- `guitest`: contiene solo helper, non test; è il posto unico per gli eventi di Attività usati dai test, le asserzioni su slice e l'attesa di un segnale, così che i test di `activity`, `view`, `forms`, `usage` e radice non duplichino lo stesso codice.

**Prior art.** I test table-driven di formattazione e token di design in gui; i test di `Manager` che usano il doppio Docker e un database temporaneo; i test della tray per le funzioni pure del menu.

**Mutation testing.** Dopo lo split si esegue `make mutate` su tutti i package, si registrano efficacia e copertura dei mutanti, e si scrive la soglia pari al misurato meno 5 punti. Un test verifica che `workers`, `timeout-coefficient` ed `exclude-files` siano identici in tutti i file Gremlins. `internal/tray` non cambia e non viene rimisurato.

## Out of Scope

- Dividere `internal/tray`, `internal/settings`, `internal/db` e `internal/notify`.
- Un package per schermata: le viste Qt restano in `internal/gui`.
- Trasformare `Do`, `Logs` e `Usage` in componenti separati da `Manager`.
- Una cartella `packages/` o `pkg/` fuori da `internal/`.
- Cambiamenti di comportamento, nuove funzionalità, restyling della UI, modifiche allo schema del database.
- Aggiornare i riferimenti storici dentro `docs/specs/` e `docs/adr/` che nominano i vecchi package.
- Rimisurare le soglie di `internal/tray`, che non viene toccato.

## Further Notes

Lo split è meccanico ma largo: tocca ogni import dei due package e richiede di esportare gli identificatori che attraversano un confine di package. L'esportazione resta confinata dentro `internal/`, quindi non allarga la superficie pubblica del modulo.

L'ordine di lavoro è per gruppi, con build e test verdi a ogni passo: prima `compose` (`model`, `docker`, `store`, `parse`), poi `gui` (`design` e `view`, poi `activity`/`logs`/`usage`, poi `project`/`forms`/`traymenu`), infine configurazioni Gremlins, misura e README. La mappa delle dipendenze dei test ha deciso anche due confini: il foglio di stile con `ServiceColor` in `design`, e i testi condivisi in `view`. Ha inoltre mostrato tre cicli (`view` ↔ `logs`, `view` ↔ `forms`, `view` ↔ `project`) che la regola del package foglia risolve, e ha reso necessario `guitest` per gli helper usati da test che finiscono in package diversi.

La decisione e le alternative scartate sono nell'ADR 0006.
