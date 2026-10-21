# 08: Profili

**What to build:** il dettaglio del Progetto mostra i Profili disponibili come interruttori. I Profili attivi vengono ricordati e usati da Up, Stop, Restart, Pull, Build e dalle Azioni sui Servizi. Down e Down con volumi agiscono invece su tutti i Profili. Cambiare un interruttore non avvia né ferma nulla.

**Blocked by:** 02

**Status:** ready-for-agent

- [ ] I Profili disponibili sono letti da `compose config --profiles` ed esposti nello snapshot (test)
- [ ] I Profili attivi sono passati come `--profile` alle Azioni (test sugli argomenti)
- [ ] Down e Down con volumi passano tutti i Profili (test sugli argomenti)
- [ ] I Profili attivi sopravvivono alla riapertura del DB (test)
- [ ] Cambiare i Profili non lancia alcuna Attività (test)
- [ ] Interruttori nella UI
