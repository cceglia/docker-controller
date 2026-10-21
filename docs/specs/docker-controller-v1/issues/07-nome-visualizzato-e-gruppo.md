# 07: Nome visualizzato e Gruppo

**What to build:** per ogni Progetto si possono impostare o togliere un Nome visualizzato e un Gruppo, scelto tra quelli esistenti o creato al momento. La lista mostra il Nome visualizzato (con il nome compose in secondo piano) ed è raggruppata per Gruppo, con i Progetti senza Gruppo in una sezione a parte.

**Blocked by:** 01

**Status:** resolved

- [ ] Un'unica operazione del modulo `compose` aggiorna le impostazioni del Progetto
- [ ] Nome visualizzato e Gruppo sopravvivono alla riapertura del DB e al Down (test)
- [ ] L'eliminazione di Nome visualizzato o Gruppo riporta al comportamento predefinito (test)
- [ ] Lista raggruppata per Gruppo nella UI, con il Gruppo modificabile scegliendo un esistente o creandone uno nuovo

## Comments

- 2026-10-21 (my-implement-orchestrator): implementato nel commit `920591f`. Review finale (fixed point `4d05bd7`) approvata al ciclo 4. `make check` e `go test -race -tags gtk3 ./...` verdi. I criteri verificati da test automatici sono coperti; i criteri di UI, tray e desktop restano da verificare a mano con `wails3 dev` (l'app non è stata avviata per la verifica).
