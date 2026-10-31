# 01: Token, font e guscio

**What to build:** La finestra si apre a 1120×760 con il guscio a due pannelli e la direzione visiva del foglio: colonna a sinistra con il marchio (tre blocchi, il superiore acceso), tela a destra con la striscia dell'Utilizzo risorse globale in cima e un'area contenuto vuota. Colori, raggi, altezze e scala tipografica vengono da un unico insieme di token; il tema scuro è il riferimento e il chiaro ne deriva; Geist e Geist Mono sono serviti dall'app, senza chiedere nulla alla rete. La striscia globale è dichiaratamente Docker e mostra Container in esecuzione, CPU e memoria con un indicatore sottile (il solo hue della schermata), rete e disco come valori al secondo; distingue "in attesa della prima misura" da "non disponibile".

**Blocked by:** None (can start immediately)

**Status:** ready-for-agent

- [ ] La finestra si apre a 1120×760, ridimensionabile, con il guscio a due pannelli visibile.
- [ ] Colori, raggi, altezze e dimensioni tipografiche sono dichiarati in un solo posto; nessun componente dichiara valori propri.
- [ ] Il tema segue `prefers-color-scheme` e i due temi usano lo stesso vocabolario di ruoli.
- [ ] I due caratteri sono serviti dall'app: la finestra non fa richieste di rete per la tipografia e non cade su un font di sistema.
- [ ] Ogni cifra che si aggiorna usa cifre tabulari; identificatori e percorsi usano il monospaziato.
- [ ] La striscia globale mostra Container, CPU, memoria, rete e disco; con dati assenti mostra lo stato di attesa, non zeri.
- [ ] Il hue dei dati compare una sola volta nella schermata, sull'indicatore di CPU.
- [ ] Nessuna ombra a riposo, nessun gradiente, nessun testo sotto i 12px.
- [ ] Il focus da tastiera è visibile su ogni controllo e con movimento ridotto non ci sono transizioni.
- [ ] `svelte-check` e la build sono verdi.
