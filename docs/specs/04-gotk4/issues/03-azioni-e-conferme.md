# 03: Azioni e conferme

**What to build:** Dal dettaglio si lanciano le Azioni sul Progetto (Up, Stop, Restart, Pull, Build, Down, Down con volumi) e sul Servizio (Start, Stop, Restart, Down). Le Azioni distruttive chiedono conferma con il testo di oggi. Durante un'Attività del Progetto i pulsanti sono disabilitati, così come quando Docker non risponde.

**Blocked by:** 02

**Status:** ready-for-agent

- [ ] Tutte le Azioni del Progetto sono disponibili e lanciano quella giusta
- [ ] Tutte le Azioni del Servizio sono disponibili e lanciano quella giusta
- [ ] Down, Down di un Servizio e Down con volumi chiedono conferma prima di partire
- [ ] Annullare la conferma non lancia nulla e non cambia la vista
- [ ] Durante un'Attività sul Progetto i pulsanti di quel Progetto sono disabilitati
- [ ] Con Docker irraggiungibile le Azioni sono disabilitate e la vista lo dice
- [ ] Un errore del Manager compare come messaggio italiano comprensibile
- [ ] Verifica manuale: Up e poi Stop su un Progetto cambiano il suo stato in lista
