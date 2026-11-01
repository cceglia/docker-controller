# 09: Modifica Progetto

**What to build:** Dal dettaglio l'utente apre un form per Nome visualizzato, Gruppo e Profili attivi. Il form usa i valori salvati, propone i Gruppi già in uso e salva dallo stesso pannello. Un salvataggio non valido mostra l'errore sotto il campo. Uno snapshot che arriva durante la modifica non cancella quello che l'utente ha scritto.

**Blocked by:** 04

**Status:** ready-for-agent

- [ ] Il pulsante di modifica apre il form con i valori salvati
- [ ] Il form modifica Nome visualizzato, Gruppo e Profili attivi
- [ ] I Gruppi già in uso sono proposti come scelta
- [ ] Il salvataggio aggiorna la lista e il dettaglio senza riavviare l'app
- [ ] Annullare riporta i valori salvati
- [ ] Un Nome visualizzato vuoto torna al nome compose
- [ ] Un campo non valido mostra il messaggio sotto il campo e il nome accessibile lo ripete
- [ ] Uno snapshot durante la modifica non sovrascrive il testo digitato
- [ ] Lo snapshot aggiorna il Progetto dopo il salvataggio
- [ ] Test da tavolo per la normalizzazione di Nome visualizzato, Gruppo e Profili
- [ ] Verifica manuale su KDE: il form si usa solo con la tastiera

## Comments

La normalizzazione dei valori resta pura e già coperta dai test.
