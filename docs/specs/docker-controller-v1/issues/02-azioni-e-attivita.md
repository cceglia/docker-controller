# 02: Azioni non distruttive e Attività

**What to build:** dalla lista si possono lanciare Up, Stop, Restart, Pull e Build su un Progetto. Ogni lancio crea un'Attività: il suo output compare in tempo reale nel pannello Attività, che conserva lo storico della sessione. Se l'Attività fallisce, compare un toast con l'errore e un link all'output. Su un Progetto può esserci una sola Attività alla volta (pulsanti disabilitati e indicatore "⏳ <Azione> in corso"); Progetti diversi lavorano in parallelo. Alla fine di ogni Attività lo snapshot si aggiorna.

**Blocked by:** 01

**Status:** resolved

- [ ] Un'unica operazione del modulo `compose` esegue un'Azione e restituisce l'id dell'Attività, consegnando l'output riga per riga e l'esito
- [ ] Build corrisponde a `up -d --build`; Up a `up -d` (verificato sugli argomenti registrati dal fake)
- [ ] Una seconda Attività sullo stesso Progetto viene rifiutata; su Progetti diversi parte (test)
- [ ] In caso di errore l'esito contiene lo stderr di compose (test)
- [ ] Watch notifica subito dopo la fine di un'Attività (test)
- [ ] Il pannello Attività mostra l'output in tempo reale e lo storico della sessione
- [ ] Verifica manuale: Stop e poi Up su `stellantisvpn` cambiano lo stato in lista

## Comments

- 2026-10-21 (my-implement-orchestrator): implementato nel commit `913627f`, correzioni dalla review finale in `91274a2`. Review finale (fixed point `4d05bd7`) approvata al ciclo 4. `make check` e `go test -race -tags gtk3 ./...` verdi. I criteri verificati da test automatici sono coperti; i criteri di UI, tray e desktop restano da verificare a mano con `wails3 dev` (l'app non è stata avviata per la verifica).
