# 04: Attività con output dal vivo

**What to build:** Il pannello inferiore ha la scheda Attività. Mostra le Attività della sessione con stato e output che scorre mentre il comando gira. Un fallimento produce un toast con il messaggio di errore e un collegamento all'output. Chiudere il pannello non ferma l'Attività.

**Blocked by:** 03

**Status:** resolved

- [x] Pannello inferiore con la scheda Attività
- [x] Una nuova Attività compare subito con Progetto, Servizio e tipo di Azione
- [x] L'output arriva in tempo reale mentre il comando gira
- [x] Un'Attività conclusa mostra l'esito positivo o l'errore
- [x] Lo storico delle Attività della sessione resta consultabile
- [x] Un fallimento mostra un toast con il messaggio italiano
- [ ] Con 5.000 righe di output, un clic su una riga la seleziona entro un secondo
- [x] Selezionare un'Attività mostra il suo output
- [x] Chiudere il pannello non termina l'Attività in corso

## Comments

Implementato nel commit `14e163b`, con le correzioni del primo ciclo di
revisione in `cb6e845`.

Il pannello si apre da solo quando compare un'Attività nuova. La conferma di
Down dalla tray non costruisce più la finestra per ospitare il dialogo.

Resta aperto il punto sulle 5000 righe: la selezione non è stata misurata con
5000 righe reali, solo la costruzione del buffer.
