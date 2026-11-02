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

`make test-ui` esegue con `QT_QPA_PLATFORM=offscreen` il package
`internal/gui` con il tag di build `ui`: lo smoke test costruisce e distrugge
la finestra, e il test dei token controlla la scala tipografica, i contrasti e
il bersaglio minimo. Il target non chiede un display. I test senza tag restano
senza display e girano anche in `make check`.

## Mutation testing

```sh
make tools       # installa gremlins v0.6.0 in $(go env GOPATH)/bin
make mutate      # mutation testing su tray, gui e compose, con le soglie
```

`make tools` installa gli strumenti di sviluppo fuori dal repository. Basta
lanciarlo una volta per macchina.

`make mutate` esegue gremlins tre volte, un package per volta, perché gremlins
accetta un solo percorso per lancio. Ogni lancio legge il suo file di
configurazione sotto `.gremlins/`. Un giro completo dura circa otto minuti su
questa macchina. Il package `gui` pesa più degli altri, perché i suoi test
compilano Qt con cgo. Lancia `make mutate` dopo una modifica ai test, non a ogni
salvataggio.

Il target resta fuori da `make check`. `make check` dura secondi, `make mutate`
dura minuti, e il repository non ha integrazione continua.

Gremlins stampa per ogni package l'efficacia dei test e la copertura dei
mutanti. L'efficacia vale `KILLED / (KILLED + LIVED)`. La copertura vale
`(KILLED + LIVED) / totale`. I mutanti `TIMED OUT` restano fuori da entrambe le
percentuali.

### Le soglie

Ogni package ha la sua coppia di soglie. Il valore è la misura del ticket 15
meno cinque punti.

| Package            | File                     | Misura ticket 15 (efficacia) | Soglia efficacia | Misura ticket 15 (copertura) | Soglia copertura |
| ------------------ | ------------------------ | ---------------------------: | ---------------: | ---------------------------: | ---------------: |
| `internal/tray`    | `.gremlins/tray.yaml`    |                       89,66% |           84,66% |                       93,55% |           88,55% |
| `internal/gui`     | `.gremlins/gui.yaml`     |                       90,33% |           85,33% |                       41,90% |           36,90% |
| `internal/compose` | `.gremlins/compose.yaml` |                       76,85% |           71,85% |                       91,53% |           86,53% |

Le tre coppie sono diverse, quindi un solo `.gremlins.yaml` non basta. Gremlins
legge un solo file per lancio, e `make mutate` passa il file del package con
`--config`. I flag `--threshold-efficacy` e `--threshold-mcover` non applicano la
soglia in gremlins v0.6.0: la prova con un file vuoto e il flag a 99,9 esce con
codice 0. Per questo le soglie vivono nei file.

I file ripetono il blocco condiviso: `workers`, `timeout-coefficient` e
`exclude-files`. Se cambi uno di questi valori, cambialo in tutti e tre.

`make mutate` esegue gremlins con `GOFLAGS=-count=1`. Il valore disabilita la
cache dei test. Gremlins calcola il timeout di ogni mutante dalla durata del giro
di copertura, e un giro servito dalla cache dura pochi millisecondi. Con quel
tempo il timeout scende sotto la compilazione dei test `gui`, e centinaia di
mutanti scadono invece di essere uccisi.

Sotto soglia il target esce con errore e nomina la soglia mancata e il valore
misurato:

```
SOGLIA MANCATA: ./internal/compose, efficacia 71,20%, soglia 71,85%
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
   `.gremlins/<package>.yaml`.

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
