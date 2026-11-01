# 16: Mutation testing: soglie

**What to build:** `make mutate` fallisce quando la qualità dei test scende. Le soglie di efficacia e di copertura dei mutanti vivono in un file per package sotto `.gremlins/`, fissate al valore misurato dal ticket 15 meno cinque punti. Sotto soglia il target esce con errore e il report dice quali mutanti sono sopravvissuti.

**Blocked by:** 15

**Status:** resolved

- [x] Le soglie di efficacia e di copertura dei mutanti sono in un file di configurazione per package, sotto `.gremlins/`
- [x] Ogni soglia vale il valore misurato dal ticket 15 meno cinque punti
- [x] `make mutate` esce con errore sotto soglia
- [x] Il messaggio d'errore nomina la soglia mancata e il valore misurato
- [x] Il report elenca i mutanti sopravvissuti con il punto del codice
- [x] Una soglia alzata a mano fa fallire il target, per prova
- [x] Il README spiega come si aggiorna una soglia, e perché serve una misura nuova
- [x] `make check` resta senza gremlins
- [x] Il primo giro con le soglie è documentato nel ticket, con i numeri

## Comments

La misura arriva dal ticket 15. Questo ticket fissa il cancello.

## Esito

Le tre coppie di soglie sono diverse, quindi un solo `.gremlins.yaml` non basta.
I file stanno in `.gremlins/tray.yaml`, `.gremlins/gui.yaml` e
`.gremlins/compose.yaml`. `make mutate` legge per ogni package la soglia dal suo
file e passa `--config`. Gremlins legge un solo file di configurazione: `--config`
sostituisce quello condiviso.

Ogni soglia è il valore del ticket 15 meno cinque punti.

| Package | Efficacia | Copertura | Soglia efficacia | Soglia copertura | Esito |
| --- | ---: | ---: | ---: | ---: | --- |
| `internal/tray` | 89,66% | 93,55% | 84,66% | 88,55% | passa |
| `internal/gui` | 90,33% | 41,90% | 85,33% | 36,90% | passa |
| `internal/compose` | 76,85% | 91,53% | 71,85% | 86,53% | passa |

**Prima prova delle soglie**, package per package con la configurazione
versionata: `tray` 26 uccisi, 3 sopravvissuti, 0 scaduti; `gui` 271 uccisi, 29
sopravvissuti, 1 scaduto; `compose` 166 uccisi, 50 sopravvissuti, 4 scaduti.

**Prova del cancello**. Con `efficacy: 95.00` nel file della tray, `make mutate`
esce con errore al primo package e il messaggio dice:

```
== ./internal/tray (soglie: efficacia 95.00%, copertura dei mutanti 88.55%)
Test efficacy: 89.66%
ERROR: below efficacy-threshold
SOGLIA MANCATA: ./internal/tray, efficacia 89.66%, soglia 95.00%
```

Poi la soglia è tornata a 84,66. Il report elenca i mutanti sopravvissuti con
file e posizione, per esempio `LIVED CONDITIONALS_BOUNDARY at logs.go:152:8`.

**Nota sulla durata**. Il giro completo dura da 8 a 15 minuti secondo il carico.
In due giri lunghi l'ambiente ha mandato `SIGTERM` a `make` durante il package
`gui`, quindi il verde è documentato package per package, con la stessa
configurazione. `make check` non esegue gremlins.
