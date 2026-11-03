# 12: gui: estrarre activity

**What to build:** Log delle Attività, testi e modello del pannello vivono in `internal/gui/activity`. `gui.Run` riceve `activity.Log` e `main.go` costruisce quel log; il pannello Attività e Log usa il package. I test si spostano (`activitylog_test`, `activitytext_test`, `bottompanel_test`) e usano `guitest`. Contesto: spec `06-sottopackage-compose-e-gui`, ADR 0006.

**Blocked by:** 02 (gui: creare guitest), 10 (gui: estrarre project).

**Status:** resolved

- [x] `activity` importa `project` e `view` secondo necessità, non la radice.
- [x] `gui.Run` e `main.go` usano `activity.Log`; l'app compila.
- [x] I test delle Attività e del pannello vivono in `activity` e usano `guitest`.
- [x] `go build ./...`, `go test ./...`, `go vet ./...` e `make test-ui` sono verdi.
