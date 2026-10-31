# 06: Cassetto Attività e Log

**What to build:** Attività e Log escono dal flusso della pagina e diventano un cassetto di 220px in fondo alla tela con due schede a sottolineatura. Si apre da solo quando parte un'Azione, sulla scheda Attività; "Mostra log" lo apre sulla scheda Log; la scheda Log esiste solo con uno stream aperto. Dentro restano filtro Servizio, ricerca, timestamp, pausa e chiusura, con l'area del log che occupa tutta la larghezza del cassetto. Chiudere il cassetto non ferma i log; cambiare Progetto o andare in Impostazioni sì. Il "Vedi output" del messaggio d'errore porta alla scheda Attività giusta.

**Blocked by:** 04

**Status:** ready-for-agent

- [ ] Attività e Log stanno in un cassetto in fondo alla tela, con due schede a sottolineatura.
- [ ] Il cassetto si apre da solo quando parte un'Azione, sulla scheda Attività.
- [ ] "Mostra log" apre il cassetto sulla scheda Log del Progetto giusto.
- [ ] La scheda Log compare solo con uno stream aperto; senza stream il cassetto mostra la sola scheda Attività.
- [ ] Filtro Servizio, ricerca, timestamp, pausa e chiusura funzionano come oggi.
- [ ] L'area del log usa la larghezza del cassetto e si adatta alla sua altezza, senza i 20rem fissi di oggi.
- [ ] Chiudere il cassetto non ferma lo stream; cambiare Progetto o aprire le Impostazioni sì.
- [ ] "Vedi output" del messaggio d'errore apre il cassetto sull'Attività giusta.
- [ ] La cronologia delle Attività resta selezionabile e l'output segue l'ultima riga.
- [ ] `svelte-check` e la build sono verdi.
