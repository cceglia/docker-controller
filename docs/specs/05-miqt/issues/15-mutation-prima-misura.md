# 15: Mutation testing: prima misura

**What to build:** Il repository misura la qualità delle asserzioni con gremlins sui test unitari. Un `.gremlins.yaml` versionato fissa `workers: 0`, `timeout-coefficient: 5` e i file esclusi. `make tools` installa gremlins `v0.6.0`. `make mutate` esegue la mutation testing su `internal/tray`, `internal/gui` e `internal/compose`, fuori da `make check`. Questo giro produce solo un report: nessuna soglia, nessun errore.

**Blocked by:** 14

**Status:** ready-for-agent

- [ ] `.gremlins.yaml` è versionato con `workers: 0` e `timeout-coefficient: 5`
- [ ] `make tools` installa gremlins `v0.6.0` con `go install`
- [ ] `make mutate` esegue gremlins sui tre package scelti
- [ ] `make check` non esegue gremlins
- [ ] Il giro produce un report con efficacia e copertura dei mutanti per ogni package
- [ ] Il giro non imposta soglie e non esce con errore
- [ ] Sono esclusi `internal/db`, il fake di compose e i file generati
- [ ] Il README spiega come lanciare `make mutate` e quanto dura
- [ ] I numeri misurati sono scritti nel ticket, per ogni package
- [ ] La durata del giro è scritta nel ticket

## Comments

La mutation testing non entra in `make check`: su questi package dura minuti, non secondi. Il repository non ha integrazione continua. Le soglie arrivano nel ticket 16.
