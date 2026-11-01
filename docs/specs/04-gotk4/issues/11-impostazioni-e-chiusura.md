# 11: Impostazioni e chiusura

**What to build:** Una finestra di impostazioni con il comportamento della chiusura (nascondi, chiedi, esci) e il comando del terminale. Chiudendo la finestra l'app fa quanto scelto; con "chiedi" compare il dialogo con l'opzione di ricordare. "Esci" termina stream e processi prima di uscire.

**Blocked by:** 06, 09

**Status:** ready-for-agent

- [ ] Il pulsante "Impostazioni" della colonna apre la finestra delle impostazioni
- [ ] La finestra delle impostazioni contiene le tre scelte di chiusura e il comando del terminale
- [ ] Un comando terminale non valido è rifiutato con un messaggio
- [ ] La scelta "Nascondi" distrugge la finestra e lascia l'app nella tray
- [ ] Alla riapertura dal tray la finestra si ricostruisce dallo snapshot
- [ ] Nascondere la finestra nella tray ferma Utilizzo risorse e ogni stream di log
- [ ] "Esci" chiude l'app
- [ ] "Chiedi ogni volta" mostra il dialogo con Nascondi, Esci e Ricorda
- [ ] Il dialogo dice che la cronologia delle Attività va persa
- [ ] La scelta ricordata vale anche dopo un riavvio
- [ ] "Esci" ferma Utilizzo risorse e i Log prima di terminare
- [ ] Le impostazioni salvate restano nel database esistente
