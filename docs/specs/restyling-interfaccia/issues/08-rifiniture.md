# 08: Rifiniture dopo la revisione e la prova sull'app

**What to build:** Chiudere i rilievi non bloccanti della revisione finale e i difetti visti nella prova sull'app vera. Un solo posto per le etichette e i toni di stato, un solo posto per le regole del punto dei Container, un tipo solo per l'utilizzo per Servizio, un componente per le coppie di rate, ProjectDetail diviso in componenti più piccoli, il cassetto che si riapre senza riavviare i log, le cifre che cambiano in Geist Mono, la striscia dei Container leggibile, percentuale di CPU come cifra principale ovunque, Down di Servizio discreto, celle di rete e disco allineate, icona dell'app con il marchio a tre blocchi.

**Blocked by:** None (can start immediately)

**Status:** ready-for-agent

- [ ] Etichette e toni di stato hanno una sola definizione, usata da colonna e dettaglio.
- [ ] Le regole del punto dei Container e della health stanno in un solo posto.
- [ ] L'utilizzo per Servizio usa il tipo generato dai binding, non un tipo locale che lo ripete.
- [ ] La coppia di rate (in/out, lettura/scrittura) è un componente condiviso dalla striscia e dai Servizi.
- [ ] `ProjectDetail` è diviso: menu del Progetto, blocco del Servizio e riga del Container sono componenti separati.
- [ ] "Mostra log" riapre il cassetto sul Log già aperto senza riavviare lo stream.
- [ ] Ogni cifra che si aggiorna usa Geist Mono.
- [ ] Lo `scrollIntoView` del messaggio d'errore rispetta `prefers-reduced-motion`.
- [ ] La striscia dei Container si legge come tacche, non come una cifra, anche con una o due tacche.
- [ ] La CPU della striscia globale mostra i core come cifra principale e la percentuale della macchina come nota; la barra misura la stessa percentuale.
- [ ] Il Down di Servizio è discreto a riposo e diventa d'errore su hover e focus.
- [ ] Le celle Rete e Disco della striscia globale allineano etichetta e valore su una riga ciascuna.
- [ ] L'icona dell'app e del tray usa il marchio a tre blocchi, non la vecchia immagine.
- [ ] `svelte-check` e la build sono verdi; `make check` verde.
