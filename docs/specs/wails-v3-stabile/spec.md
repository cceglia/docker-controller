# Wails v3.0.0 stabile

**Status:** ready-for-agent

## Problem Statement

L'app gira su Wails v3.0.0-beta.34, una prerelease. Dal system tray su KDE/Linux due cose non funzionano come l'utente se le aspetta: passando il mouse sull'icona compare il nome "Wails" invece di "Docker Controller", e il doppio click sull'icona non fa nulla, così la finestra si può riportare in primo piano solo dal menu. Inoltre il framework è fermo a una beta: la v3 definitiva (v3.0.0, 26 ottobre 2026) è uscita e modulo Go, CLI e runtime del frontend vanno tenuti sulla stessa versione.

## Solution

Aggiornare l'app a Wails v3.0.0 stabile, allineando modulo Go, CLI e runtime JavaScript del frontend e rigenerando i binding. Sul system tray di Linux il nome dell'app diventa l'etichetta dell'icona, così l'hover mostra "Docker Controller"; il click sinistro riporta la finestra in primo piano; il tasto destro continua ad aprire il menu. Il doppio click resta l'azione che riporta la finestra su Windows. Le particolarità di Linux finiscono in un ADR.

## User Stories

### Nome e click sulla tray

1. Come utente, voglio che passando il mouse sull'icona della tray compaia "Docker Controller", così da riconoscere subito l'app e non leggere "Wails".
2. Come utente, voglio che il nome nell'hover sia lo stesso della finestra e delle finestre di dialogo, così da non avere due nomi per la stessa app.
3. Come utente, voglio riportare la finestra in primo piano con un solo click sull'icona, così da non dover aprire il menu.
4. Come utente, voglio che la finestra riportata in primo piano sia anche deminimizzata e a fuoco, così da poterci lavorare subito.
5. Come utente, voglio che il tasto destro continui ad aprire il menu della tray, così da raggiungere Progetti, Servizi e Azioni come prima.
6. Come utente, voglio che il click sinistro non apra anche il menu, così che i due gesti restino distinti.
7. Come utente, voglio che l'icona della tray continui a cambiare con lo stato dei Progetti, così da avere il colpo d'occhio di prima.

### Nessuna regressione dalla tray

8. Come utente, voglio che il menu della tray mantenga la struttura per Gruppo e Progetto, con i Servizi annidati, così da non perdere nulla nell'aggiornamento.
9. Come utente, voglio che le Azioni lanciate dalla tray continuino a funzionare, comprese le conferme delle Azioni distruttive.
10. Come utente, voglio che le notifiche desktop continuino ad arrivare con lo stesso titolo, così da riconoscere l'app.
11. Come utente, voglio che il menu si aggiorni a ogni snapshot anche dopo l'aggiornamento, così da vedere sempre lo stato corrente.
12. Come utente, voglio che la finestra continui a nascondersi nella tray alla chiusura, così da mantenere il comportamento attuale.

### Aggiornamento del framework

13. Come utente, voglio che l'app giri sulla v3 definitiva e non su una beta, così da avere le correzioni e il supporto della versione stabile.
14. Come utente, voglio che l'aggiornamento non cambi il modo in cui uso l'app, così da non dover reimparare nulla.
15. Come utente, voglio che l'app si compili con Go 1.27.3, così da usare il toolchain aggiornato del sistema.
16. Come manutentore, voglio che modulo Go, CLI e runtime del frontend siano alla stessa versione, così che i binding siano generati dalla stessa versione che li consuma.
17. Come manutentore, voglio che i binding TypeScript siano rigenerati dopo l'aggiornamento, così che il frontend non usi tipi vecchi.
18. Come manutentore, voglio che il build resti su GTK3, così da non richiedere GTK4 e WebKitGTK 6.0 che non sono installati.
19. Come manutentore, voglio che l'aggiornamento passi la pipeline di verifica esistente senza modificare i test, così da sapere che non ho rotto nulla.
20. Come manutentore, voglio un ADR che spieghi perché su Linux l'etichetta della tray porta il nome dell'app e perché il click è singolo, così che nessuno "ripulisca" quel codice credendolo inutile.
21. Come manutentore, voglio che la scelta di restare sulla stabile e non sulle nightly sia scritta, così da non introdurre una dipendenza da build non rilasciate.

## Implementation Decisions

- Il framework passa dalla beta.34 alla v3.0.0 stabile. Non servono modifiche al codice per compilare: l'API usata (applicazione, finestre, tray, menu, dialoghi, notifiche, eventi) è invariata fra beta.34 e v3.0.0, le differenze sono correzioni.
- Modulo Go, CLI `wails3` e runtime JavaScript del frontend vengono allineati alla stessa versione; i binding TypeScript sono rigenerati con il CLI aggiornato.
- L'adattatore della tray resta l'unico punto che parla con il system tray di Wails. Non cambiano il modulo che costruisce il menu, gli snapshot, né il frontend.
- Su Linux l'etichetta della tray porta il nome dell'app: in v3.0.0 `SetTooltip` non ha effetto su Linux e un'etichetta non impostata vale "Wails". Il tooltip resta impostato per le altre piattaforme; l'etichetta è solo Linux perché su macOS mostrerebbe testo accanto all'icona, dove i tooltip non sono supportati.
- Su Linux il click sinistro riporta la finestra in primo piano. Il protocollo delle icone di stato di Linux non ha un evento di doppio click: arriva solo il click singolo. Il doppio click resta registrato per Windows, dove Wails lo consegna davvero.
- L'azione di riportare la finestra è quella già usata dalla voce "Mostra finestra" del menu: mostra, deminimizza, dà il fuoco.
- Il build resta con il tag GTK3: la v3.0.0 usa GTK4 e WebKitGTK 6.0 per impostazione predefinita, ma il supporto GTK3 resta fino alla v3.1 e sul computer è installato solo GTK3/WebKitGTK 4.1.
- Il toolchain Go di riferimento è 1.27.3, installato via snap, che soddisfa la raccomandazione di v3.0.0.
- Un nuovo ADR registra il passaggio alla stabile e le due particolarità della tray su Linux; l'ADR esistente sulla scelta della beta resta come storia.

## Testing Decisions

- Un buon test verifica il comportamento osservabile, non i dettagli implementativi. Qui il comportamento osservabile (testo dell'hover, effetto del click) vive nella sessione desktop, non nel codice Go: l'adattatore della tray è sottile e inoltra a Wails.
- Nessuna nuova seam. La seam più alta già esistente è la pipeline di verifica del progetto (vet, test Go, binding, typecheck e build del frontend): l'aggiornamento deve passarla senza modificare i test. I test esistenti del menu, delle icone e delle notifiche restano invariati e devono continuare a passare.
- Verifica manuale su KDE, l'unico posto dove si vede il comportamento della tray: hover con "Docker Controller", click sinistro che riporta la finestra, tasto destro che apre il menu, voci del menu e notifiche ancora funzionanti, icona che segue lo stato dei Progetti.

## Out of Scope

- Passare a GTK4 e WebKitGTK 6.0.
- Usare le nightly v3.1.0, che implementano `SetTooltip` su Linux.
- Cambiare la struttura del menu della tray o aggiungere Azioni.
- Packaging e verifica su Windows e macOS.
- Installare o aggiornare il toolchain Go: già fatto con lo snap 1.27.3.
- Modificare l'ADR sulla scelta della beta.

## Further Notes

- La v3.0.0 è uscita il 26 ottobre 2026. Rispetto alla beta.34 porta correzioni (crash del menu su Windows, dark mode, updater, `wails3 update cli`), nessuna rottura di API per questa app.
- Il fix upstream di `SetTooltip` su Linux è stato mergiato il 27 ottobre, un giorno dopo la release: esiste solo nelle nightly 3.1.0. Quando la v3.1 diventerà stabile andrà rivalutato, ma l'etichetta resterà comunque necessaria per il nome nell'hover.
- La verifica manuale va fatta dopo il build, perché la tray vive nella sessione grafica dell'utente.
