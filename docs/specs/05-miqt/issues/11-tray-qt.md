# 11: Tray Qt

**What to build:** L'icona di sistema passa a `QSystemTrayIcon`. Il menu ha le voci per Gruppo, Progetto e Servizio, si aggiorna a ogni snapshot, cambia icona quando Docker non risponde e mostra un tooltip che dice "Docker Controller" e lo stato di Docker. Il click sinistro mostra la finestra, il tasto destro apre il menu. In fondo ci sono "Mostra finestra" ed "Esci", e "Esci" chiude l'applicazione. Quando la tray non è disponibile l'app funziona lo stesso e la finestra offre "Esci" e le Impostazioni.

**Blocked by:** 05

**Status:** ready-for-agent

- [ ] L'app mostra un'icona di sistema con `QSystemTrayIcon`
- [ ] Il menu ha le voci per Gruppo, Progetto e Servizio, con le stesse gerarchie di oggi
- [ ] Il menu si ricostruisce solo quando l'albero delle voci cambia
- [ ] Il click sinistro mostra la finestra e le porta il focus
- [ ] Il tasto destro apre il menu
- [ ] Su Wayland il menu lo disegna il desktop tramite `com.canonical.dbusmenu`: il popup Qt non serve
- [ ] L'icona cambia quando Docker non risponde
- [ ] Il tooltip dice "Docker Controller" e ripete lo stato di Docker a parole
- [ ] Le Azioni della tray funzionano, comprese le conferme di un'Azione distruttiva
- [ ] "Mostra finestra" chiude il menu e mostra la finestra
- [ ] "Esci" chiude l'applicazione con `QApplication_Quit`
- [ ] Senza tray l'app parte lo stesso, e la finestra offre "Esci" e le Impostazioni
- [ ] L'avvio controlla la disponibilità della tray prima di crearla
- [ ] Test da tavolo per l'albero delle voci della tray e per il confronto fra due alberi
- [ ] Verifica manuale su KDE, su X11 e su Wayland: icona, menu, i due tasti e un menu aperto durante uno snapshot

## Comments

Il package della tray resta puro: la descrizione del menu non cambia. Cambia solo l'adapter che la applica. Il fermo degli stream all'uscita arriva nel ticket 13.
