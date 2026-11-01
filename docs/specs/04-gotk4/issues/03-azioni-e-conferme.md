# 03: Azioni e conferme

**What to build:** Dal dettaglio si lanciano le Azioni sul Progetto (Up, Stop, Restart, Pull, Build, Down, Down con volumi) e sul Servizio (Start, Stop, Restart, Down). Le Azioni distruttive chiedono conferma con il testo di oggi. Durante un'Attività del Progetto i pulsanti sono disabilitati, così come quando Docker non risponde.

**Blocked by:** 02

**Status:** resolved

- [x] Tutte le Azioni del Progetto sono disponibili e lanciano quella giusta
- [x] Tutte le Azioni del Servizio sono disponibili e lanciano quella giusta
- [x] Down, Down di un Servizio e Down con volumi chiedono conferma prima di partire
- [x] Annullare la conferma non lancia nulla e non cambia la vista
- [x] Durante un'Attività sul Progetto i pulsanti di quel Progetto sono disabilitati
- [x] Con Docker irraggiungibile le Azioni sono disabilitate e la vista lo dice
- [x] Un errore del Manager compare come messaggio italiano comprensibile
- [ ] Verifica manuale: Up e poi Stop su un Progetto cambiano il suo stato in lista

## Comments

Implementato nel commit `ddb06c6`.

L'Azione principale segue lo stato, come la vista vecchia: Up mentre Fermo o
Rimosso, Stop mentre Attivo, Restart mentre Parziale. Down con volumi richiede
il nome compose digitato, come oggi.

Verificato: test, build e schermate della conferma e del banner di Docker
irraggiungibile.
