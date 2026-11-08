# Gruppi come entità salvate

Un Gruppo era solo il nome scritto su ogni Progetto (`projects.group_name`),
derivato in elenco a ogni snapshot: non poteva esistere senza Progetti, non
aveva un ordine proprio e rinominarlo significava riscrivere ogni Progetto che
lo usava. Con la gestione dei Gruppi richiesta dall'utente questo modello non
regge: creare, rinominare ed eliminare un Gruppo sono operazioni sul Gruppo, non
sui Progetti, e l'ordine scelto dall'utente deve vivere da qualche parte. I
Gruppi diventano quindi entità salvate, con una tabella propria per nome e
posizione; i Progetti puntano al loro Gruppo con una chiave esterna che lascia
il Progetto senza Gruppo quando il Gruppo viene eliminato.

## Considered Options

Restare sul nome derivato e aggiungere al Manager le operazioni che riscrivono
il nome su tutti i Progetti: nessuna migrazione, ma niente Gruppi vuoti, ordine
sempre alfabetico e nessun posto dove mettere un Gruppo prima di assegnarlo.
Scartata perché l'utente ha chiesto una gestione separata dei Gruppi, e le tre
cose che il modello derivato non può fare sono esattamente quelle che la
gestione richiede.

## Consequences

La migrazione deve creare le righe dei Gruppi dai nomi già salvati, unire le
varianti che differiscono solo per spazi iniziali o finali o per maiuscole e
assegnare l'ordine alfabetico iniziale. La grafia del Gruppo unito è quella del
primo Progetto in ordine di nome compose. L'unione usa la stessa regola del
codice, che conosce l'Unicode, e non le funzioni ASCII di SQLite: "Città" e
"CITTÀ" sono lo stesso Gruppo. Il nome resta l'identità del Gruppo, confrontata
senza distinzione di maiuscole dal codice: un indice unico del database copre
solo la grafia esatta, perché `NOCASE` di SQLite ignora soltanto l'ASCII. Lo
snapshot pubblico continua a parlare di nomi, così GUI e tray non conoscono gli
identificativi. L'ordine manuale vale per lista, tray e campo del form, e i
Gruppi senza Progetti compaiono solo nel dialogo di gestione e nel campo, mai
come intestazioni o sottomenu vuoti. "Nessun Gruppo" e "Senza Gruppo" sono
etichette riservate: non possono essere nomi di un Gruppo.
