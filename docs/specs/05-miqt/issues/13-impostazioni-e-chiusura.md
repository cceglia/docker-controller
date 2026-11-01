# 13: Impostazioni e chiusura

**What to build:** La finestra delle Impostazioni permette di scegliere cosa fa la chiusura della finestra: nascondi nella tray, chiedi ogni volta, esci. Il dialogo "Chiedi ogni volta" ha l'opzione di ricordare la scelta. Il comando del terminale si modifica e si valida. "Esci" chiude l'app e ferma gli stream. La chiusura nasconde la finestra e non la distrugge, e la riapertura ritrova il Progetto aperto e la cronologia delle Attività. Senza tray la chiusura esce, perché l'utente non potrebbe altrimenti riaprire la finestra. Le impostazioni restano nel database esistente.

**Blocked by:** 07, 08, 11

**Status:** ready-for-agent

- [ ] Le Impostazioni offrono le tre scelte di chiusura
- [ ] La scelta salvata vale al riavvio
- [ ] "Chiedi ogni volta" apre un dialogo con l'opzione di ricordare la scelta
- [ ] Il comando del terminale si valida e un comando non valido non si salva
- [ ] Chiudere la finestra la nasconde e non la distrugge, quando la scelta è nascondi
- [ ] Riaprire la finestra ritrova il Progetto aperto e la cronologia delle Attività
- [ ] Nascondere la finestra ferma il pannello Log e l'Utilizzo risorse
- [ ] L'output di un'Attività lanciata dalla tray continua con la finestra nascosta
- [ ] "Esci" chiude l'app e ferma ogni processo esterno
- [ ] Le impostazioni salvate restano leggibili dopo l'aggiornamento
- [ ] Senza tray la chiusura esce, e l'opzione nascondi non è disponibile
- [ ] I testi delle Impostazioni e del dialogo di chiusura non promettono più la perdita della cronologia
- [ ] Il test della macchina di chiusura copre il testo nuovo e il caso senza tray
- [ ] Test da tavolo per la validazione del comando del terminale
- [ ] Verifica manuale su KDE: le tre scelte di chiusura si comportano come il testo dice

## Comments

Il comportamento di chiusura cambia rispetto all'ADR 0003: la finestra non si distrugge più. Due testi utente e un test promettono ancora la perdita della cronologia delle Attività, e vanno riscritti qui.
