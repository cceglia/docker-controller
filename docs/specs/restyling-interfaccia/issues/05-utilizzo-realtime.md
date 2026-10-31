# 05: Utilizzo risorse in tempo reale per Servizio

**What to build:** Rete e disco di ogni Servizio smettono di essere totali dall'avvio dei Container e diventano valori istantanei al secondo, calcolati per Container e sommati per Servizio, nella stessa forma della striscia globale. Il rilevatore dei delta, che oggi restituisce solo il totale, espone i rate di ogni Container della fotogramma; il campionamento li somma per il totale globale e per ciascun Servizio, che guadagna i rate accanto alle misure. La riga delle misure del Servizio nasce qui: CPU con percentuale, core e indicatore sottile, memoria, rete in ingresso e uscita, disco in lettura e scrittura.

**Blocked by:** 04

**Status:** ready-for-agent

- [ ] Ogni Servizio con Container in esecuzione mostra CPU, memoria, rete e disco.
- [ ] Rete e disco sono valori al secondo fra due fotogrammi, con la stessa formattazione della striscia globale.
- [ ] Nessuna schermata mostra più un totale dall'avvio per rete o disco.
- [ ] Un Servizio senza Container in esecuzione mostra un trattino, non zero; "in attesa della prima misura" resta uno stato distinto.
- [ ] Il campione espone per ogni Servizio i rate accanto alle misure; il totale globale resta la somma di tutte le righe, comprese quelle di Progetti non salvati.
- [ ] Un Container alla sua prima fotogramma non produce rate; un contatore che cala non produce un rate negativo.
- [ ] I test sul seam esistente coprono: rate di un Servizio come somma delle repliche, Container riavviato, Container nuovo, Servizio senza Container, totale globale.
- [ ] `go test ./...` e `svelte-check` sono verdi.
