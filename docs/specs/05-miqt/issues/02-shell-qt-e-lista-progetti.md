# 02: Shell Qt, lista Progetti e uscita di GTK

**What to build:** L'app si avvia come applicazione Qt 6 Widgets, con la finestra 1120x760 e il titolo "Docker Controller". La colonna di sinistra mostra i Progetti salvati, raggruppati per Gruppo, con il Nome visualizzato quando esiste, il simbolo di stato e le tacche dei Container. La lista si aggiorna a ogni snapshot del Manager. Il dettaglio a destra per ora invita a scegliere un Progetto. I file di vista GTK escono dal repository qui, insieme alla loro logica pura: la build non conosce più GTK. Tray e notifiche mancano fino ai ticket 11 e 12, e la chiusura esce dall'app fino al ticket 13.

**Blocked by:** 01

**Status:** ready-for-agent

- [ ] `make check` verde: `go vet`, `go test` e build con cgo contro Qt 6
- [ ] La finestra si apre a 1120x760 con il titolo "Docker Controller"
- [ ] La finestra usa `QMainWindow` con `QSplitter` fra lista e dettaglio
- [ ] La lista mostra ogni Progetto salvato con il Nome visualizzato, o il nome compose quando manca
- [ ] I Progetti con Gruppo compaiono sotto il loro Gruppo, quelli senza Gruppo in fondo nella sezione Senza Gruppo
- [ ] Ogni riga mostra il simbolo di stato: Attivo, Parziale, Fermo, Rimosso
- [ ] Le tacche mostrano fino a 12 Container e riassumono il resto in "+n"
- [ ] Un Progetto aggiunto o rimosso compare senza riavviare l'app
- [ ] La selezione di un Progetto apre il pannello di destra, anche vuoto
- [ ] La colonna ha i pulsanti "Aggiungi progetto" e "Impostazioni", visibili anche con la lista vuota
- [ ] Sotto i 900 px la finestra mostra una pagina alla volta con il pulsante indietro
- [ ] Le funzioni pure che vivono nei file GTK si spostano in file senza Qt prima della cancellazione: tacche, pannello Attività, comportamento di chiusura, preferenze e utilizzo
- [ ] I test di quelle funzioni restano verdi, e il ticket elenca quali test restano, quali si spostano e quali spariscono
- [ ] I test di compose, database, impostazioni e tray restano identici
- [ ] Nessun import di `gotk4`, `gotk4-adwaita` o `fyne.io/systray` resta nel repository
- [ ] L'applicazione dichiara l'ID `io.github.cceglia.DockerController` come nome applicazione, e la classe di finestra coincide con `StartupWMClass`
- [ ] Liste e campi hanno un nome accessibile e un ordine di tabulazione esplicito
- [ ] Verifica manuale su KDE: la finestra segue lo stile del desktop e l'icona nel menu è quella giusta

## Comments

Primo ticket di codice del branch `miqt`. L'app perde tray, notifiche e comportamento di chiusura fino ai ticket successivi: è accettato, perché GTK e Qt non convivono in un solo processo. La verifica senza display e i token di design arrivano nel ticket 03.
