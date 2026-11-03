# Docker Controller

Applicazione desktop in Go con GUI Qt 6 tramite `miqt` (binding Qt per Go) per
gestire i Progetti `docker-compose` della macchina: avvio, stop, log, stato dei
Servizi, tray di sistema e notifiche desktop.

## Prerequisiti di build

Su Debian/Ubuntu (KDE Plasma incluso):

```sh
sudo apt install build-essential pkg-config qt6-base-dev
```

- `build-essential`: compilatore C++ e toolchain cgo, che `miqt` richiede.
- `pkg-config`: cerca le librerie di Qt 6 in fase di build.
- `qt6-base-dev`: intestazioni e librerie di Qt 6. Qt resta dinamico e di
  sistema: il binario non lo incorpora.
- Go 1.26 o successiva.
- Docker con il plugin `docker compose` per eseguire l'app. Non serve per
  compilare i test.

## Build

```sh
make build
```

Il binario di produzione finisce in `bin/dockercontroller`.

Il primo build compila lo shim C++ di `miqt` e richiede circa quattro minuti
con la cache Go vuota. Con la cache calda richiede pochi secondi.

## Test e verifica

```sh
make test        # go test ./...
make vet         # go vet ./...
make check       # vet + test + build
make test-ui     # test della UI senza display, con il backend offscreen
```

`go test ./...` compila cgo come il build, quindi chiede `qt6-base-dev`,
`pkg-config` e un compilatore C++. Una macchina senza Qt non esegue nemmeno i
test.

`make test-ui` esegue con `QT_QPA_PLATFORM=offscreen` i package
`./internal/gui/...` con il tag di build `ui`: lo smoke test costruisce e distrugge
la finestra, e il test dei token controlla la scala tipografica, i contrasti e
il bersaglio minimo. Il target non chiede un display. I test senza tag restano
senza display e girano anche in `make check`.

## Mutation testing

```sh
make tools       # installa gremlins v0.6.0 in $(go env GOPATH)/bin
make mutate      # mutation testing su tutti i package di produzione, con le soglie
make mutate PKG=./internal/compose/parse  # un solo package
```

`make tools` installa gli strumenti di sviluppo fuori dal repository. Basta
lanciarlo una volta per macchina.

`make mutate` esegue gremlins una volta per package, perché gremlins accetta un
solo percorso per lancio. Ogni lancio legge il suo file di configurazione sotto
`.gremlins/`. Un giro completo sui quindici package dura circa quattro minuti su
questa macchina. Il package `gui` pesa più degli altri, perché i suoi test
compilano Qt con cgo. Lancia `make mutate` dopo una modifica ai test, non a ogni
salvataggio. `PKG=` limita il giro a un package.

Il target resta fuori da `make check`. `make check` dura secondi, `make mutate`
dura minuti, e il repository non ha integrazione continua.

Gremlins stampa per ogni package l'efficacia dei test e la copertura dei
mutanti. L'efficacia vale `KILLED / (KILLED + LIVED)`. La copertura vale
`(KILLED + LIVED) / totale`. I mutanti `TIMED OUT` restano fuori da entrambe le
percentuali.

### Le soglie

Ogni package ha la sua coppia di soglie. Le misure dei package di compose e di
gui sono del ticket 16, dopo lo split. La misura di `internal/tray` è del ticket
15: lo split non lo tocca, quindi il ticket 16 non lo rimisura. Il valore è la
misura meno cinque punti.

| Package                   | File                             | Misura (efficacia) | Soglia efficacia | Misura (copertura) | Soglia copertura |
| ------------------------- | -------------------------------- | -----------------: | ---------------: | -----------------: | ---------------: |
| `internal/compose`        | `.gremlins/compose.yaml`         |             79,82% |           74,82% |             90,83% |           85,83% |
| `internal/compose/model`  | `.gremlins/compose-model.yaml`   |            100,00% |           95,00% |             25,00% |           20,00% |
| `internal/compose/docker` | `.gremlins/compose-docker.yaml`  |             66,67% |           61,67% |             56,25% |           51,25% |
| `internal/compose/store`  | `.gremlins/compose-store.yaml`   |             96,88% |           91,88% |            100,00% |           95,00% |
| `internal/compose/parse`  | `.gremlins/compose-parse.yaml`   |             83,93% |           78,93% |             93,33% |           88,33% |
| `internal/gui`            | `.gremlins/gui.yaml`             |             90,30% |           85,30% |             41,36% |           36,36% |
| `internal/gui/design`     | `.gremlins/gui-design.yaml`      |             85,71% |           80,71% |             87,50% |           82,50% |
| `internal/gui/view`       | `.gremlins/gui-view.yaml`        |             87,95% |           82,95% |             94,32% |           89,32% |
| `internal/gui/project`    | `.gremlins/gui-project.yaml`     |             94,29% |           89,29% |             77,78% |           72,78% |
| `internal/gui/forms`      | `.gremlins/gui-forms.yaml`       |             83,67% |           78,67% |             77,78% |           72,78% |
| `internal/gui/activity`   | `.gremlins/gui-activity.yaml`    |             95,83% |           90,83% |            100,00% |           95,00% |
| `internal/gui/logs`       | `.gremlins/gui-logs.yaml`        |             93,18% |           88,18% |             97,78% |           92,78% |
| `internal/gui/usage`      | `.gremlins/gui-usage.yaml`       |             86,21% |           81,21% |             93,55% |           88,55% |
| `internal/gui/traymenu`   | `.gremlins/gui-traymenu.yaml`    |            100,00% |           95,00% |            100,00% |           95,00% |
| `internal/tray`           | `.gremlins/tray.yaml`            |             89,66% |           84,66% |             93,55% |           88,55% |

Le quindici coppie sono diverse, quindi un solo `.gremlins.yaml` non basta.
Gremlins legge un solo file per lancio, e `make mutate` passa il file del
package con `--config`. Il nome del file deriva dal percorso: `internal/` tolto
e `/` sostituito da `-` (`internal/compose/parse` -> `compose-parse.yaml`). I
flag `--threshold-efficacy` e `--threshold-mcover` non applicano la soglia in
gremlins v0.6.0: la prova con un file vuoto e il flag a 99,9 esce con codice 0.
Per questo le soglie vivono nei file.

I file ripetono il blocco condiviso: `workers`, `timeout-coefficient` e
`exclude-files`. Se cambi uno di questi valori, cambialo in tutti i quindici
file. Un test lo verifica. `exclude-files` tiene fuori `internal/db/`,
`internal/compose/docker/fake/` e `internal/gui/guitest/`: sono supporto ai test.

`make mutate` esegue gremlins con `GOFLAGS=-count=1`. Il valore disabilita la
cache dei test. Gremlins calcola il timeout di ogni mutante dalla durata del giro
di copertura, e un giro servito dalla cache dura pochi millisecondi. Con quel
tempo il timeout scende sotto la compilazione dei test `gui`, e centinaia di
mutanti scadono invece di essere uccisi.

Sotto soglia il target esce con errore e nomina la soglia mancata e il valore
misurato:

```
SOGLIA MANCATA: ./internal/compose, efficacia 71,20%, soglia 74,82%
```

Il report sopra il messaggio elenca i mutanti sopravvissuti, con posizione
`file:riga:colonna`. La ricetta esce con codice 10 per l'efficacia e 11 per la
copertura dei mutanti. Make riporta il codice e a sua volta esce con 2.

### Aggiornare una soglia

La soglia nasce da una misura, non da un giro fallito.

1. Cambia i test o il codice del package.
2. Lancia `make mutate` e leggi `Test efficacy` e `Mutator coverage` del package.
3. Scrivi nel ticket la data e i due valori nuovi.
4. Porta la soglia al valore misurato meno cinque punti, nel file
   `.gremlins/<path>.yaml`. Il nome deriva dal percorso del package, senza
   `internal/` e con `/` sostituito da `-` (`gui-view.yaml`).

La soglia non si sposta per far passare il giro. Un giro rosso dice che i test
hanno perso forza: il rimedio sono asserzioni nuove, non un numero più basso.
Una soglia alzata a mano, sopra il misurato, blocca ogni cambiamento. Per questo
ogni movimento vuole una misura nuova, scritta nel ticket.

## Installazione

```sh
make install
```

Compila l'app e installa in `~/.local`:

- `~/.local/bin/docker-controller`
- `~/.local/share/applications/io.github.cceglia.DockerController.desktop`
- `~/.local/share/icons/hicolor/{256x256,512x512}/apps/io.github.cceglia.DockerController.png`

Dopo l'installazione l'app compare nel menu di KDE con nome e icona corretti.
Per un altro percorso: `make install PREFIX=/usr/local`. Per i pacchetti:
`make install DESTDIR=/tmp/pkg`.

## Disinstallazione

```sh
make uninstall
```

Rimuove i file installati sopra. Rimuove anche la voce `docker-controller.desktop`
e le icone `docker-controller.png` della versione precedente, così non restano
due voci nel menu.

**Non rimuove il database** in `~/.local/share/docker-controller/`: i Progetti e
le impostazioni restano.
