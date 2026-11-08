# 03: L'ordine dei Gruppi si sceglie a mano

**What to build:** Nel dialogo Gruppi l'ordine si decide con "Sposta su" e "Sposta giù", attivi sulla voce selezionata e disponibili da tastiera con Alt+Su e Alt+Giù. Sposta su è spento sul primo Gruppo. Sposta giù è spento sull'ultimo. Mentre il campo del nome è aperto, i comandi di spostamento e le scorciatoie sono spenti. Spostare un Gruppo cambia l'ordine di lista, tray e campo Gruppo del form. La selezione resta sul Gruppo spostato. L'ordine sopravvive alla chiusura dell'app e resta valido dopo un'eliminazione: i vicini si spostano ancora e l'ordine resta quello atteso.

**Blocked by:** 01, 02

**Status:** resolved

- [x] "Sposta su" e "Sposta giù" spostano il Gruppo selezionato di un posto; Sposta su è spento sul primo Gruppo e Sposta giù sull'ultimo, mentre sull'ultimo Sposta su resta attivo e sul primo Sposta giù resta attivo
- [x] Alt+Su e Alt+Giù spostano il Gruppo selezionato senza usare il mouse
- [x] Senza selezione, Sposta su e Sposta giù sono spenti
- [x] Sposta su e Sposta giù entrano nell'ordine di tabulazione dopo Elimina e prima di Chiudi
- [x] Mentre il campo del nome è aperto, Sposta su, Sposta giù, Alt+Su e Alt+Giù sono spenti
- [x] Spostare un Gruppo cambia l'ordine della lista, del tray e del campo Gruppo del form, e la selezione resta sul Gruppo spostato
- [x] L'ordine sopravvive alla chiusura e alla riapertura dell'app
- [x] Dopo l'eliminazione di un Gruppo, i suoi vicini si spostano ancora e l'ordine resta quello atteso
- [x] Spostare un Gruppo non salvato restituisce l'errore Gruppo sconosciuto, mostra il messaggio italiano nella riga sotto l'elenco e lascia il dialogo aperto, senza cambiare l'ordine
- [x] Durante un'operazione Sposta su, Sposta giù, Alt+Su e Alt+Giù sono ignorati
- [x] Test: spostamento, bordi e Gruppo sconosciuto nel Manager, ordine conservato dopo la riapertura del database, ordine delle sezioni e menu del tray dopo lo spostamento, spostamento con Docker irraggiungibile, passo offscreen con Sposta su, Sposta giù, Alt+Su e Alt+Giù, comandi spenti sul primo, sull'ultimo e senza selezione, campo del form nell'ordine nuovo, selezione sul Gruppo spostato, intestazioni della lista e ordine del tray dopo lo spostamento, e campo del nome aperto
- [x] `make check` e `make test-ui` verdi

## Comments

Implementato in `2cb92a8`. La review finale non ha trovato difetti bloccanti in questo ticket. `make check` e `make test-ui` verdi; lo spostamento è coperto dal passo offscreen e dai test del Manager.
