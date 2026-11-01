# 05: Log

**What to build:** La scheda Log mostra i log del Progetto o del Servizio scelto. Le righe sono colorate per Servizio e lo stderr si distingue. Ci sono ricerca, pausa dello scorrimento, timestamp mostra e nascondi, un tetto di 5.000 righe, la selezione e la copia del testo. Chiudere la scheda termina il processo di compose.

**Blocked by:** 04

**Status:** ready-for-agent

- [ ] Log dell'intero Progetto e log di un singolo Servizio
- [ ] All'apertura il pannello mostra subito le ultime 500 righe, poi segue lo stream
- [ ] Colore per Servizio e stderr distinguibile dal resto
- [ ] La ricerca filtra le righe visibili senza fermare lo stream
- [ ] La pausa ferma lo scorrimento automatico e non lo stream
- [ ] I timestamp si mostrano e si nascondono senza riavviare lo stream
- [ ] Il pannello tiene al massimo le ultime 5.000 righe
- [ ] Il testo si seleziona e si copia
- [ ] Chiudere la scheda termina il processo `docker compose logs`
- [ ] I log di un Servizio con Profilo non attivo si leggono lo stesso
- [ ] Un Progetto con Attività in corso mostra il suo output separato dai log
- [ ] Test da tavolo per il tetto delle 5.000 righe e per la mappatura delle righe nei tag del log
