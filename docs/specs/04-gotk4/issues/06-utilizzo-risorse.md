# 06: Utilizzo risorse

**What to build:** Una barra sopra il dettaglio mostra CPU, memoria, rete e disco della macchina. Ogni Servizio mostra il proprio Utilizzo risorse con meter e valori. Lo stream parte con la finestra visibile e si ferma quando la finestra si nasconde nella tray. Con Docker irraggiungibile compare un avviso invece di numeri falsi.

**Blocked by:** 02

**Status:** resolved

- [x] Barra globale con CPU, memoria, rete in ingresso e uscita, disco in lettura e scrittura
- [ ] Ogni Servizio mostra CPU, memoria e rate con meter
- [x] Un Servizio senza Container in esecuzione mostra "—" e non zero
- [x] La formattazione coincide con quella di oggi: MiB e GiB, MB/s, core, percentuali
- [x] Lo stream si ferma quando la finestra si nasconde e riparte quando torna
- [x] Nessun valore vecchio resta a schermo dopo lo stop
- [x] Con Docker irraggiungibile compare un avviso e la barra resta vuota
- [x] Test da tavolo per byte, percentuali, core e rate
- [ ] Verifica manuale: i numeri si muovono e somigliano a `docker stats`

## Comments

Implementato nel commit `acdffe8`.

Il meter del Servizio copre la CPU, come la vista di oggi; memoria e rate sono
valori. Il punto 2 del ticket resta aperto per questo: la lettera chiede un
meter per ogni misura, la parità no.

Resta aperta la verifica manuale dei numeri contro `docker stats`.
