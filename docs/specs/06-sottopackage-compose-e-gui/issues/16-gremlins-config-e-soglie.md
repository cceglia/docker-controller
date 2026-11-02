# 16: gremlins: config e soglie

**What to build:** Ogni package di codice di produzione dello split ha il proprio file `.gremlins/<path>.yaml`, con il nome derivato dal path (`compose-model.yaml`, `gui-view.yaml`, …). `make mutate` esegue tutti i package e `make mutate PKG=…` ne esegue uno; `make test-ui` copre `./internal/gui/...`; `exclude-files` punta a `internal/compose/docker/fake/`. Le soglie sono misurate dopo lo split e pari al misurato meno 5 punti; un package senza mutanti è annotato, non inventato. `docker/fake` e `guitest` restano fuori da `MUTATE_PACKAGES`. Il README aggiorna tabella e percorsi. Contesto: spec `06-sottopackage-compose-e-gui`, ADR 0006.

**Blocked by:** 07 (compose: chiudere la radice), 15 (gui: chiudere la radice).

**Status:** ready-for-agent

- [ ] Esistono quindici file `.gremlins/<path>.yaml` con il blocco condiviso identico.
- [ ] `make mutate` gira tutti i package; `make mutate PKG=…` ne gira uno.
- [ ] Le soglie sono misurate e scritte; `docker/fake` e `guitest` restano esclusi.
- [ ] Un test verifica che `workers`, `timeout-coefficient` ed `exclude-files` siano identici in tutti i file.
- [ ] Il README riporta le nuove soglie e `./internal/gui/...`.
