# 04: Servizi e Container

**What to build:** I Servizi del Progetto aperto diventano blocchi piatti con nome, striscia dei Container e azioni Start, Stop, Restart e Down di Servizio; sotto, ogni Container è una riga leggibile con punto di stato, nome in monospaziato, immagine, stato, health e porte pubblicate apribili. La sezione dichiara quanti Servizi ha e, quando è vuota, spiega perché. "Mostra log" resta dov'è oggi e apre lo stream del Progetto.

**Blocked by:** 03

**Status:** ready-for-agent

- [ ] Ogni Servizio è un blocco piatto con nome, striscia dei Container e azioni.
- [ ] Start, Stop e Restart di Servizio funzionano come oggi; il Down di Servizio chiede conferma.
- [ ] Ogni Container è una riga con punto di stato, nome in monospaziato, immagine, stato e health.
- [ ] La pillola colorata compare solo per health problematico (`unhealthy`, `starting`), non per lo stato normale.
- [ ] Le porte pubblicate sono comandi che aprono il browser, come oggi, e restano leggibili anche con più porte.
- [ ] "Mostra log" è nello stesso punto e apre lo stream del Progetto.
- [ ] Il numero di Servizi è visibile accanto al titolo della sezione.
- [ ] Un Progetto senza Servizi mostra un messaggio che spiega perché, non un vuoto.
- [ ] `svelte-check` e la build sono verdi.
