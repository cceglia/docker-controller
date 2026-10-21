# DockerController v1

**Status:** ready-for-agent

## Problem Statement

Sul computer convivono diversi Progetti docker-compose (`sunlinkdocker` con 14 Container, alcune VPN, progetti di lavoro e personali), sparsi in cartelle diverse. Per sapere cosa è in esecuzione, avviare o fermare un Progetto, leggere i log o fare pull/build bisogna aprire un terminale, ricordare dove si trova ogni File compose e digitare i comandi a mano. Dopo un Down il Progetto sparisce da `docker compose ls`, quindi bisogna ricordarsi che esiste. Non c'è un colpo d'occhio sullo stato di tutto, e non c'è un modo rapido di agire senza passare dal terminale.

## Solution

Un'app desktop per KDE/Linux che mostra in un'unica finestra tutti i Progetti salvati, con il loro stato aggiornato in tempo reale, i Container raggruppati per Servizio, i log in streaming e tutte le Azioni (Up, Stop, Restart, Pull, Build, Down, Down con volumi), sia a livello di Progetto sia di Servizio. L'app vive nel system tray: da lì, con menu annidati per Gruppo e Progetto, si vedono gli stati e si eseguono le Azioni non distruttive senza aprire la finestra. I Progetti vengono scoperti in automatico da Docker e restano in lista anche dopo un Down. Si possono aggiungere a mano Progetti mai avviati. Ogni Progetto può avere un Nome visualizzato, un Gruppo e dei Profili attivi. In cima alla finestra c'è l'Utilizzo risorse globale di Docker, e ogni Servizio mostra il proprio.

## User Stories

### Lista e scoperta
1. Come utente, voglio vedere in una sola finestra tutti i miei Progetti, così da non dover ricordare dove si trova ogni File compose.
2. Come utente, voglio che i Progetti che Docker conosce compaiano automaticamente, così da non doverli registrare a mano.
3. Come utente, voglio che un Progetto resti in lista dopo un Down, così da poterlo riavviare con un clic.
4. Come utente, voglio vedere per ogni Progetto lo stato (Attivo, Parziale, Fermo, Rimosso) e il numero di Container in esecuzione sul totale, così da capire subito cosa sta girando.
5. Come utente, voglio che lo stato si aggiorni appena un Container cambia, anche quando il cambio arriva dal terminale, così da potermi fidare di quello che vedo.
6. Come utente, voglio vedere la cartella di ogni Progetto, così da riconoscere Progetti con nomi simili.
7. Come utente, voglio che la lista sia raggruppata per Gruppo, con i Progetti senza Gruppo in una sezione a parte, così da separare lavoro e personale.
8. Come utente, voglio aggiungere a mano un Progetto scegliendo il suo File compose principale, così da gestire anche Progetti mai avviati.
9. Come utente, quando aggiungo un Progetto, voglio che il nome compose venga proposto come farebbe compose (dal `name:` del file oppure dal nome della cartella) e che sia modificabile.
10. Come utente, voglio aggiungere ulteriori File compose (es. un override) a un Progetto, così che le Azioni li usino tutti.
11. Come utente, voglio un messaggio chiaro se provo ad aggiungere un Progetto con un nome già salvato, così da non creare duplicati.
12. Come utente, voglio vedere l'avviso "File compose mancante" quando i file di un Progetto non esistono più, così da sapere che posso toglierlo.
13. Come utente, voglio usare "Rimuovi dalla lista" su un Progetto Rimosso, così da eliminare i Progetti che non esistono più.
14. Come utente, voglio che "Rimuovi dalla lista" non sia disponibile per un Progetto che ha ancora Container, così che non ricompaia subito dopo.

### Personalizzazione del Progetto
15. Come utente, voglio dare a un Progetto un Nome visualizzato, così da leggere "VPN Stellantis" invece di `stellantisvpn`.
16. Come utente, voglio assegnare un Progetto a un Gruppo, scegliendone uno esistente o creandone uno nuovo, così da organizzare lista e tray.
17. Come utente, voglio togliere Nome visualizzato o Gruppo, così da tornare al comportamento predefinito.

### Azioni sul Progetto
18. Come utente, voglio eseguire Up, Stop, Restart, Pull e Build su un Progetto con un clic, così da non usare il terminale.
19. Come utente, voglio eseguire Down su un Progetto dopo una conferma in un modal, così da non rimuovere Container per errore.
20. Come utente, voglio un pulsante separato e rosso per Down con volumi, che richieda di digitare il nome del Progetto nel modal di conferma, così da non cancellare dati per sbaglio.
21. Come utente, voglio che durante un'Attività i pulsanti di quel Progetto siano disabilitati e che sia visibile "⏳ <Azione> in corso", così da non lanciare comandi in conflitto.
22. Come utente, voglio poter lavorare su Progetti diversi in parallelo, così da non aspettare un Build per fermare un'altra cosa.
23. Come utente, voglio vedere in tempo reale l'output di ogni Attività in un pannello Attività, così da seguire Pull e Build lunghi.
24. Come utente, voglio un toast con l'errore quando un'Attività fallisce, con un link all'output completo, così da capire cosa è andato storto.
25. Come utente, voglio consultare lo storico delle Attività della sessione corrente, così da rivedere un output passato.

### Servizi e Container
26. Come utente, voglio vedere nel dettaglio di un Progetto i Container raggruppati per Servizio, con immagine, stato, health e porte, così da capire cosa gira.
27. Come utente, voglio eseguire Start, Stop, Restart su un singolo Servizio, così da riavviare solo quello che serve.
28. Come utente, voglio che un Servizio con repliche mostri tutti i suoi Container, ma che le Azioni valgano per il Servizio.
29. Come utente, voglio cliccare su una porta pubblicata e aprire `http://localhost:<porta>` nel browser.
30. Come utente, voglio un pulsante "Apri cartella" che apra la cartella del Progetto nel file manager.
31. Come utente, voglio un pulsante "Apri terminale" che apra un terminale nella cartella del Progetto, con un comando configurabile.

### Profili
32. Come utente, voglio vedere i Profili disponibili di un Progetto come interruttori, così da scegliere i Profili attivi.
33. Come utente, voglio che i Profili attivi siano ricordati, così da non doverli riselezionare.
34. Come utente, voglio che Up, Stop, Restart, Pull e Build usino i Profili attivi.
35. Come utente, voglio che Down e Down con volumi agiscano su tutti i Profili, così da non lasciare Container orfani.
36. Come utente, voglio che cambiare un Profilo non avvii né fermi nulla finché non lancio un'Azione.

### Log
37. Come utente, voglio aprire i log di un Progetto e vedere le ultime 500 righe seguite dallo streaming in tempo reale.
38. Come utente, voglio filtrare i log per Servizio, così da concentrarmi su un solo Servizio.
39. Come utente, voglio che ogni Servizio abbia un colore nei log, così da distinguere le righe.
40. Come utente, voglio mettere in pausa lo scorrimento, cercare nel testo e mostrare o nascondere i timestamp.
41. Come utente, voglio che il pannello non trattenga più di 5000 righe, così che l'app resti reattiva.
42. Come utente, voglio che lo streaming dei log si fermi quando chiudo il pannello, così da non sprecare risorse.

### Tray
43. Come utente, voglio un'icona nel system tray che rifletta se Docker è raggiungibile.
44. Come utente, voglio un menu tray annidato Gruppo ▸ Progetto, con i Progetti senza Gruppo al primo livello, così che il menu resti compatto.
45. Come utente, voglio vedere nel menu tray lo stato di ogni Progetto (● Attivo, ◐ Parziale, ○ Fermo/Rimosso).
46. Come utente, voglio eseguire Up, Stop e Restart su un Progetto dal tray.
47. Come utente, voglio un sottomenu Servizi ▸ Servizio ▸ Start/Stop/Restart nel tray.
48. Come utente, voglio che Down, Down con volumi, Pull e Build non siano disponibili nel tray, così che non compaiano mai modal da lì.
49. Come utente, voglio che nel tray le voci di un Progetto con un'Attività in corso siano disabilitate.
50. Come utente, voglio una notifica desktop quando finisce un'Attività lanciata dal tray, o quando un'Attività fallisce, così da conoscere l'esito anche a finestra nascosta.
51. Come utente, voglio le voci "Mostra finestra" ed "Esci" nel tray.

### Finestra e impostazioni
52. Come utente, voglio scegliere cosa succede alla chiusura della finestra: Nascondi nel tray (default), Esci o Chiedi ogni volta.
53. Come utente, con "Chiedi ogni volta" voglio una casella "ricorda la scelta" nel dialogo.
54. Come utente, voglio configurare il comando "Apri terminale", con default `konsole --workdir {dir}`.
55. Come utente, voglio che le impostazioni e i dati dei Progetti sopravvivano ai riavvii dell'app.
56. Come utente, voglio un tema chiaro/scuro che segua quello di KDE.

### Docker non raggiungibile
57. Come utente, voglio un banner "Docker non raggiungibile" quando il demone non risponde, con la lista in grigio e le Azioni disabilitate.
58. Come utente, voglio che l'app torni operativa da sola quando Docker torna disponibile, riprovando ogni 5 secondi.

### Installazione
59. Come utente, voglio installare l'app con un solo comando, con voce nel menu KDE e icona.

### Utilizzo delle risorse
60. Come utente, voglio vedere in cima alla finestra l'utilizzo globale delle risorse da parte di Docker (CPU, memoria, rete, I/O disco di tutti i Container in esecuzione), così da capire subito quanto Docker sta pesando sul computer.
61. Come utente, voglio che CPU e memoria globali siano espresse anche in percentuale sul totale della macchina, così da confrontarle con le risorse disponibili.
62. Come utente, voglio vedere dentro ogni Servizio le risorse che usa (CPU, memoria, rete, I/O disco), sommate su tutti i suoi Container, così da individuare il Servizio che consuma di più.
63. Come utente, voglio che i valori si aggiornino in continuazione (circa ogni 2 secondi) senza rallentare l'app.
64. Come utente, voglio che un Servizio senza Container in esecuzione mostri "—" invece di zero, così da non confonderlo con un Servizio in esecuzione che non consuma nulla.
65. Come utente, voglio che con Docker non raggiungibile la barra globale mostri lo stato di errore invece di valori vecchi.

## Implementation Decisions

- **Stack**: Go + Wails v3 (versione fissata, vedi ADR-0001), frontend Svelte + TypeScript + Vite, SQLite tramite `modernc.org/sqlite`. Interfaccia solo in italiano.
- **Docker**: tutte le interazioni passano dalla CLI (`docker compose …`, `docker events`, vedi ADR-0002). Solo il Docker locale ufficiale, context di default, utente nel gruppo docker.
- **Modulo profondo `compose`**: contiene tutta la logica di dominio. La sua interface, l'unica superficie usata da finestra, tray e test, offre:
  - uno **snapshot** completo e già risolto dei Progetti salvati: nome, Nome visualizzato, Gruppo, cartella, File compose, File compose mancante, stato, Container per Servizio con stato/health/porte, Profili disponibili e attivi, Attività in corso, e se Docker è raggiungibile;
  - **Watch**: notifica un nuovo snapshot solo quando qualcosa cambia; finestra e tray si abbonano qui e nessuno dei due fa polling;
  - **un'unica operazione per eseguire un'Azione** su un Progetto o su un Servizio. Rifiuta di avviare una seconda Attività sullo stesso Progetto, restituisce un identificativo di Attività e consegna l'output riga per riga e l'esito finale;
  - **Logs**: streaming per Progetto ed eventuale Servizio, con righe già attribuite al Servizio, fino alla cancellazione;
  - aggiunta manuale di un Progetto, "Rimuovi dalla lista" (rifiutato se il Progetto non è Rimosso) e un'unica operazione per aggiornare le impostazioni del Progetto (Nome visualizzato, Gruppo, Profili attivi).
- **Dentro `compose`, nascosti**: costruzione di `docker compose -p <nome> -f … --project-directory …` con i Profili attivi (Down usa tutti i Profili); mapping Azione→argomenti (Build = `up -d --build`, Down con volumi = `down -v`); parsing JSON di `compose ls -a`, `compose ps -a`, `compose config --profiles`; salvataggio automatico dei Progetti scoperti; calcolo dello stato; ascolto di `docker events` con aggiornamento del solo Progetto interessato più un aggiornamento completo ogni 30 secondi; controllo di raggiungibilità di Docker ogni 5 secondi quando non risponde; diff degli snapshot.
- **Seam verso Docker**: un port con due operazioni, "esegui e restituisci l'output" e "esegui in streaming". Ha due adapter: quello di produzione basato su processi esterni e un fake per i test che risponde con output reali catturati e registra gli argomenti ricevuti.
- **Utilizzo delle risorse**: il modulo `compose` legge in streaming `docker stats --format json`, cioè tutti i Container in esecuzione, non solo quelli dei Progetti salvati. Le misure sono CPU, memoria usata, rete in/out e I/O disco. La CPU riportata da Docker (somma per core, può superare il 100%) viene normalizzata sul numero di CPU della macchina per la percentuale globale; la memoria globale viene confrontata con la memoria totale della macchina (`docker info`). Le misure per Servizio sono la somma dei suoi Container. L'utilizzo è esposto separatamente dallo snapshot dei Progetti, con un proprio flusso a cadenza di circa 2 secondi, così le notifiche di Watch (e la ricostruzione del tray) non avvengono a ogni campione. Lo streaming parte quando la finestra è visibile e si ferma quando è nascosta nel tray. Il tray non mostra risorse.
- **Persistenza**: un solo file SQLite in `~/.local/share/docker-controller/docker-controller.db`, creato al primo avvio, con migrazioni versionate. Tabelle concettuali: Progetti salvati (nome, File compose ordinati, cartella, Nome visualizzato, Gruppo), Profili attivi per Progetto, impostazioni dell'app (chiave/valore). Nei test SQLite è sostituibile in locale con un file temporaneo, senza bisogno di un port.
- **Impostazioni dell'app**: comportamento alla chiusura (`hide` | `quit` | `ask`, default `hide`) e comando terminale (default `konsole --workdir {dir}`). Le gestisce un piccolo modulo sullo stesso DB.
- **Adapter sottili**: binding Wails (traduce lo snapshot in un evento verso il frontend, espone le operazioni e mappa gli id dei flussi di log e di Attività sulla loro cancellazione), tray (ricostruisce il menu annidato a ogni snapshot e manda le notifiche desktop per le Attività lanciate da lì), wiring dell'app.
- **Frontend**: un solo modulo store parla con i binding: snapshot dei Progetti, lancio di Azioni con stato per Progetto, storico delle Attività della sessione, log con buffer di 5000 righe che si chiude all'unsubscribe. I componenti sono solo presentazione: lista raggruppata, dettaglio, Profili, log, Attività, modal di conferma, aggiunta Progetto, impostazioni.
- **Azioni distruttive**: Down, Down con volumi, Rimuovi dalla lista; sempre con un modal, e Down con volumi richiede di digitare il nome. Nessuna è disponibile nel tray.
- **Installazione**: un target make che compila e installa il binario, il file `.desktop` e l'icona in `~/.local`.

## Testing Decisions

- Un buon test agisce solo attraverso l'interface del modulo `compose` e verifica il comportamento osservabile (snapshot, argomenti passati a Docker, esiti, righe di log), mai lo stato interno. I test devono sopravvivere ai refactoring interni.
- **Modulo testato**: `compose`, con il fake Docker alimentato da output JSON reali catturati dai Progetti presenti sul computer (`sunlinkdocker`, `stellantisvpn`, `personalvpn`, …) e con SQLite su un file temporaneo. Casi principali:
  - merge tra scoperti e salvati senza duplicati, e Progetto che resta in lista dopo un Down;
  - calcolo dei quattro stati;
  - argomenti corretti per ogni Azione: `-f` dei Progetti aggiunti a mano, Profili attivi, tutti i Profili per Down, `-v` per Down con volumi;
  - rifiuto di una seconda Attività sullo stesso Progetto e parallelismo tra Progetti diversi;
  - errore con lo stderr di compose;
  - log divisi per Servizio;
  - Watch che non notifica se nulla cambia e notifica dopo un evento Docker o un'Azione;
  - "Rimuovi dalla lista" rifiutato su un Progetto non Rimosso;
  - File compose mancante;
  - persistenza di Nome visualizzato, Gruppo e Profili attivi dopo la riapertura sullo stesso DB;
  - Docker non raggiungibile e successivo ritorno;
  - utilizzo delle risorse: somma per Servizio con repliche, normalizzazione della CPU globale sul numero di CPU, Servizio fermo senza misure, parsing delle unità di `docker stats` (es. `1.5GiB`, `12.3kB`).
- Modulo impostazioni: un test di lettura/scrittura con i default su un DB temporaneo.
- Binding, tray e frontend sono adapter sottili: si verificano a mano con `wails3 dev` seguendo i criteri di accettazione dei ticket. Nessun test automatico della UI nella v1.
- Non c'è codice preesistente da prendere come modello: i test del modulo `compose` fanno da riferimento per i successivi.

## Out of Scope

- Docker remoto o context diversi da quello di default; Podman; accesso tramite sudo.
- Modifica dei File compose dall'app.
- Gestione di immagini, volumi e reti fuori dai Progetti; prune.
- Exec o shell dentro un Container.
- Storico e grafici dell'utilizzo delle risorse (solo valori correnti); utilizzo delle risorse nel tray.
- Avvio automatico al login.
- Storico persistente delle Attività e dei log.
- Più Gruppi per Progetto (tag).
- Traduzioni e infrastruttura i18n.
- Pacchetti AppImage o .deb.

## Further Notes

- Prerequisiti di build da installare a mano: `libgtk-3-dev`, `libwebkit2gtk-4.1-dev`, `build-essential`, `pkg-config`, la CLI `wails3`.
- Per i test di verifica manuale usare `stellantisvpn` (non critico), mai `sunlinkdocker` per Down con volumi.
- Vocabolario: `GLOSSARY.md`. Decisioni: `docs/adr/`.
