# 04: Dettaglio Progetto

**What to build:** Il pannello di destra mostra il Progetto selezionato: intestazione con nome e stato, avviso quando Docker non risponde, avviso per un File compose mancante, i dati del Progetto, i comandi e l'elenco dei Servizi con il loro stato. La riga delle Azioni è presente e disabilitata, con il motivo scritto accanto. Il contenuto si ricostruisce solo quando lo snapshot cambia, e lo scorrimento resta dov'era. Il tema usa i ruoli del foglio applicativo per le etichette di stato.

**Blocked by:** 02

**Status:** ready-for-agent

- [ ] L'intestazione mostra nome del Progetto, stato e tacche dei Container
- [ ] Docker non raggiungibile mostra un avviso con il testo di oggi, al posto dei numeri
- [ ] Durante il caricamento, prima del primo snapshot, il dettaglio non dichiara Docker non raggiungibile
- [ ] Un File compose mancante mostra l'avviso e il comando per rimuovere il Progetto dalla lista
- [ ] I dati del Progetto mostrano Gruppo, Profili attivi e File compose
- [ ] La riga delle Azioni è disabilitata quando il Progetto non ammette Azioni, con il motivo visibile
- [ ] Ogni Servizio mostra nome, stato dei suoi Container e lo spazio dell'indicatore di utilizzo
- [ ] Un Progetto senza Servizi mostra il testo di oggi
- [ ] Un aggiornamento dello snapshot non muove lo scorrimento sotto gli occhi dell'utente
- [ ] Il dettaglio si aggiorna da solo a ogni snapshot, senza perdere la selezione
- [ ] Lo stato non si distingue solo dal colore: ogni etichetta di stato ha un testo
- [ ] Test da tavolo per la scelta del testo di stato e per il motivo di un'Azione disabilitata
- [ ] Verifica manuale su KDE: il pannello regge la scala al 125% e al 150%

## Comments

Le funzioni pure già coperte dai test restano: `ProjectName`, `DockerNotice`, `ActionDisabledReason`.
