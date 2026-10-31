# 02: Colonna dei Progetti

**What to build:** I Progetti diventano la navigazione. A sinistra una colonna con "Aggiungi progetto" in cima, i Progetti raggruppati per Gruppo e "Impostazioni" in fondo; ogni riga porta punto di stato, Nome visualizzato, nome compose e la striscia dei Container — piena se in esecuzione, vuota se fermo, in errore se `unhealthy`, con "+n" oltre dodici tacche. La riga del Progetto aperto è evidenziata. Scegliere un Progetto chiude le Impostazioni; senza selezione la tela mostra un riepilogo dei Progetti e l'invito a sceglierne uno. Sotto i 900px la colonna si riduce alle sole tacche. Il comando per tornare alla lista sparisce.

**Blocked by:** 01

**Status:** ready-for-agent

- [ ] Si passa da un Progetto all'altro dalla colonna, senza tornare a una lista separata.
- [ ] I Gruppi sono nello stesso ordine di oggi, con i Progetti senza Gruppo in una sezione propria; senza Gruppi la lista resta piatta.
- [ ] Ogni riga mostra stato, nome visualizzato, nome compose e striscia dei Container.
- [ ] La riga aperta si distingue senza ricorrere al colore e ha `aria-current` impostato.
- [ ] "Aggiungi progetto" e "Impostazioni" sono raggiungibili da qualunque schermata.
- [ ] A finestra stretta la colonna mostra le tacche e il resto dell'interfaccia resta usabile.
- [ ] Scegliere un Progetto mentre sono aperte le Impostazioni mostra il dettaglio.
- [ ] Senza selezione la tela mostra il riepilogo dei Progetti e come sceglierne uno, non una schermata vuota.
- [ ] Il comando di ritorno alla lista non esiste più.
- [ ] `svelte-check` e la build sono verdi.
