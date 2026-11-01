# 03: Verifica della UI e token di design

**What to build:** Il repository verifica la UI senza display. `make test-ui` costruisce la finestra con `QT_QPA_PLATFORM=offscreen`, la distrugge e controlla i token di design: la scala tipografica, il corpo minimo di 16 px, i rapporti di contrasto 4,5:1 e 3:1 e il bersaglio minimo di 24x24 px. I ticket successivi estendono il test quando aggiungono un ruolo al foglio di stile.

**Blocked by:** 02

**Status:** ready-for-agent

- [ ] `make test-ui` esiste ed esegue i test della UI con `QT_QPA_PLATFORM=offscreen`
- [ ] `make check` resta senza display: i test offscreen e gremlins non entrano nel target
- [ ] Lo smoke test costruisce la finestra vuota, la mostra con il backend offscreen e la distrugge
- [ ] Lo smoke test fallisce quando un costruttore o un cablaggio è rotto
- [ ] Il test dei token controlla la scala tipografica e il corpo minimo di 16 px
- [ ] Il test dei token controlla i contrasti 4,5:1 per il testo e 3:1 per componenti e bordi
- [ ] Il test dei token controlla il bersaglio minimo di 24x24 px
- [ ] Il test legge i valori da una sola sorgente, la stessa che il codice applica
- [ ] Un ruolo nuovo senza controllo fa fallire il test
- [ ] Il README spiega `make test-ui` e il suo prerequisito Qt

## Comments

Il test dei token nasce con i ruoli della lista e delle etichette di stato. Il ticket 07 aggiunge i colori dei Servizi nel log.
