# Gestione dei Gruppi

**Status:** resolved

**Amends:** `docs/adr/0006-sottopackage-compose-e-gui.md` (le etichette riservate passano al modello)

## Problem Statement

I Gruppi esistono solo come nome scritto su un Progetto: nel campo "Gruppo" del
form del Progetto si può digitare un nome nuovo e il Gruppo nasce da solo, ma
non c'è modo di rinominarlo, eliminarlo o decidere in che ordine compaiono.
Per correggere il nome di un Gruppo bisogna aprire ogni Progetto che lo usa;
per eliminarlo bisogna svuotarlo Progetto per Progetto. Un Gruppo che non ha
ancora Progetti non può esistere, e l'ordine della lista e del tray è sempre
alfabetico.

L'utente che organizza i Progetti per contesti — lavoro, personale, clienti —
non ha una superficie dove gestire i Gruppi, e i nomi sbagliati o quasi
duplicati restano perché correggerli costa troppo.

## Solution

I Gruppi diventano entità salvate, con un nome e una posizione tutti loro: un
Gruppo esiste anche senza Progetti e vive indipendentemente da chi lo usa.

Dal form del Progetto un comando "Gruppi…" apre un dialogo dedicato: l'elenco
dei Gruppi con quanti Progetti contengono, e i comandi per crearne uno,
rinominarlo, eliminarlo e spostarlo su o giù. L'ordine scelto dall'utente è
quello della lista, del tray e del campo "Gruppo" del form; i Gruppi nuovi
finiscono in fondo e due Gruppi non possono avere lo stesso nome, nemmeno con
maiuscole diverse.

Il campo "Gruppo" del form resta modificabile: scrivere un nome non in elenco
crea il Gruppo al salvataggio, come oggi, ma ora è una scorciatoia dentro una
gestione che ha anche una porta esplicita.

Una rinomina su un nome già esistente unisce i due Gruppi, con conferma.
Sopravvive il Gruppo che aveva già il nome di arrivo: tiene il suo nome e il
suo posto relativo. L'eliminazione di un Gruppo non tocca i Progetti: i suoi
Progetti tornano "Senza Gruppo", che è una sezione della lista quando la lista
ha sezioni, e non un Gruppo.

## User Stories

1. Come utente, voglio che i Gruppi siano gestiti in un dialogo dedicato, così da non dover aprire un Progetto per ogni correzione.
2. Come utente, voglio creare un Gruppo dal dialogo anche senza Progetti, così da preparare l'organizzazione prima di assegnare.
3. Come utente, voglio creare un Gruppo scrivendo un nome nuovo nel campo del form, così da non dover passare dal dialogo quando ho già il Progetto aperto.
4. Come utente, voglio che il campo Gruppo proponga i Gruppi salvati nell'ordine che ho scelto, così da ritrovare lo stesso ordine della lista.
5. Come utente, voglio rinominare un Gruppo una volta sola e vedere il nuovo nome su tutti i suoi Progetti, così da correggere un refuso senza toccare ogni Progetto.
6. Come utente, voglio che rinominare un Gruppo sul nome di un altro chieda conferma e unisca i due, così da poter accorpare due Gruppi senza spostare i Progetti a mano.
7. Come utente, voglio che l'unione tenga il nome e il posto relativo del Gruppo che aveva già il nome di arrivo, così da sapere quale dei due resta.
8. Come utente, voglio eliminare un Gruppo con una conferma che dice quanti Progetti passeranno a Senza Gruppo, così da non farlo per sbaglio.
9. Come utente, voglio che eliminare un Gruppo non tocchi i Progetti né Docker, così da perdere solo l'organizzazione.
10. Come utente, voglio spostare un Gruppo su e giù con due comandi, così da decidere l'ordine di lista e tray.
11. Come utente, voglio che l'ordine dei Gruppi sopravviva alla chiusura dell'app, così da non rifarlo a ogni avvio.
12. Come utente, voglio che un Gruppo nuovo finisca in fondo all'elenco, così da trovarlo subito e spostarlo dove voglio.
13. Come utente, voglio che lista e tray mostrino i Gruppi nello stesso ordine, così da riconoscere la stessa organizzazione in entrambi.
14. Come utente, voglio che "Senza Gruppo" resti sempre in fondo e non sia gestibile, così da non confonderla con un Gruppo.
15. Come utente, voglio vedere i Gruppi vuoti nel dialogo e nel campo del form, ma non come intestazioni vuote nella lista né come sottomenu vuoti nel tray, così da poterli preparare senza sporcare la vista.
16. Come utente, voglio che un nome già usato, anche con maiuscole diverse, sia rifiutato nella creazione dal dialogo e mi dica quale Gruppo esiste, così da non creare quasi duplicati.
17. Come utente, voglio che nel form un nome che differisce solo per maiuscole assegni il Gruppo esistente invece di dare errore, così da non restare bloccato.
18. Come utente, voglio vedere il numero di Progetti di ogni Gruppo nel dialogo, così da sapere cosa sto per rinominare o eliminare.
19. Come utente, voglio che i comandi non disponibili siano spenti (nessuna selezione, primo o ultimo posto), così da non lanciare operazioni impossibili.
20. Come utente, voglio un messaggio in riga nel dialogo quando un nome è rifiutato, così da correggerlo senza chiudere nulla.
21. Come utente, voglio usare il dialogo solo con la tastiera, con ordine di tabulazione e scorciatoie per spostare, così da non dipendere dal mouse.
22. Come utente, voglio che una modifica fatta nel dialogo si veda subito nella lista e nel form senza riavviare, così da controllare il risultato.
23. Come utente, voglio che i Gruppi si possano gestire anche con Docker irraggiungibile, così da sistemare l'organizzazione in ogni momento.
24. Come utente con Gruppi già salvati, voglio ritrovarli tutti dopo l'aggiornamento, con un ordine alfabetico senza distinzione di maiuscole, così da non riconfigurare nulla.
25. Come utente con Gruppi che differiscono solo per maiuscole o per spazi iniziali e finali, voglio che l'aggiornamento li unisca invece di lasciare due Gruppi con lo stesso nome, così da rispettare la regola nuova.
26. Come utente, voglio che il dialogo si chiuda con "Chiudi" o Escape e restituisca il focus al comando "Gruppi…", così da riprendere da dove ero.

## Implementation Decisions

### Modello e persistenza

I Gruppi sono entità salvate, non più una stringa derivata dai Progetti. La
tabella nuova tiene nome e posizione; ogni Progetto punta al suo Gruppo con una
chiave esterna che, alla cancellazione del Gruppo, lascia il Progetto senza
Gruppo. La colonna del nome sul Progetto esce dallo schema.

```
groups(id, name, position)
  name UNIQUE (esatto)
projects.group_id → groups.id ON DELETE SET NULL
```

Il nome è l'identità del Gruppo: viene ripulito dagli spazi iniziali e finali e
confrontato senza distinzione tra maiuscole e minuscole. Gli spazi interni
fanno parte del nome. Il confronto è fatto dal codice, che conosce l'Unicode;
l'indice unico del database copre solo la grafia esatta, perché `NOCASE` di
SQLite ignora soltanto l'ASCII. Il confronto usa il folding semplice di
`strings.EqualFold`: «Straße» e «STRASSE» restano due Gruppi. La regola vive in un unico punto del modello,
che non importa altri package: normalizzazione, confronto, validazione. La
validazione riceve il nome, l'elenco dei nomi esistenti, il modo e, per la
rinomina, il nome di partenza. Restituisce uno di tre esiti: valido, non valido
con il motivo, duplicato con il nome esistente. Nella creazione un duplicato è
un errore. Nella rinomina un duplicato sul nome di partenza è valido, perché
cambia solo la grafia; un duplicato su un altro Gruppo chiede conferma per
l'unione. Nel salvataggio del Progetto un duplicato è il Gruppo da assegnare.
Il duplicato della creazione lo giudica il Manager: il dialogo non ripete il
controllo. Nella rinomina il dialogo confronta il nome con l'elenco per chiedere
conferma, e il Manager esegue la rinomina o l'unione.
Anche il salvataggio del Progetto richiama la validazione: un nome riservato è
rifiutato con `ErrInvalidGroupName`, e il form mostra il messaggio sotto il
campo Gruppo. Il modello del form richiama quella funzione e la riesporta per
il dialogo. Lo store non ripete il confronto: usa la regola del modello. La
vista e il modello del form
riesportano le due etichette per i testi dell'interfaccia. La rinomina su un nome esistente unisce i due Gruppi. Nel salvataggio del
Progetto un nome che corrisponde a un Gruppo esistente lo assegna, con la
grafia salvata. "Nessun Gruppo" e "Senza Gruppo" sono etichette riservate: non
possono essere nomi di un Gruppo. Nel form il valore vuoto del campo azzera il
Gruppo del Progetto; il nome vuoto è rifiutato solo dalla creazione e dalla
rinomina nel dialogo. "Nessun Gruppo" è il segnaposto del campo vuoto, non una
voce selezionabile.

Il modello pubblico resta basato sui nomi: il Progetto porta il nome del suo
Gruppo e lo snapshot porta l'elenco ordinato dei nomi, Gruppi vuoti inclusi.
Gli identificativi restano dentro la persistenza, così GUI e tray non cambiano
contratto.

Le posizioni sono intere e compatte, da zero in su: la creazione aggiunge in
fondo, "Sposta su/giù" scambia con il vicino, la rinomina conserva la
posizione, l'unione lascia il posto relativo del Gruppo che aveva già il nome
di arrivo e l'eliminazione chiude il buco rinumerando. Il posto relativo è la
posizione nell'ordine rispetto agli altri Gruppi, dopo la chiusura del buco: la
posizione assoluta di un Gruppo unito può diminuire di uno. L'ordine alfabetico
è crescente e confronta i nomi senza distinzione di maiuscole; a parità,
confronta la grafia esatta. Vale per la migrazione iniziale e per i Gruppi che
lo snapshot non elenca. L'ordine scelto resta quello principale.

### Stato e operazioni

Le operazioni sui Gruppi vivono sull'interfaccia del modulo compose, accanto a
quella che oggi salva le impostazioni del Progetto: crea, rinomina (con
unione), elimina, sposta. Sono serializzate con lo stesso lock che serializza
il refresh e non parlano con Docker: pubblicano subito uno snapshot nuovo, così
funzionano anche con Docker irraggiungibile.

Lo snapshot si costruisce leggendo i Gruppi dalla persistenza a ogni refresh,
e dopo un'operazione si aggiorna in memoria: l'elenco dei nomi, il nome del
Gruppo su ogni Progetto toccato e, per l'unione, il nome del Gruppo che
sopravvive. La derivazione dei Gruppi dai Progetti sparisce.

Il salvataggio delle impostazioni del Progetto resta l'unica porta del form:
se il nome digitato non esiste, il Gruppo viene creato nella stessa
transazione del Progetto, in fondo all'elenco.

Gli errori nuovi sono tre sentinelle: `ErrUnknownGroup`, `ErrDuplicateGroup` e
`ErrInvalidGroupName`. Diventano messaggi italiani nella tabella unica dei testi
degli errori, come gli altri errori del backend. `ErrInvalidGroupName` copre il
nome vuoto e le due etichette riservate, e il suo messaggio dice: "Nome del
Gruppo non valido: non può essere vuoto, né «Nessun Gruppo», né «Senza
Gruppo»." Così non si confonde con `ErrInvalidName` dei Progetti.
`ErrUnknownGroup` dice: "Gruppo sconosciuto." `ErrDuplicateGroup` dice: "Esiste
già un Gruppo con questo nome." Il rifiuto di un nome
duplicato dice quale Gruppo esiste: lo store restituisce
`model.DuplicateGroupError`, con campo `Existing string`, che avvolge
`ErrDuplicateGroup`. Il dialogo lo legge con `errors.As` e compone il messaggio.
`view.KnownMessage` resta il messaggio generico per la tabella unica.

### Migrazione

La migrazione ha tre versioni. La versione 0002 crea la tabella, l'indice e la
chiave esterna, e lascia la colonna del nome sul Progetto. Se
`schema_migrations` manca o è vuota e `user_version` è almeno 1, l'apertura
forza quella versione prima della scelta. L'apertura
sceglie la versione dello schema e decide. Se la versione è minore di 2, applica le
migrazioni fino alla 0002. Se la versione è 2, esegue il passo di dati in Go e
poi applica la 0003. Se la versione è 3, non fa nulla. La scelta è una funzione
pura, provata con le versioni 0, 1, 2, 3 e 4; la 4 è un errore. L'apertura ripete la scelta finché
la versione è minore di 3: dalla 0 applica la 0001 e la 0002, dalla 1 applica la 0002,
poi esegue il passo e applica la 0003 nella stessa apertura. Una versione
maggiore di 3 è un errore di apertura, con un messaggio chiaro. Il passo di dati è una funzione
esportata del package del database e riceve il contesto e il database, perché
solo il codice conosce gli spazi e le maiuscole Unicode. Il passo legge i nomi
dai Progetti, li ripulisce dagli spazi iniziali e finali, unisce le varianti
con la regola del codice, anche per le maiuscole accentate, e sceglie la grafia
del primo Progetto in ordine crescente di nome compose, senza distinzione di
maiuscole. Un nome che dopo la ripulitura è vuoto non crea un Gruppo. Un nome
che dopo la ripulitura
coincide con un'etichetta riservata, senza distinzione di maiuscole, non crea
un Gruppo: i suoi Progetti restano senza Gruppo. Poi scrive i Gruppi, i
collegamenti e la posizione alfabetica senza distinzione di maiuscole. Il passo
gira in una transazione e committa alla fine. Se la colonna `group_name` manca,
il passo non fa nulla. È idempotente: una seconda esecuzione non cambia nulla.
Il test lo esegue due volte su un database alla versione 2. Dopo il passo,
l'apertura applica la versione 0003, che toglie la colonna vecchia solo se
esiste. Il test apre anche un database alla versione 2 senza `group_name` e
verifica che l'apertura arrivi alla 3. La versione finale dello schema è la 3.
La discesa della 0003 ricrea la colonna e vi copia i nomi
dei Gruppi; la discesa della 0002 ricrea `projects` senza `group_id` e vi
conserva `group_name`, già copiato dalla discesa della 0003, poi cancella
l'indice e la tabella `groups`. La discesa resta in una transazione del runner. Copia `project_files` e
`active_profiles` in tabelle temporanee, ricostruisce `projects` con le colonne
`name`, `folder`, `display_name` e `group_name` dalla colonna esistente di
`projects`, reinserisce le righe figlie dalle tabelle temporanee, elimina le
tabelle temporanee e committa. Solo la discesa della 0003 legge i nomi dai
Gruppi. Nessun cambio al pragma delle chiavi esterne. Un test scende dalla
versione 3 alla 1 e verifica Progetti, Nome visualizzato, File compose, Profili
attivi, `group_name` e la perdita dei Gruppi vuoti. Un secondo test scende dalla
2 alla 1 con i nomi già nella colonna. Il test copre la discesa riuscita.
Una discesa fallita lascia il flag `dirty`: il test che simula l'errore
ripristina la versione con `Force` e toglie il flag.

### Presentazione

La lista costruisce le sue sezioni seguendo l'ordine dello snapshot. I Progetti
senza Gruppo restano in fondo nella sezione "Senza Gruppo". La lista resta
piatta e senza intestazione quando nessun Progetto porta un Gruppo; un Gruppo
salvato senza Progetti non produce un'intestazione. Un Progetto che porta un
Gruppo fuori dall'elenco dello snapshot non sparisce: le sezioni dei Gruppi
sconosciuti vanno in coda, in ordine alfabetico, dopo i Gruppi noti e prima di
"Senza Gruppo".

Il tray segue lo stesso elenco e salta i Gruppi senza Progetti, come già fa con
un sottomenu vuoto. I Progetti senza Gruppo restano voci di primo livello in
fondo, come oggi. Un Progetto che porta un Gruppo fuori dall'elenco non
sparisce dal tray: il suo sottomenu va dopo i Gruppi noti e prima dei Progetti
senza Gruppo, in ordine alfabetico.

Nel form del Progetto il campo Gruppo propone i Gruppi dello snapshot
nell'ordine scelto, con il segnaposto "Nessun Gruppo" quando il valore è vuoto;
resta modificabile e il suggerimento dice che un nome non in elenco crea un
Gruppo. Mostra anche il Gruppo del Progetto quando lo snapshot non lo elenca, e
il salvataggio lo assegna o lo crea. Il Gruppo non elencato va dopo i Gruppi
noti, in ordine alfabetico. Accanto al campo un comando "Gruppi…" apre
il dialogo. Alla chiusura il form si ridisegna con
l'elenco aggiornato. Il dialogo restituisce alla shell un record per ogni
operazione riuscita: il tipo (creazione, rinomina, unione o eliminazione), il
Gruppo di partenza e il Gruppo di arrivo. Per una creazione la partenza è vuota
e l'arrivo è il nome salvato del Gruppo nuovo.
Per un'unione l'arrivo è il Gruppo sopravvissuto. Per un'eliminazione l'arrivo è
vuoto. Il form applica l'elenco in ordine,
non un confronto fra snapshot. L'elenco contiene solo le operazioni che
cambiano i nomi: uno spostamento non entra nell'elenco. Il
campo confronta il valore con i Gruppi dopo la
ripulitura e senza distinzione di maiuscole. Cambia solo se mostra un Gruppo che
il dialogo ha toccato: dopo una rinomina o un'unione passa al Gruppo
sopravvissuto, dopo un'eliminazione si svuota, e se il dialogo ha creato quel
nome mostra la grafia salvata. Se mostra un altro Gruppo, o un nome che non
corrisponde a nessun Gruppo, resta com'è. Nome visualizzato e Profili restano
quelli digitati.

Il dialogo di gestione è modale e separato dal form. Contiene l'elenco dei
Gruppi con il numero di Progetti ("nessun Progetto" quando è vuoto), i comandi
Nuovo, Rinomina, Elimina, Sposta su, Sposta giù e Chiudi. Nuovo e Rinomina
aprono in riga un campo per il nome, con i comandi Conferma e Annulla; il
messaggio di rifiuto sta sotto il campo e si ripete nella sua descrizione
accessibile. Un errore del Manager che non riguarda il campo, come un Gruppo
sconosciuto, compare in una riga sotto l'elenco con il messaggio italiano; se il
duplicato non porta il nome nell'errore, il messaggio è quello generico.
Rinomina parte dal nome del Gruppo scelto.
Le operazioni si applicano subito, senza un Salva del dialogo; Elimina e
l'unione chiedono conferma con il conteggio dei Progetti che si spostano. Il
dialogo esegue ogni operazione fuori dal thread Qt e spegne i comandi finché non
torna l'esito: un secondo clic non avvia una seconda operazione. Al ritorno
riaccende i comandi. Durante l'operazione i comandi e le scorciatoie sono
spenti. Escape e Chiudi sono ignorati. Il dialogo applica l'esito prima di
ripristinare la selezione dallo snapshot. I
comandi non disponibili sono spenti: Rinomina, Elimina e Sposta senza selezione,
Sposta su sul primo Gruppo e Sposta giù sull'ultimo. Nuovo resta attivo quando
il campo del nome è chiuso.
Mentre il campo del nome è aperto, Nuovo, Rinomina, Elimina, Sposta e Chiudi
sono spenti, e l'elenco ignora i clic. Il Gruppo da rinominare è quello scelto
all'apertura del campo, e la selezione non cambia. Conferma chiude il campo solo quando l'operazione riesce. Annulla
lo chiude sempre. Un rifiuto lo lascia aperto, con il messaggio sotto il campo.
Uno snapshot conserva il testo digitato.
L'elenco conserva la selezione quando arriva uno snapshot, e i comandi si
aggiornano di conseguenza. Dopo un'unione la selezione va al Gruppo
sopravvissuto; dopo una rinomina semplice resta sul Gruppo rinominato; dopo
Nuovo va al Gruppo nuovo; dopo uno spostamento resta sul Gruppo spostato. Dopo
un'eliminazione va al Gruppo che prende la posizione vuota; se non c'è,
all'ultimo; se l'elenco è vuoto, a nessuno. Una rinomina che cambia solo le
maiuscole dello stesso Gruppo non è un'unione e non chiede conferma.

Tastiera e accessibilità seguono le regole della finestra: ordine di
tabulazione esplicito — elenco, Nuovo, Rinomina, Elimina, Sposta su, Sposta
giù, Chiudi; con il campo aperto, campo, Conferma, Annulla —. Invio conferma il
campo aperto, Escape annulla il campo aperto e lascia il dialogo aperto; senza
campo aperto, Escape chiude il dialogo. Alt+Su e Alt+Giù spostano il Gruppo
selezionato, nomi accessibili su elenco, campo e comandi, bersagli di almeno
24 px.

Il codice morto del vecchio "Nuovo Gruppo…" nel modello del form esce dal
repository: la creazione ha ora due porte reali, il campo e il dialogo.

## Testing Decisions

Un buon test guarda il comportamento esterno al modulo e non la sua
implementazione. I seam esistenti sono le funzioni pure esportate della vista e
del modello del form (`Sections`, `KnownMessage`, le scelte del campo), il menu
del tray, il Manager con un database SQLite temporaneo e la shell offscreen. Due
seam sono nuovi: la scelta della versione di migrazione, funzione pura nel
package del database, e il passo di dati esportato dallo stesso package, che
riceve il contesto e il database. Un terzo seam nuovo è la validazione del nome
del Gruppo, funzione pura del modello, riesportata dal modello del form. La validazione con il duplicato
serve alla creazione e alla rinomina del dialogo. Il form rifiuta solo le
etichette riservate, il valore vuoto azzera il Gruppo e un nome che differisce
per maiuscole assegna il Gruppo salvato. Un test non dipende da nomi di funzioni
interne, da query SQL o dalla struttura dei widget oltre a ciò che l'utente
vede. Il database legacy del test di migrazione è l'input del test, e si
costruisce con lo schema della versione 1.

**Manager.** Le quattro operazioni e il salvataggio del form che crea un Gruppo
si provano sull'interfaccia del modulo compose, con un database SQLite
temporaneo e il finto Docker già in uso. Si verificano: ordine nello snapshot,
Gruppo nuovo in fondo, rinomina che segue su tutti i Progetti, unione che lascia
nome e posto relativo del Gruppo che aveva già il nome di arrivo, eliminazione
che lascia i Progetti senza Gruppo, un Gruppo senza Progetti preparato salvando
un Progetto con quel Gruppo e togliendolo dalla lista, spostamenti ai bordi che
non cambiano nulla,
ordine conservato dopo la riapertura del database, form con il Gruppo vuoto che
azzera il Gruppo, errori per Gruppo sconosciuto, nome duplicato e nome non
valido (vuoto o riservato), e funzionamento con Docker irraggiungibile. Precedente: i test delle impostazioni
e delle azioni del Manager.

**Messaggi.** Un test verifica, per ognuno dei tre errori nuovi, che
`view.KnownMessage` restituisca un messaggio italiano. Il test non legge la
tabella interna dei testi. Il messaggio di `ErrInvalidGroupName` è quello esatto
della spec e non si confonde con quello di `ErrInvalidName` dei Progetti.
Precedente: i test dei messaggi degli errori.

**Tray.** Un test del menu con un Gruppo vuoto nello snapshot verifica che il
sottomenu vuoto non compaia; un test con un Progetto che porta un Gruppo fuori
dall'elenco verifica che il suo sottomenu vada dopo i Gruppi noti e prima dei
Progetti senza Gruppo. Due Gruppi noti in un ordine scelto non alfabetico
restano in quell'ordine, e due Gruppi sconosciuti vanno in ordine alfabetico.
Precedente: i test del menu del tray.

**Migrazione.** Si apre un database nuovo (schema aggiornato, versione finale
3) e un database "legacy" alla versione 1 con Progetti e Gruppi già scritti,
compresi un Nome visualizzato e dei Profili attivi. I
casi includono
due varianti che differiscono solo per maiuscole ASCII, una coppia con
maiuscole accentate ("Città" e "CITTÀ"), un nome con uno spazio non ASCII
intorno e un nome con uno spazio interno. Si verifica che i Progetti restino,
che ogni coppia diventi un Gruppo solo con la grafia del primo Progetto in
ordine di nome compose, che lo spazio intorno sparisca, che il nome con lo
spazio interno resti distinto, che la posizione iniziale sia alfabetica senza
distinzione di maiuscole e che una seconda apertura non cambi nulla. Si
verifica che il Nome visualizzato e i Profili attivi restino. Un nome
legacy che coincide con un'etichetta riservata non crea un Gruppo, e i suoi
Progetti restano senza Gruppo. Un test interno al package del database apre un
database alla versione 2 con lo stesso motore delle migrazioni ed esegue il
passo di dati due volte: Gruppi, collegamenti e ordine non cambiano. La
decisione per la versione 2 richiede il passo; l'apertura lo esegue anche dopo
aver portato 0 o 1 alla 2. La versione 4 produce un errore. Precedente: il test di adozione del
database legacy.

**Funzioni pure.** L'ordine delle sezioni dato l'elenco dei Gruppi. I casi
includono un Gruppo sconosciuto allo snapshot, che va dopo i Gruppi noti e
prima di "Senza Gruppo"; due Gruppi noti in un ordine scelto non alfabetico,
che restano in quell'ordine; l'assenza di Progetti con Gruppo, anche con un
Gruppo salvato vuoto, che lascia la lista piatta e senza intestazione; e un
Gruppo salvato senza Progetti, che non produce intestazione. Si provano anche
le scelte del campo Gruppo, il segnaposto "Nessun Gruppo", il Gruppo del
Progetto che lo snapshot non elenca, e la validazione del nome (spazi iniziali e
finali, vuoto, duplicato senza maiuscole, etichette riservate). Precedente: i
test delle sezioni e del modello del form.

**Shell offscreen.** Un passo del test della finestra senza display: aprire il
form, aprire "Gruppi…", creare, rinominare, eliminare e spostare attraverso i
widget del dialogo, e controllare l'elenco del dialogo, il campo Gruppo del
form, le intestazioni della lista e l'ordine. Si controllano anche i comandi
spenti (compresi quelli spenti mentre il campo è aperto), il messaggio di
rifiuto e la sua descrizione accessibile, il testo digitato che sopravvive a uno
snapshot, il focus che torna al comando "Gruppi…", i bersagli di almeno 24 px,
il giro di tastiera e il funzionamento con Docker irraggiungibile. Si
controllano la selezione dopo una rinomina semplice, dopo Nuovo, dopo un'unione,
dopo un'eliminazione e dopo uno spostamento. Le tre operazioni di creazione,
rinomina ed eliminazione fanno parte del passo. I callback del dialogo, come il
callback `onSave` del form, permettono a un test di bloccare un'operazione e
verificare che il dialogo resti reattivo. Precedente: il test
offscreen esistente.

La verifica finale è `make check` e `make test-ui`; gli screenshot restano un
passaggio manuale.

## Out of Scope

- Più Gruppi per Progetto (tag).
- Colore, icona o descrizione di un Gruppo.
- Assegnare o spostare più Progetti in un Gruppo in blocco.
- Un comando "Riordina alfabeticamente".
- Un secondo punto d'ingresso alla gestione dalla colonna Progetti, dal tray o dalle Impostazioni: si entra solo dal form del Progetto.
- L'ordine dei Progetti dentro un Gruppo, che resta quello dello snapshot.
- Import, export o sincronizzazione dei Gruppi.

## Further Notes

La decisione "i Gruppi sono entità salvate" è registrata nell'ADR 0007, con
l'alternativa scartata (Gruppi derivati dai Progetti) e il motivo: Gruppi
vuoti, ordine scelto e rinomina a costo di una riga.

`GLOSSARY.md` è già aggiornato con Gruppo, Senza Gruppo, Rinomina di un Gruppo
ed Eliminazione di un Gruppo.

Nel working tree ci sono modifiche non committate che anticipano la migrazione
e parte della persistenza dei Gruppi. Compilano, ma lasciano rossi i test che
codificano il comportamento vecchio: l'ordine alfabetico, la versione 1 dello
schema e il Gruppo che resta dopo l'uscita dell'ultimo Progetto. Vanno
completate, con quei test aggiornati, o scartate; il punto di partenza è un
albero verde (`make check`, `make test-ui`). Se restano, devono rispettare la
spec: l'unione dei nomi usa le regole Unicode del codice e non le funzioni
ASCII di SQLite, la grafia del Gruppo unito viene dal primo Progetto in ordine
di nome compose e non da `MIN()` della SQL, un nome digitato che corrisponde a
un Gruppo salvato lascia al Progetto la grafia salvata, e il commento dello
snapshot descrive l'ordine scelto con i Gruppi vuoti inclusi.

Un database locale creato dalla 0002 provvisoria ha già rimosso `group_name`.
Va eliminato prima di provare la versione nuova. Il passo di dati non fa nulla
se la colonna manca, quindi l'app non si blocca.

La migrazione inversa esiste e un test scende dalla versione 3 alla 1.
