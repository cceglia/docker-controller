# 10: Notifiche

**What to build:** Ogni Attività lanciata dalla tray e ogni Attività fallita produce una notifica desktop con titolo "Docker Controller" e il testo di oggi. La notifica arriva una volta sola per Attività. Se il servizio di notifica non risponde, l'app non si blocca.

**Blocked by:** 09

**Status:** resolved

- [ ] Notifica a fine di un'Attività lanciata dalla tray, riuscita o fallita
- [ ] Notifica per ogni Attività fallita, anche lanciata dalla finestra
- [ ] Nessuna notifica per un'Attività riuscita lanciata dalla finestra
- [x] Una sola notifica per Attività
- [x] Titolo e testo coincidono con quelli di oggi
- [ ] Se il servizio di notifica non risponde, l'app registra l'errore e continua
- [x] I test del `Notifier` restano invariati e verdi

## Comments

Implementato nel commit `cd0a676`.

Il glue alimenta `tray.Notifier` con `Launching`, `Aborted` e `Observe`, e il
trasporto D-Bus gira fuori dal loop GTK con un timeout.

Restano aperti i punti che richiedono un servizio di notifica vero: comparsa
della notifica, testo a schermo, comportamento senza servizio. La decisione e
l'invio sono coperti da test con un `Sender` finto, e i test del `Notifier`
restano invariati.
