# 07: Log

**What to build:** Il tab Log mostra i log di un Progetto o di un Servizio in un `QPlainTextEdit` in sola lettura, con le righe colorate per Servizio e gli errori distinti. Il pannello ha ricerca, pausa dello scorrimento, mostra o nascondi i timestamp, tiene al massimo 5.000 righe e permette di selezionare e copiare il testo. Chiudere il pannello termina il processo `docker compose logs`.

**Blocked by:** 06

**Status:** ready-for-agent

- [ ] Il tab Log mostra i log del Progetto selezionato
- [ ] Un Servizio del dettaglio apre i log del solo Servizio
- [ ] Ogni Servizio ha un colore stabile, preso dalla palette dei ruoli
- [ ] Le righe di stderr si distinguono da quelle di stdout
- [ ] La ricerca evidenzia le righe che corrispondono e conta i risultati
- [ ] La pausa ferma lo scorrimento e riprendere torna all'ultima riga
- [ ] Il pulsante dei timestamp mostra e nasconde l'ora di ogni riga
- [ ] Il pannello tiene al massimo 5.000 righe e scarta le più vecchie
- [ ] Il testo si seleziona e si copia dal widget
- [ ] Chiudere il pannello o cambiare Progetto termina il processo dei log
- [ ] I colori dei Servizi entrano nel test dei token del ticket 03
- [ ] Test da tavolo per la costruzione dei tag del log e per il tetto delle righe
- [ ] Verifica manuale su KDE: i colori restano leggibili in tema chiaro e scuro

## Comments

La macchina degli stream resta nel package della GUI senza Qt, con i test di oggi. Il fermo degli stream alla chiusura della finestra arriva nel ticket 13, quando esiste il nascondi.
