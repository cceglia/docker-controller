# Docker Controller

Applicazione desktop in Go con GUI GTK4 (`gotk4` e `libadwaita`) per gestire i
Progetti `docker-compose` della macchina: avvio, stop, log, stato dei Servizi,
tray di sistema e notifiche desktop.

## Prerequisiti di build

Su Debian/Ubuntu (KDE Plasma incluso):

```sh
sudo apt install build-essential pkg-config libgtk-4-dev libadwaita-1-dev gobject-introspection
```

- `build-essential`, `pkg-config`: toolchain C e cgo, che `gotk4` richiede.
- `libgtk-4-dev`: intestazioni e librerie di GTK4.
- `libadwaita-1-dev`: intestazioni e librerie di libadwaita.
- `gobject-introspection`: fornisce la voce `pkg-config` `gobject-introspection-1.0`
  e `libgirepository-1.0.so`, che `gotk4` cerca in fase di build.
- Go 1.26 o successiva.
- Docker con il plugin `docker compose` per eseguire l'app. Non serve per
  compilare i test.

## Build

```sh
make build
```

Il binario di produzione finisce in `bin/dockercontroller`.

## Test e verifica

```sh
make test        # go test ./...
make vet         # go vet ./...
make check       # vet + test + build
make test-ui     # test della UI senza display, con il backend offscreen
```

`make test-ui` esegue con `QT_QPA_PLATFORM=offscreen` il package
`internal/gui` con il tag di build `ui`: lo smoke test costruisce e distrugge
la finestra, e il test dei token controlla la scala tipografica, i contrasti e
il bersaglio minimo. Il target non chiede un display, ma chiede Qt 6
(`qt6-base-dev`), `pkg-config` e un compilatore C++, perché la build usa cgo.
I test senza tag restano senza display e girano anche in `make check`.

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
