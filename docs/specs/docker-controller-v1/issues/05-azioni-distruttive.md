# 05: Azioni distruttive

**What to build:** nel dettaglio compaiono Down, con un modal di conferma, e un pulsante rosso separato Down con volumi, il cui modal richiede di digitare il nome del Progetto. Un Progetto Rimosso può essere tolto con "Rimuovi dalla lista" (anche qui con modal), che non tocca Docker. Se i File compose di un Progetto non esistono più, il Progetto mostra l'avviso "File compose mancante" e "Rimuovi dalla lista" viene messo in evidenza.

**Blocked by:** 02

**Status:** ready-for-agent

- [ ] Down corrisponde a `down`, Down con volumi a `down -v` (test sugli argomenti)
- [ ] Dopo un Down il Progetto resta in lista come Rimosso (test)
- [ ] "Rimuovi dalla lista" viene rifiutato se il Progetto ha Container (test)
- [ ] File compose mancante viene rilevato nello snapshot (test con un file temporaneo cancellato)
- [ ] Modal di conferma per tutte e tre le Azioni distruttive; Down con volumi si abilita solo dopo aver digitato il nome esatto
