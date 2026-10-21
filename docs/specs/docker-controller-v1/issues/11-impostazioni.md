# 11: Impostazioni

**What to build:** una schermata Impostazioni permette di scegliere il comportamento alla chiusura della finestra (Nascondi nel tray come default, Esci, Chiedi ogni volta) e il comando del terminale (default `konsole --workdir {dir}`). Con "Chiedi ogni volta", la chiusura apre un dialogo con la casella "ricorda la scelta". Nel dettaglio del Progetto, "Apri terminale" usa il comando configurato nella cartella del Progetto.

**Blocked by:** 04, 10

**Status:** resolved

- [ ] Il modulo impostazioni restituisce i default su un DB nuovo e persiste le modifiche (test)
- [ ] Nascondi nel tray / Esci / Chiedi funzionano; "ricorda la scelta" salva l'impostazione
- [ ] "Apri terminale" sostituisce `{dir}` con la cartella del Progetto

## Comments

- 2026-10-21 (my-implement-orchestrator): implementato nel commit `46f0ac0`. Review finale (fixed point `4d05bd7`) approvata al ciclo 4. `make check` e `go test -race -tags gtk3 ./...` verdi. I criteri verificati da test automatici sono coperti; i criteri di UI, tray e desktop restano da verificare a mano con `wails3 dev` (l'app non è stata avviata per la verifica).
