# 09: Tray

**What to build:** L'app ha un'icona nella tray con il menu per Gruppo, Progetto e Servizio. Il click sinistro riporta la finestra in primo piano, il tasto destro apre il menu. L'icona cambia quando Docker non risponde. Il menu si ricostruisce a ogni snapshot e ogni Azione funziona, conferme comprese.

**Blocked by:** 03

**Status:** ready-for-agent

- [ ] L'icona compare su KDE con il tooltip "Docker Controller"
- [ ] Il menu ha Gruppi, Progetti, Servizi, "Mostra finestra" ed "Esci"
- [ ] I Gruppi compaiono in ordine, i Progetti senza Gruppo dopo di loro
- [ ] Il click sinistro riporta la finestra in primo piano, deminimizzata e a fuoco
- [ ] Il tasto destro apre il menu
- [ ] Le Azioni del menu sono Up, Stop, Restart e Down sul Progetto, Start, Stop, Restart e Down sul Servizio
- [ ] Pull, Build e Down con volumi restano fuori dal menu, come oggi
- [ ] Con Docker irraggiungibile il menu mostra "Docker non raggiungibile" e disabilita le Azioni
- [ ] Il menu si aggiorna a ogni snapshot senza voci stantie
- [ ] L'icona cambia quando Docker è irraggiungibile
- [ ] "Esci" chiude l'app
