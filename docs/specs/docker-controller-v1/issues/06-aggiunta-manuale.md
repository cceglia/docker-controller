# 06: Aggiunta manuale di un Progetto

**What to build:** con un pulsante "Aggiungi Progetto" si sceglie il File compose principale tramite file picker. Il nome compose viene proposto come farebbe compose (dal `name:` nel file oppure dal nome della cartella) ed è modificabile. Si possono aggiungere altri File compose, per esempio un override. Il Progetto aggiunto compare come Rimosso e Up lo avvia usando tutti i suoi File compose. Un nome già salvato viene rifiutato con un messaggio chiaro.

**Blocked by:** 02

**Status:** ready-for-agent

- [ ] Il nome proposto rispetta `name:` e altrimenti usa il nome della cartella (test)
- [ ] Le Azioni su un Progetto aggiunto passano tutti i `-f` nell'ordine salvato, più `--project-directory` (test)
- [ ] Un nome duplicato viene rifiutato (test)
- [ ] Il Progetto aggiunto sopravvive alla riapertura del DB (test)
- [ ] Dialogo con file picker, nome modificabile e lista dei File compose
