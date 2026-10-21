# 04: Dettaglio del Progetto e azioni sui Servizi

**What to build:** selezionando un Progetto si apre il dettaglio, con i Container raggruppati per Servizio (immagine, stato, health, porte pubblicate). Ogni Servizio ha i pulsanti Start, Stop e Restart, che agiscono su tutte le sue repliche come un'Attività del Progetto. Una porta pubblicata apre `http://localhost:<porta>` nel browser. "Apri cartella" apre la cartella del Progetto nel file manager.

**Blocked by:** 02

**Status:** ready-for-agent

- [ ] Lo snapshot contiene i Container per Servizio con stato, health e porte (test con le fixture di `sunlinkdocker`)
- [ ] Un Servizio con repliche mostra tutti i suoi Container
- [ ] Le Azioni sul Servizio passano il nome del Servizio a compose (test sugli argomenti)
- [ ] Le Azioni sul Servizio rispettano la regola di una sola Attività per Progetto
- [ ] Porte cliccabili e "Apri cartella" funzionanti
