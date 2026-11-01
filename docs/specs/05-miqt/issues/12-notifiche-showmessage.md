# 12: Notifiche con ShowMessage

**What to build:** Le notifiche desktop passano a `QSystemTrayIcon.ShowMessage`, con titolo "Docker Controller" e il testo di oggi. La decisione su cosa notificare non cambia: la fine di un'Attività lanciata dalla tray e ogni Attività fallita. Una notifica non compare due volte per la stessa Attività. Il click sulla notifica mostra la finestra. Il trasporto D-Bus esce dal repository.

**Blocked by:** 06, 11

**Status:** ready-for-agent

- [ ] La fine di un'Attività lanciata dalla tray produce una notifica
- [ ] Un'Attività fallita produce una notifica, da qualunque origine
- [ ] Il titolo della notifica è "Docker Controller" e il testo è quello di oggi
- [ ] Una stessa Attività non produce due notifiche
- [ ] Il click sulla notifica mostra la finestra
- [ ] Senza tray non arriva nessuna notifica desktop e l'app resta usabile
- [ ] Senza tray un'Attività fallita mostra l'avviso nella finestra
- [ ] Il mittente nasce dopo `QApplication` e vive sul thread Qt
- [ ] La notifica si invia dal thread Qt con `mainthread.Start`
- [ ] Le tre chiamate restano quelle di `tray.Notifier`: `Launching`, `Aborted` e `Observe`
- [ ] La decisione su cosa notificare resta nel package puro della tray
- [ ] Il package di trasporto D-Bus esce dal repository con i suoi test
- [ ] Test esistenti del Notifier restano verdi con un mittente finto
- [ ] Verifica manuale su KDE: la notifica arriva con l'icona giusta

## Comments

La regola "fine di un'Attività lanciata dalla tray, più ogni fallimento" resta in `internal/tray`, con i test di oggi.
