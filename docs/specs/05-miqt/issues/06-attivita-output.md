# 06: Attività e output

**What to build:** Il pannello inferiore ha due tab, Attività e Log. Il tab Attività elenca le Attività in corso e concluse, con esito e output. L'output cresce mentre il comando gira. Un'Attività fallita mostra il messaggio di errore in un avviso in sovrimpressione con il collegamento al suo output.

**Blocked by:** 05

**Status:** ready-for-agent

- [ ] Il pannello inferiore usa un `QSplitter` verticale con i due tab
- [ ] Il tab Attività mostra le Attività in corso e quelle concluse
- [ ] Ogni Attività mostra il suo esito e il suo output
- [ ] L'output si aggiorna mentre il comando gira
- [ ] Un'Attività fallita mostra il messaggio utente in un avviso in sovrimpressione
- [ ] L'avviso porta all'output dell'Attività fallita
- [ ] Ogni Progetto ha al massimo un'Attività alla volta, come oggi
- [ ] L'elenco si aggiorna a ogni snapshot senza perdere lo scorrimento
- [ ] Test da tavolo per l'esito di un'Attività e per il testo dell'Attività fallita
- [ ] Verifica manuale su KDE: un'Attività lunga si segue dall'inizio alla fine

## Comments

La macchina delle Attività resta nel package della GUI senza Qt, con i test di oggi.
