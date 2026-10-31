# 07: Impostazioni, dialoghi e messaggio d'errore

**What to build:** Le Impostazioni diventano una schermata a due colonne — etichetta e spiegazione a sinistra, controllo a destra — con le stesse due impostazioni di oggi (comportamento alla chiusura, comando del terminale), Salva primario, esito ed errore inline. I tre dialoghi (aggiunta Progetto, chiusura finestra, conferme) usano gli stessi token del resto della finestra, un solo pulsante primario e un backdrop più denso. Il messaggio d'errore ha un punto d'errore, il testo e le azioni, e conserva il timer con pausa su hover e focus già presente nel working tree. Ovunque sentence case, senza cambiare i significati.

**Blocked by:** 01

**Status:** ready-for-agent

- [ ] Le Impostazioni mostrano le stesse due impostazioni di oggi, con la stessa validazione e lo stesso salvataggio.
- [ ] Una voce "Progetti" nella colonna riporta al riepilogo da qualunque schermata, anche con la lista vuota.
- [ ] Salva, esito ed errore sono leggibili e vicini al comando che li produce.
- [ ] I tre dialoghi usano i token comuni, hanno un solo primario e un backdrop leggibile sul contenuto sottostante.
- [ ] La conferma testuale del Down con volumi resta identica.
- [ ] Il messaggio d'errore mostra punto d'errore, testo e le azioni "Vedi output" e "Chiudi".
- [ ] Il messaggio non sparisce mentre è letto o puntato, come già oggi.
- [ ] Le stringhe sono in sentence case e dicono le stesse cose di prima.
- [ ] Nessuno stile locale duplica colori, raggi o spaziature già dichiarati nei token.
- [ ] `svelte-check` e la build sono verdi.
