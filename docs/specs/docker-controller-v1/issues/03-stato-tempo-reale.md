# 03: Stato in tempo reale e Docker non raggiungibile

**What to build:** lo stato dei Progetti si aggiorna appena un Container cambia, anche se il cambio arriva dal terminale, grazie a `docker events`. Un aggiornamento completo ogni 30 secondi fa da rete di sicurezza. Se Docker non risponde, la finestra mostra il banner "Docker non raggiungibile", la lista è in grigio e le Azioni sono disabilitate; l'app riprova ogni 5 secondi e torna operativa da sola.

**Blocked by:** 01

**Status:** ready-for-agent

- [ ] Un evento container fa aggiornare solo il Progetto interessato e Watch notifica (test con il fake in streaming)
- [ ] Watch non notifica se lo snapshot non è cambiato (test)
- [ ] Lo snapshot indica se Docker è raggiungibile; quando torna disponibile riprende tutto da solo (test)
- [ ] Banner, lista in grigio e Azioni disabilitate nella UI
- [ ] Verifica manuale: `docker compose stop` da terminale si riflette nell'app entro un secondo circa
