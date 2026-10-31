# 03: Dettaglio del Progetto: identità e azioni

**What to build:** Il dettaglio si apre con l'intestazione del Progetto — nome, stato come punto colorato più testo, striscia dei Container, nome compose e cartella in monospaziato — e con le azioni di Progetto lì, non più nella lista. Una sola azione è primaria e cambia secondo lo stato (Up se Fermo o Rimosso, Stop se Attivo, Restart se Parziale); le altre restano discrete; Down, Down con volumi e Rimuovi dalla lista stanno dietro un menu. Sotto, le celle con Cartella, Nome compose e Gruppo, i comandi Cartella, Terminale e Modifica, i Profili come caselle e gli avvisi per Attività in corso e File compose mancante.

**Blocked by:** 02

**Status:** resolved

- [x] L'intestazione mostra nome, stato (punto e testo), striscia dei Container, nome compose e cartella.
- [x] Le cinque azioni di Progetto sono nell'intestazione e una sola è primaria, secondo lo stato.
- [x] Le azioni non primarie sono raggiungibili ma discrete; tutte sono disabilitate durante un'Attività o con Docker irraggiungibile.
- [x] Down, Down con volumi e Rimuovi dalla lista stanno nel menu `…` e le conferme sono identiche a oggi, compresa la digitazione del nome per il Down con volumi.
- [x] Cartella, Nome compose e Gruppo sono leggibili in celle; Cartella e Terminale aprono il gestore file e il terminale come oggi.
- [x] Modifica apre gli stessi campi di oggi (Nome visualizzato, Gruppo esistente o nuovo) con la stessa validazione.
- [x] I Profili sono caselle con la spiegazione che cambiarli non avvia né ferma nulla, e il salvataggio resta quello di oggi.
- [x] Un'Attività in corso sul Progetto è visibile con la stessa informazione di oggi.
- [x] Con File compose mancante compare un avviso che, per un Progetto Rimosso, offre la rimozione dalla lista.
- [x] `svelte-check` e la build sono verdi.

## Comments

Implementato nel commit `6a8fdbe`.

Revisione finale del 2026-10-31 con punto fisso `0ab9372`: nessun rilievo
bloccante su standard o spec. I rilievi non bloccanti sono nel report del run.
