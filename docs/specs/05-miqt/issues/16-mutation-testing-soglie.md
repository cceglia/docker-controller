# 16: Mutation testing: soglie

**What to build:** `make mutate` fallisce quando la qualità dei test scende. Le soglie di efficacia e di copertura dei mutanti vivono in `.gremlins.yaml`, fissate al valore misurato dal ticket 15 meno cinque punti. Sotto soglia il target esce con errore e il report dice quali mutanti sono sopravvissuti.

**Blocked by:** 15

**Status:** ready-for-agent

- [ ] Le soglie di efficacia e di copertura dei mutanti sono in `.gremlins.yaml`
- [ ] Ogni soglia vale il valore misurato dal ticket 15 meno cinque punti
- [ ] `make mutate` esce con errore sotto soglia
- [ ] Il messaggio d'errore nomina la soglia mancata e il valore misurato
- [ ] Il report elenca i mutanti sopravvissuti con il punto del codice
- [ ] Una soglia alzata a mano fa fallire il target, per prova
- [ ] Il README spiega come si aggiorna una soglia, e perché serve una misura nuova
- [ ] `make check` resta senza gremlins
- [ ] Il primo giro con le soglie è documentato nel ticket, con i numeri

## Comments

La misura arriva dal ticket 15. Questo ticket fissa il cancello.
