# 10: Notifiche

**What to build:** Ogni Attività lanciata dalla tray e ogni Attività fallita produce una notifica desktop con titolo "Docker Controller" e il testo di oggi. La notifica arriva una volta sola per Attività. Se il servizio di notifica non risponde, l'app non si blocca.

**Blocked by:** 09

**Status:** ready-for-agent

- [ ] Notifica a fine di un'Attività lanciata dalla tray, riuscita o fallita
- [ ] Notifica per ogni Attività fallita, anche lanciata dalla finestra
- [ ] Nessuna notifica per un'Attività riuscita lanciata dalla finestra
- [ ] Una sola notifica per Attività
- [ ] Titolo e testo coincidono con quelli di oggi
- [ ] Se il servizio di notifica non risponde, l'app registra l'errore e continua
- [ ] I test del `Notifier` restano invariati e verdi
