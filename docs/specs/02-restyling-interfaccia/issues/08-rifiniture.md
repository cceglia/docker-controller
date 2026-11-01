# 08: Rifiniture dopo la revisione e la prova sull'app

**What to build:** Chiudere i rilievi non bloccanti della revisione finale e i difetti visti nella prova sull'app vera. Un solo posto per le etichette e i toni di stato, un solo posto per le regole del punto dei Container, un tipo solo per l'utilizzo per Servizio, un componente per le coppie di rate, ProjectDetail diviso in componenti più piccoli, il cassetto che si riapre senza riavviare i log, le cifre che cambiano in Geist Mono, la striscia dei Container leggibile, percentuale di CPU come cifra principale ovunque, Down di Servizio discreto, celle di rete e disco allineate, icona dell'app con il marchio a tre blocchi.

**Blocked by:** None (can start immediately)

**Status:** resolved

- [x] Etichette e toni di stato hanno una sola definizione, usata da colonna e dettaglio.
- [x] Le regole del punto dei Container e della health stanno in un solo posto.
- [x] L'utilizzo per Servizio usa il tipo generato dai binding, non un tipo locale che lo ripete.
- [x] La coppia di rate (in/out, lettura/scrittura) è un componente condiviso dalla striscia e dai Servizi.
- [x] `ProjectDetail` è diviso: menu del Progetto, blocco del Servizio e riga del Container sono componenti separati.
- [x] "Mostra log" riapre il cassetto sul Log già aperto senza riavviare lo stream.
- [x] Ogni cifra che si aggiorna usa Geist Mono.
- [x] Lo `scrollIntoView` del messaggio d'errore rispetta `prefers-reduced-motion`.
- [x] La striscia dei Container si legge come tacche, non come una cifra, anche con una o due tacche.
- [x] La CPU della striscia globale mostra i core come cifra principale e la percentuale della macchina come nota; la barra misura la stessa percentuale.
- [x] Il Down di Servizio è discreto a riposo e diventa d'errore su hover e focus.
- [x] Le celle Rete e Disco della striscia globale allineano etichetta e valore su una riga ciascuna.
- [x] L'icona dell'app e del tray usa il marchio a tre blocchi, non la vecchia immagine.
- [x] `svelte-check` e la build sono verdi; `make check` verde.

## Comments

Implementato nel commit `bdc69ea`.

Verifica sull'app vera con Docker reale, su Xvfb, a 1120x760: guscio, dettaglio
e cassetto con log, tema scuro e chiaro. `make check` verde.

La divisione di `ProjectDetail` porta il file a 647 righe: il taglio ulteriore
dell'intestazione resta possibile e non e richiesto.
