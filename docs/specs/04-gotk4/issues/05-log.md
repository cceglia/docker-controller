# 05: Log

**What to build:** La scheda Log mostra i log del Progetto o del Servizio scelto. Le righe sono colorate per Servizio e lo stderr si distingue. Ci sono ricerca, pausa dello scorrimento, timestamp mostra e nascondi, un tetto di 5.000 righe, la selezione e la copia del testo. Chiudere la scheda termina il processo di compose.

**Blocked by:** 04

**Status:** resolved

- [x] Log dell'intero Progetto e log di un singolo Servizio
- [x] All'apertura il pannello mostra subito le ultime 500 righe, poi segue lo stream
- [x] Colore per Servizio e stderr distinguibile dal resto
- [x] La ricerca filtra le righe visibili senza fermare lo stream
- [x] La pausa ferma lo scorrimento automatico e non lo stream
- [x] I timestamp si mostrano e si nascondono senza riavviare lo stream
- [x] Il pannello tiene al massimo le ultime 5.000 righe
- [x] Il testo si seleziona e si copia
- [x] Chiudere la scheda termina il processo `docker compose logs`
- [ ] I log di un Servizio con Profilo non attivo si leggono lo stesso
- [x] Un Progetto con Attività in corso mostra il suo output separato dai log
- [x] Test da tavolo per il tetto delle 5.000 righe e per la mappatura delle righe nei tag del log

## Comments

Implementato nel commit `97c1648`.

Provato con log veri: le ultime 500 righe arrivano subito, il processo
`docker compose logs` sparisce dopo `Chiudi log`. Il filtro su 5000 righe costa
circa 1,3 ms nel buffer vero.

Resta aperto il caso del Servizio con Profilo non attivo: su questa macchina non
esiste un Progetto così, e `compose.Logs` chiede già `--profile *`.
