# 02: Colonna dei Progetti

**What to build:** I Progetti diventano la navigazione. A sinistra una colonna con "Aggiungi progetto" in cima, i Progetti raggruppati per Gruppo e "Impostazioni" in fondo; ogni riga porta punto di stato, Nome visualizzato, nome compose e la striscia dei Container — piena se in esecuzione, vuota se fermo, in errore se `unhealthy`, con "+n" oltre dodici tacche. La riga del Progetto aperto è evidenziata. Scegliere un Progetto chiude le Impostazioni; senza selezione la tela mostra un riepilogo dei Progetti e l'invito a sceglierne uno. Sotto i 900px la colonna si riduce alle sole tacche. Il comando per tornare alla lista sparisce.

**Blocked by:** 01

**Status:** resolved

- [x] Si passa da un Progetto all'altro dalla colonna, senza tornare a una lista separata.
- [x] I Gruppi sono nello stesso ordine di oggi, con i Progetti senza Gruppo in una sezione propria; senza Gruppi la lista resta piatta.
- [x] Ogni riga mostra stato, nome visualizzato, nome compose e striscia dei Container.
- [x] La riga aperta si distingue senza ricorrere al colore e ha `aria-current` impostato.
- [x] "Aggiungi progetto" e "Impostazioni" sono raggiungibili da qualunque schermata.
- [x] A finestra stretta la colonna mostra le tacche e il resto dell'interfaccia resta usabile.
- [x] Scegliere un Progetto mentre sono aperte le Impostazioni mostra il dettaglio.
- [x] Senza selezione la tela mostra il riepilogo dei Progetti e come sceglierne uno, non una schermata vuota.
- [x] Il comando di ritorno alla lista non esiste più.
- [x] `svelte-check` e la build sono verdi.

## Comments

Implementato nel commit `a3de8aa`.

Revisione finale del 2026-10-31 con punto fisso `0ab9372`: nessun rilievo
bloccante su standard o spec. I rilievi non bloccanti sono nel report del run.
