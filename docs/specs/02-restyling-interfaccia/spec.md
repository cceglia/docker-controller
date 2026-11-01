# Restyling dell'interfaccia

**Status:** resolved

## Problem Statement

La finestra di DockerController è una colonna centrata di card larga al massimo 56rem, in una finestra da 1000×700: i Progetti, il dettaglio e il pannello Attività si susseguono verticalmente, e aprire un Progetto fa sparire la lista, con un "← Progetti" per tornare. Ogni riga della lista porta cinque pulsanti sempre visibili; lo stato è una pillola colorata che, ripetuta su ogni riga, rende la lista un semaforo. L'Attività in corso e i Log finiscono sotto tutto e si contendono lo spazio con i Servizi, così un log affollato vive in 20rem di altezza.

L'aspetto non è deciso: `system-ui`, un blu d'accento usato quasi solo per l'hover del bordo, la stessa ombra su ogni superficie, etichette in maiuscoletto spaziato. Somiglia a qualunque dashboard generata, non a uno strumento che si guarda tutto il giorno.

C'è anche un'incongruenza nei numeri: la barra in cima mostra CPU, memoria, rete e disco di Docker come valori **istantanei** (byte al secondo fra due fotogrammi), mentre il blocco di ogni Servizio mostra rete e disco **totali dall'avvio dei Container**. Due forme diverse per la stessa misura nella stessa finestra, e la seconda contraddice il glossario, che definisce l'Utilizzo risorse come ciò che è consumato *in questo momento*.

## Solution

Un guscio a due pannelli: i Progetti in una colonna a sinistra, il Progetto aperto su una tela a destra, con la striscia dell'Utilizzo risorse globale in cima e un cassetto in fondo che tiene Attività e Log, separati dai Servizi. Le azioni di Progetto salgono nell'intestazione del dettaglio, con una sola azione primaria che cambia secondo lo stato; le distruttive stanno dietro un menu.

La direzione visiva è **"Achromatic ledger"**, servita dalla skill `design-taste-frontend` (`draw.mjs --app --kind devtools`): il chrome non ha colore, le superfici sono piatte, la separazione è una hairline o un gradino di tono, il colore resta sui dati con un solo hue. Tipografia Geist, cifre e identificatori in Geist Mono, font incorporati nell'app perché funziona offline. Tema scuro come riferimento — sul desktop è attivo BreezeDark — e chiaro derivato.

Un elemento caratteristico costruito dai dati veri dell'app: la **striscia dei Container**, una tacca per Container accanto al nome del Progetto e del Servizio, così il rapporto "3 su 5" si legge come una forma.

Nell'Utilizzo risorse, rete e disco di ogni Servizio diventano valori istantanei come quelli della barra globale, calcolati per Container e sommati per Servizio: una sola forma di misura in tutta la finestra.

## User Stories

### Guscio e navigazione

1. Come utente, voglio i Progetti in una colonna a sinistra e il Progetto aperto a destra, così da passare da un Progetto all'altro senza perdere il contesto.
2. Come utente, voglio che la riga del Progetto aperto sia evidenziata, così da sapere sempre dove sono.
3. Come utente, voglio i Progetti raggruppati per Gruppo anche nella colonna, così da ritrovare lavoro e personale come nella lista attuale.
4. Come utente, voglio vedere nella stessa riga il Nome visualizzato e il nome compose, così da riconoscere il Progetto anche dopo averlo rinominato.
5. Come utente, voglio vedere in ogni riga lo stato, il numero di Container e la striscia dei Container, così da capire cosa gira senza aprire il dettaglio.
6. Come utente, voglio che la finestra si apra a 1120×760, così da avere spazio per due pannelli senza coprire il desktop.
7. Come utente, voglio che sotto una certa larghezza la colonna si riduca alle sole tacche di stato, così da poter stringere la finestra senza perdere il dettaglio.
8. Come utente, voglio "Impostazioni" sempre raggiungibile in fondo alla colonna, così da non dover tornare alla lista.
9. Come utente, voglio "Aggiungi progetto" in cima alla colonna, così da poter aggiungere un Progetto da qualunque schermata.
10. Come utente, voglio che aprire un Progetto dalla colonna chiuda l'eventuale schermata Impostazioni, così da non restare su una schermata che non risponde ai clic.

### Aspetto e leggibilità

11. Come utente, voglio che l'app segua il tema del desktop, con lo scuro curato quanto il chiaro, così da non avere una finestra fuori posto la sera.
12. Come utente, voglio superfici piatte separate da linee sottili invece di card con ombre, così che la gerarchia si legga dai contenuti e non dagli effetti.
13. Come utente, voglio un solo colore "di dati" usato una volta sola nella schermata, così che l'occhio sappia dove guardare.
14. Come utente, voglio i numeri in un carattere a cifre fisse, così che non ballino a ogni aggiornamento della misura.
15. Come utente, voglio che i font siano incorporati nell'app, così che funzioni identica senza rete.
16. Come utente, voglio che ogni testo rispetti un contrasto leggibile, così da poter usare l'app anche con una vista non perfetta.
17. Come utente, voglio vedere dove si trova il focus da tastiera, così da poter usare l'app senza mouse.
18. Come utente, voglio che le animazioni si fermino se il sistema chiede movimento ridotto.
19. Come utente, voglio un marchio disegnato per questa app, così da riconoscerla nella barra delle applicazioni e nella finestra.
20. Come utente, voglio uno stato che sia un punto colorato più il testo, invece di una pillola su ogni riga, così che la lista non sembri un semaforo.
21. Come utente, voglio la pillola colorata solo dove c'è un problema vero — Container `unhealthy`, File compose mancante, errore — così che il colore segnali qualcosa.
22. Come utente, voglio etichette in minuscolo normale invece che in maiuscoletto spaziato, così che l'interfaccia parli la stessa lingua dei suoi contenuti.

### Dettaglio del Progetto

23. Come utente, voglio eseguire Up, Stop, Restart, Pull e Build dall'intestazione del Progetto, così da poter agire senza tornare alla lista.
24. Come utente, voglio che una sola azione sia quella primaria e che cambi secondo lo stato del Progetto — Up se Fermo o Rimosso, Stop se Attivo, Restart se Parziale — così che il prossimo passo ovvio sia a un clic.
25. Come utente, voglio che le altre azioni restino disponibili ma discrete, così da non avere sei pulsanti dello stesso peso.
26. Come utente, voglio Down, Down con volumi e Rimuovi dalla lista dietro un menu, così che le azioni distruttive non stiano accanto a Up.
27. Come utente, voglio che le conferme di oggi restino identiche, compresa la richiesta di digitare il nome del Progetto per il Down con volumi.
28. Come utente, voglio vedere Cartella, Nome compose e Gruppo in celle leggibili, così da riconoscere il Progetto prima di agire.
29. Come utente, voglio aprire la cartella e il terminale del Progetto a un clic da quelle celle, così da non cercarle nel menu.
30. Come utente, voglio modificare Nome visualizzato e Gruppo dal dettaglio come oggi, con gli stessi campi e la stessa validazione.
31. Come utente, voglio i Profili come caselle con la spiegazione di cosa fanno, così da sapere che cambiarli non avvia né ferma nulla.
32. Come utente, voglio ogni Servizio in un blocco piatto con nome, striscia dei Container e azioni, così da vedere la struttura del Progetto.
33. Come utente, voglio ogni Container con punto di stato, nome, immagine, stato, health e porte, così da leggere una riga per Container invece di un ammasso di testo.
34. Come utente, voglio aprire una porta pubblicata con un clic, come oggi.
35. Come utente, voglio vedere in evidenza un'Attività in corso sul Progetto, così da sapere perché i pulsanti sono disabilitati.
36. Come utente, voglio un avviso con il pulsante di rimozione quando i File compose non esistono più, così da poter sistemare la lista subito.
37. Come utente, voglio che a finestra aperta senza Progetto selezionato la tela non sia vuota: un riepilogo dei Progetti e l'invito a sceglierne uno, così da sapere dove sono e cosa fare.

### Utilizzo risorse

38. Come utente, voglio che rete e disco di ogni Servizio siano valori istantanei al secondo, così da vedere il traffico di adesso e non la somma da quando i Container sono partiti.
39. Come utente, voglio che quei valori abbiano la stessa forma di quelli della barra in alto, così da poterli confrontare senza tradurre unità.
40. Come utente, voglio che CPU e memoria restino quelli istantanei di oggi.
41. Come utente, voglio distinguere "in attesa della prima misura" da "nessun Container in esecuzione", così da non leggere uno zero dove non c'è misura.
42. Come utente, voglio che la barra in alto sia dichiaratamente globale — etichettata Docker — con Container, CPU, memoria, rete e disco, così da non confonderla con i numeri del Progetto aperto.
43. Come utente, voglio un indicatore sottile di CPU e memoria con il numero accanto, così da vedere una proporzione senza dipendere da un grafico.

### Attività e Log

44. Come utente, voglio un cassetto in fondo alla tela con due schede, Attività e Log, così che l'output non si mangi lo spazio dei Servizi.
45. Come utente, voglio che il cassetto si apra da solo quando parte un'Azione, così da vedere subito cosa sta succedendo.
46. Come utente, voglio che "Mostra log" apra il cassetto sulla scheda Log, così da arrivare dove volevo.
47. Come utente, voglio che la scheda Log esista solo quando c'è uno stream aperto, così da non avere una scheda vuota.
48. Come utente, voglio conservare filtro Servizio, ricerca, timestamp, pausa e chiusura dei log, con l'area del log che occupa tutta la larghezza del cassetto.
49. Come utente, voglio chiudere il cassetto senza fermare i log, e fermare i log quando cambio Progetto o vado in Impostazioni, come oggi.
50. Come utente, voglio che "Vedi output" nel messaggio d'errore apra il cassetto sull'Attività giusta.
51. Come utente, voglio che la cronologia delle Attività resti selezionabile, con l'output che segue l'ultima riga.

### Impostazioni, dialoghi e messaggi

52. Come utente, voglio le Impostazioni a due colonne — etichetta e spiegazione a sinistra, controllo a destra — così da leggere cosa cambia prima di cambiarlo.
53. Come utente, voglio le stesse impostazioni di oggi: comportamento alla chiusura e comando del terminale, con Salva, esito ed errore.
54. Come utente, voglio i dialoghi (aggiunta Progetto, chiusura finestra, conferme) coerenti con il resto della finestra e con un solo pulsante primario.
55. Come utente, voglio che il messaggio d'errore resti leggibile e non sparisca mentre lo leggo, come già fa oggi.
56. Come utente, voglio le stesse funzioni di prima: nessuna azione, nessuna conferma, nessuna scorciatoia deve cambiare posto fino a sparire o cambiare effetto.

### Invarianti

57. Come utente, voglio stringhe in italiano con la maiuscola solo a inizio frase, così che l'interfaccia sia coerente.
58. Come utente, voglio che cambiare aspetto non cambi cosa fa l'app: gli stessi eventi, gli stessi binding, gli stessi comportamenti.

## Implementation Decisions

- **Direzione visiva** "Achromatic ledger" adottata dal foglio servito, con tre deviazioni dichiarate e motivate: tema scuro come riferimento (BreezeDark è il tema del desktop), font incorporati invece che da CDN (l'app è offline), finestra 1120×760 invece dei 1440 su cui il foglio è misurato.
- **Token**: un unico foglio di stile con i ruoli `ground`, `surface`, `raised`, `active`, `border`, `rule`, `control`, `text`, `muted`, `faint`, `primary`/`primary-ink`, `accent`, `positive`, `negative`, `warning`, i quattro gruppi di token delle pillole, raggi 8/6, altezze 44 (riga) e 32 (controllo), spaziatura 12/24, scala 24/16/30/14/12. Scuro: fondo `#17181a`, superficie `#1e2023`, hairline `#2c2f33`, inchiostro `#e8e9ea`, hue dei dati `#5b9dff`. Chiaro: fondo `#f8f8f8`, superficie `#ffffff`, hairline `#e6e6e6`, inchiostro `#0a0a0a`, hue dei dati `#1d5fbf`. Nessun componente dichiara colori propri: i valori stanno in un solo posto.
- **Tipografia**: Geist per il testo, Geist Mono per ogni cifra che cambia, per porte, nomi compose e percorsi; due file `.woff2` del sottoinsieme latino incorporati con la licenza OFL accanto. Cifre tabulari su ogni numero.
- **Tema**: segue `prefers-color-scheme`; lo scuro è il riferimento e il chiaro è derivato dagli stessi ruoli, verificato a contrasto.
- **Guscio**: colonna da 248px con marchio, "Aggiungi progetto", una voce "Progetti" che riporta al riepilogo, Progetti raggruppati e Impostazioni in fondo; tela con striscia di utilizzo, titolo, contenuto scorrevole e cassetto. Sotto i 900px la colonna si riduce a 64px con le sole tacche.
- **Lista Progetti**: diventa navigazione; perde i pulsanti azione, che si spostano nel dettaglio. Ogni riga: punto di stato, nome con nome compose sotto, striscia dei Container.
- **Striscia dei Container**: componente che riceve i conteggi oppure gli stati per Container; piena per in esecuzione, vuota per fermo, in `negative` per `unhealthy`/`dead`, con "+n" oltre dodici. È l'elemento caratteristico del prodotto.
- **Marchio**: tre blocchi impilati, quello superiore nel hue dei dati; SVG disegnato, nessun quadrato con una lettera.
- **Dettaglio**: intestazione con nome, punto di stato, striscia, nome compose e cartella; azioni con primaria secondo lo stato; menu `…` per Down, Down con volumi e Rimuovi dalla lista, con le conferme esistenti. Celle per Cartella, Nome compose e Gruppo con i comandi Cartella, Terminale e Modifica. Profili come riga di caselle. Servizi come blocchi piatti con azioni (Start, Stop, Restart e Down) e utilizzo risorse in celle divise.
- **Azioni di Progetto**: cinque azioni nell'intestazione; primaria quella adatta allo stato; le altre discrete; tutte disabilitate durante un'Attività o con Docker irraggiungibile.
- **Stato**: punto da 6px nel colore semantico più testo in inchiostro; pillola solo per problemi. I quattro stati del Progetto e gli stati dei Container usano la stessa grammatica.
- **Utilizzo risorse globale**: striscia dichiaratamente "Docker" con Container in esecuzione, CPU (indicatore nel hue dei dati), memoria, rete in ingresso/uscita e disco in lettura/scrittura. "In attesa della prima misura" e "non disponibile" sono stati distinti dai valori.
- **Utilizzo risorse per Servizio**: rete e disco diventano `Rates` istantanei, con la stessa formattazione della barra globale. CPU e memoria restano istantanee.
- **Modello dati**: il rilevatore dei delta, che oggi lavora per Container e produce solo il totale, restituisce i rate di ogni Container della fotogramma; il campionamento somma quei rate per il totale globale e per ciascun Servizio. `ServiceUsage` guadagna un campo `Rate` accanto a `Measures`; `UsageSample.Rate` resta la somma su tutte le righe, comprese quelle di Progetti non salvati. Un Container nuovo non ha rate (assente, non zero) e un contatore che cala non produce un rate negativo.
- **Store del frontend**: il cassetto (aperto/chiuso, scheda attiva) è stato di interfaccia nello store, perché lo aprono tre punti diversi — avvio di un'Attività, apertura dei log, "Vedi output" del messaggio d'errore — e nessuno di essi è un genitore comune dei componenti coinvolti. L'Utilizzo risorse per Servizio espone misure e rate insieme.
- **Attività e Log**: due componenti esistenti, riportati dentro un cassetto alto 220px con schede a sottolineatura; l'area del log si adatta all'altezza del cassetto invece dei 20rem fissi di oggi.
- **Componenti di appoggio**: marchio, striscia dei Container e indicatore di proporzione sono componenti condivisi, così che la stessa forma non venga riscritta in tre punti.
- **Finestra**: 1120×760 all'avvio, ridimensionabile, nessun vincolo minimo imposto.
- **Stringhe**: stesse informazioni, sentence case. Nessun termine nuovo di dominio; il vocabolario resta quello del glossario (Progetto, Servizio, Container, Attività, Azione, Utilizzo risorse).

## Testing Decisions

Un buon test guarda il comportamento esterno e non l'implementazione: per il backend significa osservare i campioni che `Manager.Usage` consegna quando il port Docker riceve una sequenza nota di fotogrammi di `docker stats`; non significa testare la funzione privata che calcola i delta, né il numero di chiamate interne.

- **Un solo seam, quello che esiste già**: il port `Docker` con il fake usato dai test di Utilizzo risorse e i fixture di fotogrammi reali già presenti. Il test nuovo si aggancia a `Manager.Usage` e legge `UsageSample`/`ServiceUsage`, esattamente come i test di oggi leggono i totali globali e per Servizio. Non serve un seam nuovo.
- **Casi da coprire, sul seam esistente**:
  - due fotogrammi con delta noti producono rate per Servizio, somma delle sue repliche;
  - un Container che si riavvia (contatori che calano) dà rate zero, mai negativo;
  - un Container alla sua prima fotogramma non ha rate;
  - un Servizio che compare in due fotogrammi mantiene il rate anche quando il totale globale include Container di Progetti non salvati;
  - il totale globale resta la somma di tutte le righe, comprese quelle non attribuite a un Progetto salvato;
  - nessun Container in esecuzione resta un campione vuoto, non un errore.
- **Prior art**: i test del pacchetto `compose` che usano il fake Docker e i fixture di `docker stats`, e i test esistenti del rilevatore di rate, che fissano già il comportamento su contatori che calano. Il nuovo test estende quel file invece di aprire una seconda strada.
- **Frontend**: oggi non esiste un test runner e questa spec non ne introduce uno. Il seam scelto è **`svelte-check`**: controllo dei tipi e della struttura dei componenti, più la build di produzione, più la revisione a schermo delle schermate nei due temi a 1120×760 e a 1440×900. Non si aggiungono dipendenze di test. Il comportamento dei componenti è coperto indirettamente dal backend, che è l'unico punto in cui una modifica può cambiare dei numeri.
- **Decisione sul frontend**: `svelte-check` è la verifica del frontend. Ogni ticket che tocca componenti chiude con `svelte-check` e build verdi; nessun harness di rendering, nessun test di componente.

## Out of Scope

- Menu del tray, icone del tray e notifiche desktop.
- Nuove Azioni, nuovi comandi, nuove impostazioni.
- Modifiche al contratto dei binding oltre al campo dei rate per Servizio; nessun cambio di eventi o di forma dello Snapshot.
- Internazionalizzazione: l'interfaccia resta in italiano.
- Disposizione per telefono: l'app è desktop; la finestra stretta è l'unico caso previsto.
- Cornice della finestra, title bar personalizzata, schermo intero.
- Animazioni d'ingresso e transizioni fra schermate.
- Politica di ritenzione dei log (resta il massimo di 5000 righe).
- Test del frontend oltre a `svelte-check` e alla build: nessun harness di rendering, nessun test di componente.
- Modifiche a `GLOSSARY.md` e nuovi ADR: i nomi nuovi sono elementi d'interfaccia, non vocabolario di dominio, e la scelta è reversibile.

## Further Notes

- Il foglio è stato servito da `draw.mjs --app --kind devtools` con seed `458267769`: la direzione, i valori dei token e la struttura del guscio vengono da lì, non da preferenze personali. Dove il foglio parla di uno schermo a 1440px, la geometria è stata adattata a una finestra desktop da 1000-1120px, senza cambiare i ruoli dei token.
- La modifica non committata già presente in `ErrorToast.svelte` (timer che si mette in pausa su hover e focus, pulsante "Chiudi") fa parte del lavoro e va conservata: il restyling cambia solo l'aspetto, non quella logica.
- La rinomina di `CLAUDE.md` in `AGENTS.md` è già stata fatta ed è indipendente da questa spec.
- L'Utilizzo risorse per Servizio, mostrato come totale dall'avvio, contraddiceva il glossario prima ancora che l'aspetto: la correzione non introduce un termine nuovo, riallinea l'implementazione alla definizione già scritta.
