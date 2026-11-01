# 11: Impostazioni e chiusura

**What to build:** Una finestra di impostazioni con il comportamento della chiusura (nascondi, chiedi, esci) e il comando del terminale. Chiudendo la finestra l'app fa quanto scelto; con "chiedi" compare il dialogo con l'opzione di ricordare. "Esci" termina stream e processi prima di uscire.

**Blocked by:** 06, 09

**Status:** resolved

- [x] Il pulsante "Impostazioni" della colonna apre la finestra delle impostazioni
- [x] La finestra delle impostazioni contiene le tre scelte di chiusura e il comando del terminale
- [x] Un comando terminale non valido è rifiutato con un messaggio
- [x] La scelta "Nascondi" distrugge la finestra e lascia l'app nella tray
- [x] Alla riapertura dal tray la finestra si ricostruisce dallo snapshot
- [x] Nascondere la finestra nella tray ferma Utilizzo risorse e ogni stream di log
- [x] "Esci" chiude l'app
- [x] "Chiedi ogni volta" mostra il dialogo con Nascondi, Esci e Ricorda
- [x] Il dialogo dice che la cronologia delle Attività va persa
- [x] La scelta ricordata vale anche dopo un riavvio
- [x] "Esci" ferma Utilizzo risorse e i Log prima di terminare
- [x] Le impostazioni salvate restano nel database esistente

## Comments

Implementato nel commit `15f129e`.

Ciclo di vita provato senza una persona, con un `docker` finto, click XTest e
`WM_DELETE_WINDOW`: chiudere lascia il processo vivo, il tray ricostruisce la
finestra dallo snapshot, `Esci` termina e non resta nessun processo `docker`.
La cronologia delle Attività riparte vuota con ogni finestra, così il testo del
dialogo resta vero.

Nota: il salvataggio non sta nell'header perché `AdwWindow` non ammette
`gtk_window_set_titlebar`; è una riga `Salva` in fondo alla pagina.
