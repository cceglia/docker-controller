# 05: Azioni e conferme

**What to build:** La riga delle Azioni lancia Up, Stop, Restart, Pull, Build, Down e Down con volumi sul Progetto, e Start, Stop, Restart e Down su ogni Servizio. Ogni Azione distruttiva chiede conferma in un dialogo modale, con il testo di oggi e il pulsante affermativo come predefinito. Un'Azione rifiutata dal Manager mostra l'errore utente in un avviso in sovrimpressione. Il doppio click non lancia due volte la stessa Azione.

**Blocked by:** 04

**Status:** ready-for-agent

- [ ] Il Progetto offre Up, Stop, Restart, Pull, Build, Down e Down con volumi
- [ ] Ogni Servizio offre Start, Stop, Restart e Down
- [ ] Un'Azione non ammessa è disabilitata, con il motivo visibile
- [ ] Down, Down di un Servizio e Down con volumi chiedono conferma con il testo di oggi
- [ ] Annullare la conferma non lancia niente
- [ ] Il pulsante affermativo del dialogo è predefinito, e Escape annulla
- [ ] Il dialogo restituisce il focus al controllo che lo ha aperto
- [ ] Un'Azione rifiutata mostra il messaggio utente in un avviso in sovrimpressione
- [ ] Durante l'Azione il comando che l'ha lanciata è disabilitato
- [ ] Il doppio click non produce due Attività
- [ ] Test da tavolo per l'elenco delle Azioni ammesse e per i testi di conferma
- [ ] Verifica manuale su KDE: ogni Azione arriva a Docker

## Comments

La decisione su quali Azioni sono ammesse resta nelle funzioni pure già coperte: `ProjectActions`, `ServiceActions`, `ActionDisabledReason`.
