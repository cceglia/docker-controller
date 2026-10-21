## CodeGraph — obbligatorio prima di esplorare il codice

**CodeGraph** è la fonte di riferimento per capire il codice di questo progetto: restituisce simboli, call path, dipendenze e impatto di una modifica. Ci sono due modi equivalenti per usarlo:

- il tool MCP `mcp__codegraph__codegraph_explore`, se è disponibile nella sessione (è differito: caricane lo schema con ToolSearch `select:mcp__codegraph__codegraph_explore` prima della prima chiamata);
- il comando `codegraph explore "<domanda o nomi di simboli>"` tramite Bash, che funziona sempre.

**REGOLA: usa CodeGraph prima di Read, Grep, Glob o di `grep`/`rg`/`find` da shell** per relazioni nel codice, dipendenze, chiamanti, percorsi di esecuzione, comportamento che attraversa più file e qualsiasi altra esplorazione del sorgente. Puoi saltarlo solo quando conosci già il percorso esatto del file e il numero di riga che ti servono. La regola vale anche per ogni subagent a cui deleghi.

Interrogalo con una domanda in linguaggio naturale o con i nomi dei file e dei simboli rilevanti. Per le domande su un flusso, includi i simboli che coprono il flusso, così CodeGraph può restituire il call path tra di essi. Considera il sorgente restituito come già letto e non ripetere l'esplorazione con Grep o Read.

### Interrogarlo bene

L'indice copre il backend Go e il frontend Svelte/TypeScript. L'output di `explore` è limitato (circa 12 file di default). Quindi:

- Metti nella query nomi di simboli, di file o di package (`internal/compose Manager Do Attività`, `frontend/src/lib/store.ts`), non parole generiche come "handler" o "service": una query vaga trova simboli omonimi dell'altro linguaggio e manca quello cercato.
- Quando la risposta è troncata o mostra solo frammenti (`... (gap) ...`), allargala con `codegraph explore --max-files <n> "…"` o restringi la query; non ripiegare su Grep.
- Per un intervallo di righe preciso o un intero simbolo usa la CLI invece di Read, `sed` o `cat` (il server MCP espone solo `explore`): `codegraph node -f <file relativo al repo> --offset <riga> --limit <n>` (modalità file, con numeri di riga) oppure `codegraph node <Simbolo> [-f <file>]` (il simbolo completo con chiamanti e chiamati).
- Ignora le voci "used by" tra linguaggi diversi (per esempio un file `.ts` del frontend elencato come dipendente di un file Go): sono collisioni di nomi, non dipendenze reali. I binding TypeScript generati da Wails sono esclusi dall'indice (`codegraph.json`).

CodeGraph aggiorna l'indice da solo tramite il server MCP. **Non eseguire mai `codegraph update`, ricostruzioni manuali o altri aggiornamenti dell'indice** dopo aver modificato il sorgente. Senza server MCP attivo, `codegraph sync` è l'unico comando ammesso per allineare l'indice.

### Delegare ai subagent

Ogni prompt di delega che può esplorare codice sorgente inizia con un blocco **CodeGraph first**:

- CodeGraph è il primo passo di esplorazione e anche ogni passo successivo;
- Grep, Read, `sed` o `cat` sul sorgente servono solo per un intervallo che CodeGraph non è riuscito a restituire;
- le regole di "Interrogarlo bene" qui sopra, copiate per intero, perché un subagent non applica questo file in modo affidabile da solo;
- mai `codegraph update`;
- il report finale elenca le query CodeGraph eseguite, così chi ha delegato può verificarle.

## Agent skills

### Issue tracker

Issue e spec sono file markdown locali in `docs/specs/`. See `docs/agents/issue-tracker.md`.

### Triage labels

Le cinque etichette canoniche predefinite (needs-triage, needs-info, ready-for-agent, ready-for-human, wontfix). See `docs/agents/triage-labels.md`.

### Domain docs

Single-context: `GLOSSARY.md` + `docs/adr/` nella root. See `docs/agents/domain.md`.
