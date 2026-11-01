# 02: Dettaglio del Progetto

**What to build:** La colonna di destra mostra il Progetto selezionato: nome, stato, File compose e Profili attivi. Sotto compaiono i Servizi, ognuno con i suoi Container: nome, immagine, stato, salute e porte pubblicate. Tutto si aggiorna a ogni snapshot senza perdere la selezione.

**Blocked by:** 01

**Status:** ready-for-agent

- [ ] Intestazione con nome, stato e File compose del Progetto
- [ ] Profili attivi visibili quando esistono
- [ ] Un blocco per Servizio, con i Container nell'ordine del Progetto
- [ ] Ogni Container mostra stato, salute e immagine
- [ ] Le porte pubblicate compaiono con porta host, porta del Container e protocollo
- [ ] Un Progetto senza Container mostra lo stato Rimosso senza errori
- [ ] Un Progetto con File compose mancante lo segnala
- [ ] L'aggiornamento dello snapshot non perde la selezione
- [ ] Verifica manuale: il dettaglio mostra gli stessi Servizi e Container dell'app Wails
