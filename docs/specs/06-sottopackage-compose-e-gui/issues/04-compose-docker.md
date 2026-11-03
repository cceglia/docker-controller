# 04: compose: estrarre docker

**What to build:** La porta verso la CLI Docker, l'adattatore che esegue i processi e il doppio di test vivono in `internal/compose/docker`, con il doppio sotto `docker/fake`. La radice riesporta `ExecDocker` con un alias, così `main.go` non cambia. Il test dell'adattatore si sposta con l'adattatore. Contesto: spec `06-sottopackage-compose-e-gui`, ADR 0006.

**Blocked by:** 01 (compose: estrarre model).

**Status:** resolved

- [x] `docker` contiene la porta, `ExecDocker` e `fake`; la radice non esegue più processi.
- [x] `main.go` compila senza cambiare il proprio import.
- [x] Il test dell'adattatore si sposta in `docker` e passa.
- [x] `go build ./...`, `go test ./...` e `go vet ./...` sono verdi.
