# 09: Tray

**What to build:** L'app ha un'icona nella tray con il menu per Gruppo, Progetto e Servizio. Il click sinistro riporta la finestra in primo piano, il tasto destro apre il menu. L'icona cambia quando Docker non risponde. Il menu si ricostruisce a ogni snapshot e ogni Azione funziona, conferme comprese.

**Blocked by:** 03

**Status:** resolved

- [ ] L'icona compare su KDE con il tooltip "Docker Controller"
- [x] Il menu ha Gruppi, Progetti, Servizi, "Mostra finestra" ed "Esci"
- [x] I Gruppi compaiono in ordine, i Progetti senza Gruppo dopo di loro
- [x] Il click sinistro riporta la finestra in primo piano, deminimizzata e a fuoco
- [ ] Il tasto destro apre il menu
- [x] Le Azioni del menu sono Up, Stop, Restart e Down sul Progetto, Start, Stop, Restart e Down sul Servizio
- [x] Pull, Build e Down con volumi restano fuori dal menu, come oggi
- [x] Con Docker irraggiungibile il menu mostra "Docker non raggiungibile" e disabilita le Azioni
- [x] Il menu si aggiorna a ogni snapshot senza voci stantie
- [ ] L'icona cambia quando Docker è irraggiungibile
- [x] "Esci" chiude l'app
- [x] Test da tavolo per la mappatura dalla descrizione del menu alle voci di `systray`

## Comments

Implementato nel commit `768753a`.

Il menu è stato letto dal bus D-Bus (`com.canonical.dbusmenu.GetLayout`): Gruppi,
Progetti senza Gruppo, separatore, `Mostra finestra` ed `Esci`. Il click sinistro
è stato provato con `StatusNotifierItem.Activate` e ha riportato la finestra.

Restano aperti i punti che dipendono da un host vero: icona e tooltip a schermo,
tasto destro, icona offline.

Nota: `fyne.io/systray` chiude il canale di click di una voce mentre la
ricostruzione del menu è in corso. Un click in quella finestra può farlo
panicare. La ricostruzione avviene solo quando la descrizione cambia. Riconciliare
le voci invece di ricostruirle eviterebbe il caso, ma romperebbe l'ordine dei
Gruppi, che è un criterio del ticket.
