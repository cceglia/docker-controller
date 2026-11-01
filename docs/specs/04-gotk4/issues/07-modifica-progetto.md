# 07: Modifica del Progetto

**What to build:** Dal dettaglio si modificano Nome visualizzato, Gruppo e Profili attivi del Progetto. Il salvataggio aggiorna dettaglio e lista senza chiudere la finestra. L'annullamento lascia tutto com'è.

**Blocked by:** 02

**Status:** resolved

- [x] Nome visualizzato modificabile e cancellabile
- [x] Gruppo scegliibile tra quelli esistenti o lasciabile vuoto
- [x] Profili attivi selezionabili tra quelli disponibili per il Progetto
- [x] Il salvataggio aggiorna il dettaglio e la lista
- [x] L'annullamento riporta i valori precedenti
- [x] Un errore di salvataggio lascia i valori precedenti e mostra il messaggio
- [x] Un Nome visualizzato vuoto torna al nome compose

## Comments

Implementato nel commit `288e879`, con il Gruppo nuovo in `288e879`.

Il Gruppo ammette la voce `Nuovo Gruppo…` con il campo di testo, come la vista
vecchia, per decisione esplicita dell'utente. È un'aggiunta rispetto alla
lettera del ticket.
