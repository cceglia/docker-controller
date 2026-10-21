# 07: Nome visualizzato e Gruppo

**What to build:** per ogni Progetto si possono impostare o togliere un Nome visualizzato e un Gruppo, scelto tra quelli esistenti o creato al momento. La lista mostra il Nome visualizzato (con il nome compose in secondo piano) ed è raggruppata per Gruppo, con i Progetti senza Gruppo in una sezione a parte.

**Blocked by:** 01

**Status:** ready-for-agent

- [ ] Un'unica operazione del modulo `compose` aggiorna le impostazioni del Progetto
- [ ] Nome visualizzato e Gruppo sopravvivono alla riapertura del DB e al Down (test)
- [ ] L'eliminazione di Nome visualizzato o Gruppo riporta al comportamento predefinito (test)
- [ ] Lista raggruppata per Gruppo nella UI, con il Gruppo modificabile scegliendo un esistente o creandone uno nuovo
