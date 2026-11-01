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
