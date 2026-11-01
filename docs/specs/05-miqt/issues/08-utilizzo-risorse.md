# 08: Utilizzo risorse

**What to build:** Una barra in alto nel dettaglio mostra CPU, memoria, rete e disco della macchina. Ogni Servizio mostra l'utilizzo dei suoi Container con un indicatore e i valori numerici. Una macchina non raggiungibile mostra l'avviso al posto dei numeri. Le cifre si fermano quando la finestra si nasconde nella tray e ripartono quando torna visibile.

**Blocked by:** 04

**Status:** ready-for-agent

- [ ] La barra globale mostra CPU, memoria, rete e disco della macchina
- [ ] Ogni Servizio mostra l'utilizzo dei suoi Container con indicatore e valori
- [ ] Gli indicatori sono `QProgressBar` con il valore numerico accanto
- [ ] Un campione arrivato dopo lo stop non riporta cifre vecchie sullo schermo
- [ ] Docker non raggiungibile mostra un avviso invece di numeri falsi
- [ ] Un Servizio senza Container non mostra indicatore
- [ ] Il numero di Container non cambia l'altezza delle righe
- [ ] Test da tavolo per la formattazione di byte, percentuali, core e rate
- [ ] Il fermo e il riavvio alla visibilità della finestra si verificano nel ticket 13

## Comments

Le funzioni di formattazione restano pure e già coperte dai test. Il segnale di visibilità della finestra arriva nel ticket 13, insieme al comportamento di chiusura.
