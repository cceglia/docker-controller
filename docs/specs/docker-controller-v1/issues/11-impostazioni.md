# 11: Impostazioni

**What to build:** una schermata Impostazioni permette di scegliere il comportamento alla chiusura della finestra (Nascondi nel tray come default, Esci, Chiedi ogni volta) e il comando del terminale (default `konsole --workdir {dir}`). Con "Chiedi ogni volta", la chiusura apre un dialogo con la casella "ricorda la scelta". Nel dettaglio del Progetto, "Apri terminale" usa il comando configurato nella cartella del Progetto.

**Blocked by:** 04, 10

**Status:** ready-for-agent

- [ ] Il modulo impostazioni restituisce i default su un DB nuovo e persiste le modifiche (test)
- [ ] Nascondi nel tray / Esci / Chiedi funzionano; "ricorda la scelta" salva l'impostazione
- [ ] "Apri terminale" sostituisce `{dir}` con la cartella del Progetto
