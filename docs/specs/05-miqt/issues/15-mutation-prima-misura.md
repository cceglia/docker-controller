# 15: Mutation testing: prima misura

**What to build:** Il repository misura la qualità delle asserzioni con gremlins sui test unitari. Un `.gremlins.yaml` versionato fissa `workers: 0`, `timeout-coefficient: 20` e i file esclusi. `make tools` installa gremlins `v0.6.0`. `make mutate` esegue la mutation testing su `internal/tray`, `internal/gui` e `internal/compose`, fuori da `make check`. Questo giro produce solo un report: nessuna soglia, nessun errore.

**Blocked by:** 14

**Status:** resolved

- [x] `.gremlins.yaml` è versionato con `workers: 0` e `timeout-coefficient: 20`
- [x] `make tools` installa gremlins `v0.6.0` con `go install`
- [x] `make mutate` esegue gremlins sui tre package scelti
- [x] `make check` non esegue gremlins
- [x] Il giro produce un report con efficacia e copertura dei mutanti per ogni package
- [x] Il giro non imposta soglie e non esce con errore
- [x] Sono esclusi `internal/db`, il fake di compose e i file generati
- [x] Il README spiega come lanciare `make mutate` e quanto dura
- [x] I numeri misurati sono scritti nel ticket, per ogni package
- [x] La durata del giro è scritta nel ticket

## Comments

La mutation testing non entra in `make check`: su questi package dura minuti, non secondi. Il repository non ha integrazione continua. Le soglie arrivano nel ticket 16.

## Esito

Misura del 1 novembre 2026, con `make mutate` e la configurazione versionata.
Il comando è uscito con codice 0. Nessuna soglia è impostata.

| Package            | Uccisi | Sopravvissuti | Non coperti | Scaduti | Totale | Efficacia | Copertura | Durata |
| ------------------ | -----: | ------------: | ----------: | ------: | -----: | --------: | --------: | -----: |
| `internal/tray`    |     26 |             3 |           2 |       0 |     31 |    89,66% |    93,55% |  4,0 s |
| `internal/gui`     |    271 |            29 |         416 |       1 |    717 |    90,33% |    41,90% | 5m 18s |
| `internal/compose` |    166 |            50 |          20 |       4 |    240 |    76,85% |    91,53% | 2m 04s |

Durata totale di `make mutate`: 452 s, cioè 7 minuti e 32 secondi.

Lettura dei numeri: `internal/gui` ha molti mutanti non coperti, perché i test
con tag `ui` non girano senza display. `internal/compose` ha l'efficacia più
bassa: cinquanta mutanti sopravvivono.

### Il coefficiente di timeout

La prima misura usava `timeout-coefficient: 5`, come chiedeva la specifica. Con
quel valore il package `tray` contava 28 mutanti scaduti su 31: il timeout vale
cinque volte il giro di copertura, cioè circa 687 ms, e la compilazione
parallela lo supera. L'efficacia del 100% veniva da un solo mutante ucciso.

La prova con `--timeout-coefficient=20` dà 26 uccisi, 3 sopravvissuti, 0
scaduti in 4 secondi. La specifica e `.gremlins.yaml` passano a 20, con il
motivo scritto accanto al valore. I numeri qui sopra sono quelli del coefficiente
20, misurati con un giro completo.

Restano quattro mutanti scaduti in `compose` e uno in `gui`: sono cicli resi
infiniti dalla mutazione, non un limite di tempo.

### Note di configurazione

- `.gremlins.yaml` non è nel `.gitignore`.
- `gremlins --version` stampa `dev`, perché `go install` da modulo non imposta
  le variabili di link. Il codice installato è il tag `v0.6.0`.
- Le esclusioni coprono `internal/db/`, `internal/compose/composefake/` e i file
  generati con i suffissi `_gen.go`, `_generated.go` e `.gen.go`, più la
  cartella `generated/`. Oggi il repository non ha file generati.
