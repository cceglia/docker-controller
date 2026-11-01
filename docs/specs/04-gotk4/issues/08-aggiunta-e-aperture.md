# 08: Aggiunta, rimozione e aperture

**What to build:** Sempre dal dettaglio: aggiungere un Progetto scegliendo più File compose dal selettore di file, con il nome proposto da compose, rimuoverlo dalla lista con conferma, e aprire la cartella, il terminale e la porta nel browser.

**Blocked by:** 02

**Status:** resolved

- [ ] Il pulsante "Aggiungi progetto" della colonna apre il selettore dei File compose
- [x] Il selettore di file permette più di un File compose
- [x] Il nome proposto da compose compare nel modulo di aggiunta e si può cambiare
- [x] Un nome già salvato è rifiutato con un messaggio
- [x] Rimuovere dalla lista chiede conferma e non tocca Docker
- [x] Un Progetto con Container non si rimuove dalla lista
- [ ] Apre la cartella del Progetto nel file manager
- [ ] Apre il terminale sul Progetto con il comando configurato
- [ ] Apre `http://localhost:<porta>` per una porta pubblicata
- [x] Una porta o un comando terminale non valido mostra un errore

## Comments

Implementato nel commit `e52b62f`.

Il pulsante della colonna apre subito il selettore multiplo, come chiede il
ticket. Il selettore passa dal portale del desktop e non risponde su `Xvfb`:
restano aperti il punto 1 e l'apertura vera di cartella, terminale e browser. I
comandi sono costruiti e testati, non avviati.
