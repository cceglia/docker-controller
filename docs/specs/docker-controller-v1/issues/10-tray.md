# 10: Tray

**What to build:** un'icona nel system tray, che cambia aspetto quando Docker non è raggiungibile, con un menu annidato Gruppo ▸ Progetto (i Progetti senza Gruppo al primo livello). Ogni Progetto mostra lo stato (● Attivo, ◐ Parziale, ○ Fermo/Rimosso) e offre Up, Stop, Restart e il sottomenu Servizi ▸ Servizio ▸ Start/Stop/Restart. Durante un'Attività le voci di quel Progetto sono disabilitate. Una notifica desktop segnala la fine di un'Attività lanciata dal tray e ogni Attività fallita. Il menu ha anche "Mostra finestra" ed "Esci". Down, Down con volumi, Pull e Build non sono nel tray.

**Blocked by:** 03, 04, 07

**Status:** ready-for-agent

- [ ] Il menu si ricostruisce a ogni snapshot di Watch, senza polling proprio
- [ ] Struttura annidata per Gruppo, Progetto e Servizi come descritto
- [ ] Voci disabilitate durante un'Attività e quando Docker non è raggiungibile
- [ ] Notifiche desktop per le Attività lanciate dal tray e per gli errori
- [ ] Icona diversa quando Docker non è raggiungibile
- [ ] Verifica manuale su KDE: Restart di `personalvpn` dal tray con la finestra nascosta mostra la notifica
