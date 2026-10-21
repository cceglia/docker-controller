# Docker Controller

Applicazione desktop (Wails v3, Go + Svelte) per gestire i Progetti docker-compose della macchina: avvio, stop, log e stato dei Servizi.

## Prerequisiti di build

Su Debian/Ubuntu (KDE Plasma incluso):

```sh
sudo apt install build-essential pkg-config libgtk-3-dev libwebkit2gtk-4.1-dev
```

- `build-essential`, `pkg-config`: toolchain C/CGO richiesta da Wails.
- `libgtk-3-dev`, `libwebkit2gtk-4.1-dev`: librerie GTK3 e WebKitGTK 4.1. L'app usa solo GTK3, quindi il build richiede il tag `gtk3`.
- Go (1.26 o successiva), Node.js 24 e npm, per Go e per il frontend Svelte.
- Docker con il plugin `docker compose` per eseguire l'app (non serve per compilare i test).
- CLI `wails3`, versione fissata `v3.0.0-beta.34`:

```sh
go install -tags gtk3 github.com/wailsapp/wails/v3/cmd/wails3@v3.0.0-beta.34
```

Il binario finisce in `$(go env GOPATH)/bin`; il `Makefile` aggiunge questa cartella al `PATH`.

## Installazione

```sh
make install
```

Compila l'app in modalità produzione e installa in `~/.local`:

- `~/.local/bin/docker-controller`
- `~/.local/share/applications/docker-controller.desktop`
- `~/.local/share/icons/hicolor/{256x256,512x512}/apps/docker-controller.png`

Dopo l'installazione l'app compare nel menu di KDE. Per un altro percorso: `make install PREFIX=/usr/local` (con `DESTDIR=` per i pacchetti).

## Disinstallazione

```sh
make uninstall
```

Rimuove i file installati sopra. **Non rimuove il database** in `~/.local/share/docker-controller/`: i dati dei Progetti restano.

## Sviluppo e test

```sh
make test        # test Go (con tag gtk3)
make check       # vet + test + typecheck e build del frontend
make dev         # app con hot reload (wails3 dev)
make build       # binario di produzione in bin/
```

`make dev` e `make build` richiedono `wails3` nel `PATH`.
