# DockerController

App desktop per vedere e governare tutti i progetti docker-compose presenti sul computer, dalla finestra e dal system tray.

## Progetti

**Progetto**:
Un progetto docker-compose, identificato dal suo nome compose (es. `sunlinkdocker`). Ha uno o più File compose; la cartella in cui si trova non ne determina l'identità.
_Avoid_: stack, compose, applicazione

**File compose**:
Uno dei file YAML che definiscono un Progetto (es. `docker-compose.yml` più un override).
_Avoid_: config, yml

**Nome visualizzato**:
Un nome facoltativo scelto dall'utente, mostrato al posto del nome compose del Progetto.
_Avoid_: etichetta, alias

**Progetto scoperto**:
Un Progetto che Docker conosce in questo momento perché ha almeno un Container.

**Progetto salvato**:
Un Progetto presente nella lista persistente dell'app; ogni Progetto scoperto diventa automaticamente salvato e resta tale anche dopo un Down, finché l'utente non lo rimuove dalla lista.

**Rimuovi dalla lista**:
Toglie un Progetto Rimosso dalla lista dei Progetti salvati, senza toccare Docker.
_Avoid_: elimina, cancella

**File compose mancante**:
Condizione di un Progetto salvato i cui File compose non esistono più su disco.

## Gruppi

**Gruppo**:
Una categoria nominata dall'utente (es. "Lavoro") che raccoglie Progetti nella lista e nel tray. Ogni Progetto appartiene al massimo a un Gruppo; un Gruppo esiste anche senza Progetti, e due Gruppi non possono avere lo stesso nome, nemmeno con maiuscole diverse o spazi iniziali e finali.
_Avoid_: etichetta, tag, categoria

**Nessun Gruppo**:
Il segnaposto del campo Gruppo quando il Progetto non ha Gruppo. Il nome è riservato e non è un Gruppo.

**Senza Gruppo**:
La condizione dei Progetti che non appartengono a un Gruppo: quando la lista ha sezioni, formano una sezione in fondo, e nel tray stanno in fondo al primo livello. Non è un Gruppo, il suo nome è riservato e non si gestisce.
_Avoid_: gruppo vuoto, gruppo predefinito

**Rinomina di un Gruppo**:
Cambia il nome di un Gruppo per tutti i suoi Progetti. Rinominarlo sul nome di un Gruppo esistente unisce i due Gruppi: sopravvive quello che aveva già il nome di arrivo.
_Avoid_: modifica del Gruppo, spostamento

**Eliminazione di un Gruppo**:
Toglie il Gruppo e lascia i suoi Progetti senza Gruppo, senza toccare i Progetti né Docker.
_Avoid_: rimozione dalla lista, cancellazione

## Stato di un Progetto

**Attivo**:
Tutti i Container del Progetto sono in esecuzione.

**Parziale**:
Alcuni Container del Progetto sono in esecuzione, altri no.

**Fermo**:
Il Progetto ha Container, ma nessuno è in esecuzione.

**Rimosso**:
Il Progetto non ha alcun Container (dopo un Down, o mai avviato).
_Avoid_: down, cancellato

## Servizi

**Servizio**:
Una voce `services:` di un Progetto; è l'unità su cui si eseguono le azioni.

**Container**:
Un'istanza in esecuzione o ferma di un Servizio; un Servizio con repliche ha più Container.

**Utilizzo risorse**:
CPU, memoria, rete e I/O disco consumati in questo momento da un insieme di Container: tutti quelli di Docker (globale) oppure quelli di un Servizio.
_Avoid_: stats, consumi, metriche

**Profilo**:
Un insieme nominato di Servizi che partono solo quando il Profilo è attivo.

**Profili attivi**:
I Profili che l'utente ha scelto di usare per un Progetto salvato; valgono per tutte le azioni tranne Down e Down con volumi, che agiscono su tutti i Profili.

## Azioni

**Azione**:
Un'operazione lanciata su un Progetto o su un suo Servizio: Up, Stop, Restart, Start, Pull, Build, Down, Down con volumi.
_Avoid_: comando, operazione

**Attività**:
Una singola esecuzione di un'Azione, con il suo output e il suo esito; su ogni Progetto può esserci al massimo un'Attività in corso.
_Avoid_: job, task

**Down**:
Rimuove Container e reti di un Progetto, conservando i volumi.

**Down di un Servizio**:
Ferma e rimuove i Container di un solo Servizio, conservando i volumi. Richiede conferma.

**Down con volumi**:
Down che elimina anche i volumi del Progetto, e quindi i suoi dati.

**Azione distruttiva**:
Un'azione che rimuove qualcosa e richiede conferma esplicita: Down, Down di un Servizio, Down con volumi, Rimuovi dalla lista.
