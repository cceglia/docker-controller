# 10: Aggiunta e aperture

**What to build:** Dalla finestra l'utente aggiunge un Progetto scegliendo i File compose con il selettore di file di Qt, rimuove un Progetto dalla lista, apre la cartella, il terminale e la porta nel browser. La rimozione dalla lista chiede conferma. L'aggiunta rifiutata mostra il motivo con il testo di oggi.

**Blocked by:** 04, 05

**Status:** ready-for-agent

- [ ] "Aggiungi progetto" apre il selettore di file di Qt
- [ ] Il selettore accetta più File compose e li mostra nell'ordine scelto
- [ ] Un nome compose già presente mostra il rifiuto con il testo di oggi
- [ ] Un insieme di file che compose non accetta mostra il rifiuto con il testo di oggi
- [ ] Il Progetto aggiunto compare nella lista senza riavviare l'app
- [ ] "Rimuovi dalla lista" chiede conferma e non tocca Docker
- [ ] "Apri cartella" e "Apri terminale" usano `xdg-open` e il comando configurato
- [ ] "Apri porta" apre l'indirizzo nel browser
- [ ] I comandi restano disabilitati quando il File compose manca
- [ ] Test da tavolo per l'ordine dei File compose e per i testi di rifiuto
- [ ] Verifica manuale su KDE: il selettore di file mostra i file nascosti quando serve

## Comments

La validazione dell'aggiunta resta nel backend, con i test di oggi.
