# 01: Shell GTK4 e lista Progetti

**What to build:** L'app si avvia come applicazione GTK4 con il tema Adwaita, una finestra 1120x760 e il titolo "Docker Controller". La colonna di sinistra mostra i Progetti salvati, raggruppati per Gruppo, con il Nome visualizzato quando esiste, il simbolo di stato e le tacche dei Container. La lista si aggiorna a ogni snapshot del Manager. Il binario prodotto dalla build è questo: il `main` passa alla shell nuova, mentre gli adapter Wails restano nel repository senza essere usati fino al ticket 12.

**Blocked by:** None (can start immediately)

**Status:** ready-for-agent

- [ ] La build produce un binario che apre una finestra GTK4 con il tema Adwaita
- [ ] La finestra si apre a 1120x760 con il titolo "Docker Controller"
- [ ] La lista mostra ogni Progetto salvato con il Nome visualizzato, o il nome compose quando manca
- [ ] I Progetti con Gruppo compaiono sotto il loro Gruppo, quelli senza Gruppo in fondo nella sezione Senza Gruppo
- [ ] Ogni riga mostra il simbolo di stato: Attivo, Parziale, Fermo, Rimosso
- [ ] Le tacche mostrano fino a 12 Container e riassumono il resto in "+n"
- [ ] Un Progetto aggiunto o rimosso compare senza riavviare l'app
- [ ] La selezione di un Progetto apre il pannello di destra, anche vuoto
- [ ] La colonna ha il pulsante "Aggiungi progetto" e il pulsante "Impostazioni"
- [ ] Entrambi i pulsanti restano visibili con la lista vuota
- [ ] Il pannello di destra senza Progetto selezionato mostra un invito a sceglierne uno
- [ ] Sotto i 900 px la finestra mostra una pagina alla volta con il pulsante indietro
- [ ] Test da tavolo per l'ordine delle sezioni e per lo stato dei Container, tacche comprese
- [ ] Verifica manuale: la lista coincide con quella dell'app Wails sullo stesso database
