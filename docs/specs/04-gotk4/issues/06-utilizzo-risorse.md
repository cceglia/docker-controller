# 06: Utilizzo risorse

**What to build:** Una barra sopra il dettaglio mostra CPU, memoria, rete e disco della macchina. Ogni Servizio mostra il proprio Utilizzo risorse con meter e valori. Lo stream parte con la finestra visibile e si ferma quando la finestra si nasconde nella tray. Con Docker irraggiungibile compare un avviso invece di numeri falsi.

**Blocked by:** 02

**Status:** ready-for-agent

- [ ] Barra globale con CPU, memoria, rete in ingresso e uscita, disco in lettura e scrittura
- [ ] Ogni Servizio mostra CPU, memoria e rate con meter
- [ ] Un Servizio senza Container in esecuzione mostra "—" e non zero
- [ ] La formattazione coincide con quella di oggi: MiB e GiB, MB/s, core, percentuali
- [ ] Lo stream si ferma quando la finestra si nasconde e riparte quando torna
- [ ] Nessun valore vecchio resta a schermo dopo lo stop
- [ ] Con Docker irraggiungibile compare un avviso e la barra resta vuota
- [ ] Verifica manuale: i numeri si muovono e somigliano a `docker stats`
