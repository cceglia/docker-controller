# 16: gremlins: config e soglie

**What to build:** Ogni package di codice di produzione dello split ha il proprio file `.gremlins/<path>.yaml`, con il nome derivato dal path (`compose-model.yaml`, `gui-view.yaml`, …). `make mutate` esegue tutti i package e `make mutate PKG=…` ne esegue uno; `make test-ui` copre `./internal/gui/...`; `exclude-files` punta a `internal/compose/docker/fake/`. Le soglie sono misurate dopo lo split e pari al misurato meno 5 punti; un package senza mutanti è annotato, non inventato. `docker/fake` e `guitest` restano fuori da `MUTATE_PACKAGES`. Il README aggiorna tabella e percorsi. Contesto: spec `06-sottopackage-compose-e-gui`, ADR 0006.

**Blocked by:** 07 (compose: chiudere la radice), 15 (gui: chiudere la radice).

**Status:** resolved

- [x] Esistono quindici file `.gremlins/<path>.yaml` con il blocco condiviso identico.
- [x] `make mutate` gira tutti i package; `make mutate PKG=…` ne gira uno.
- [x] Le soglie sono misurate e scritte; `docker/fake` e `guitest` restano esclusi.
- [x] Un test verifica che `workers`, `timeout-coefficient` ed `exclude-files` siano identici in tutti i file.
- [x] Il README riporta le nuove soglie e `./internal/gui/...`.

## Misure

Misure del ticket 16 per compose e gui, del ticket 15 per `internal/tray`. La soglia è il valore misurato meno 5 punti.

| Package | Efficacia | Copertura |
| --- | ---: | ---: |
| `internal/compose` | 79,82% | 90,83% |
| `internal/compose/model` | 100,00% | 25,00% |
| `internal/compose/docker` | 66,67% | 56,25% |
| `internal/compose/store` | 96,88% | 100,00% |
| `internal/compose/parse` | 83,93% | 93,33% |
| `internal/gui` | 90,30% | 41,36% |
| `internal/gui/design` | 85,71% | 87,50% |
| `internal/gui/view` | 87,95% | 94,32% |
| `internal/gui/project` | 94,29% | 77,78% |
| `internal/gui/forms` | 83,67% | 77,78% |
| `internal/gui/activity` | 95,83% | 100,00% |
| `internal/gui/logs` | 93,18% | 97,78% |
| `internal/gui/usage` | 86,21% | 93,55% |
| `internal/gui/traymenu` | 100,00% | 100,00% |
| `internal/tray` | 89,66% | 93,55% |
