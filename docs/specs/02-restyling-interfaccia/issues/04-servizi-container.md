# 04: Servizi e Container

**What to build:** I Servizi del Progetto aperto diventano blocchi piatti con nome, striscia dei Container e azioni Start, Stop, Restart e Down di Servizio; sotto, ogni Container è una riga leggibile con punto di stato, nome in monospaziato, immagine, stato, health e porte pubblicate apribili. La sezione dichiara quanti Servizi ha e, quando è vuota, spiega perché. "Mostra log" resta dov'è oggi e apre lo stream del Progetto.

**Blocked by:** 03

**Status:** resolved

- [x] Ogni Servizio è un blocco piatto con nome, striscia dei Container e azioni.
- [x] Start, Stop e Restart di Servizio funzionano come oggi; il Down di Servizio chiede conferma.
- [x] Ogni Container è una riga con punto di stato, nome in monospaziato, immagine, stato e health.
- [x] La pillola colorata compare solo per health problematico (`unhealthy`, `starting`), non per lo stato normale.
- [x] Le porte pubblicate sono comandi che aprono il browser, come oggi, e restano leggibili anche con più porte.
- [x] "Mostra log" è nello stesso punto e apre lo stream del Progetto.
- [x] Il numero di Servizi è visibile accanto al titolo della sezione.
- [x] Un Progetto senza Servizi mostra un messaggio che spiega perché, non un vuoto.
- [x] `svelte-check` e la build sono verdi.

## Comments

Implementato nel commit `0c0cdd1`.

Revisione finale del 2026-10-31 con punto fisso `0ab9372`: nessun rilievo
bloccante su standard o spec. I rilievi non bloccanti sono nel report del run.
