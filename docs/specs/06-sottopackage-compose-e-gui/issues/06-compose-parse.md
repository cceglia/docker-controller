# 06: compose: estrarre parse

**What to build:** Il parsing dell'output della CLI vive in `internal/compose/parse`: elenco Progetti, stato Container, Log, Utilizzo risorse, Profili e dimensione macchina, con i tipi di appoggio `Machine` e `StatsRow`. I test sono table-driven e leggono le fixture esistenti da `../testdata`. `Manager` usa le funzioni esportate. Contesto: spec `06-sottopackage-compose-e-gui`, ADR 0006.

**Blocked by:** 01 (compose: estrarre model).

**Status:** ready-for-agent

- [ ] `parse` contiene i parser e non importa la radice.
- [ ] I test di `parse` coprono i formati reali di `compose ls`, `compose ps`, Log, Utilizzo risorse e Profili.
- [ ] Nessuna fixture viene duplicata.
- [ ] `go build ./...`, `go test ./...` e `go vet ./...` sono verdi.
