# 07: Impostazioni, dialoghi e messaggio d'errore

**What to build:** Le Impostazioni diventano una schermata a due colonne — etichetta e spiegazione a sinistra, controllo a destra — con le stesse due impostazioni di oggi (comportamento alla chiusura, comando del terminale), Salva primario, esito ed errore inline. I tre dialoghi (aggiunta Progetto, chiusura finestra, conferme) usano gli stessi token del resto della finestra, un solo pulsante primario e un backdrop più denso. Il messaggio d'errore ha un punto d'errore, il testo e le azioni, e conserva il timer con pausa su hover e focus già presente nel working tree. Ovunque sentence case, senza cambiare i significati.

**Blocked by:** 01

**Status:** resolved

- [x] Le Impostazioni mostrano le stesse due impostazioni di oggi, con la stessa validazione e lo stesso salvataggio.
- [x] Una voce "Progetti" nella colonna riporta al riepilogo da qualunque schermata, anche con la lista vuota.
- [x] Salva, esito ed errore sono leggibili e vicini al comando che li produce.
- [x] I tre dialoghi usano i token comuni, hanno un solo primario e un backdrop leggibile sul contenuto sottostante.
- [x] La conferma testuale del Down con volumi resta identica.
- [x] Il messaggio d'errore mostra punto d'errore, testo e le azioni "Vedi output" e "Chiudi".
- [x] Il messaggio non sparisce mentre è letto o puntato, come già oggi.
- [x] Le stringhe sono in sentence case e dicono le stesse cose di prima.
- [x] Nessuno stile locale duplica colori, raggi o spaziature già dichiarati nei token.
- [x] `svelte-check` e la build sono verdi.

## Comments

Implementato nel commit `4385652`.

Revisione finale del 2026-10-31 con punto fisso `0ab9372`: nessun rilievo
bloccante su standard o spec. I rilievi non bloccanti sono nel report del run.
