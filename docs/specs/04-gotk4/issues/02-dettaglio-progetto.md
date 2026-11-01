# 02: Dettaglio del Progetto

**What to build:** La colonna di destra mostra il Progetto selezionato: nome, stato, File compose e Profili attivi. Sotto compaiono i Servizi, ognuno con i suoi Container: nome, immagine, stato, salute e porte pubblicate. Tutto si aggiorna a ogni snapshot senza perdere la selezione.

**Blocked by:** 01

**Status:** resolved

- [x] Intestazione con nome, stato e File compose del Progetto
- [x] Profili attivi visibili quando esistono
- [x] Un blocco per Servizio, con i Container nell'ordine del Progetto
- [x] Ogni Container mostra stato, salute e immagine
- [x] Le porte pubblicate compaiono con porta host, porta del Container e protocollo
- [x] Un Progetto senza Container mostra lo stato Rimosso senza errori
- [x] Un Progetto con File compose mancante lo segnala
- [x] L'aggiornamento dello snapshot non perde la selezione
- [ ] Verifica manuale: il dettaglio mostra gli stessi Servizi e Container dell'app Wails

## Comments

Implementato nel commit `c036a56`.

Corretto anche il clic sulla riga: GTK consegna al gestore un wrapper e non il
puntatore costruito, quindi la riga si risolve per indice. La lista si attiva al
primo clic, come la vista vecchia.

Verificato: test, build e schermate a 1120 e 880 px, con un Progetto Rimosso e
uno con File compose mancante.
