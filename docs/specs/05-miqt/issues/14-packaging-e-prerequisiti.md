# 14: Packaging, prerequisiti Qt e pulizia

**What to build:** Il progetto si prepara e si installa con Qt 6 al posto di GTK. Il README elenca `build-essential`, `pkg-config` e `qt6-base-dev`, spiega che il primo build compila lo shim C++ e che anche `go test` richiede Qt. Il README non nomina più GTK, Adwaita o gobject-introspection. Il `go.mod` non contiene più `gotk4`, `gotk4-adwaita` e `fyne.io/systray`. `make install` e `make uninstall` installano e rimuovono lo stesso binario, la stessa voce di menu e le stesse icone di oggi.

**Blocked by:** 02, 11, 12, 13

**Status:** resolved

- [x] `go mod tidy` toglie `gotk4`, `gotk4-adwaita` e `fyne.io/systray`. `godbus` resta per `internal/notify`
- [x] Il README nomina i prerequisiti Qt 6 e non nomina GTK
- [x] Il README spiega che il primo build compila lo shim C++ e richiede circa quattro minuti con la cache Go vuota
- [x] Il README spiega che `go test ./...` richiede `qt6-base-dev` e un compilatore C++
- [x] Il Makefile non nomina GTK fra le dipendenze di build, e `make build` usa il cgo
- [x] `gtk-update-icon-cache` o un comando equivalente resta per aggiornare la cache delle icone
- [x] `make install` installa binario, voce desktop e icone con gli stessi nomi di oggi
- [x] `make uninstall` rimuove anche i nomi della versione precedente e conserva il database
- [x] La voce desktop e le icone restano associate all'ID `io.github.cceglia.DockerController`
- [x] Il test del packaging controlla i nuovi prerequisiti, l'assenza dei vecchi e la corrispondenza fra nome applicazione e `StartupWMClass`
- [x] `make check` verde da un albero pulito, con la cache di build vuota
- [ ] Verifica manuale su KDE: l'app installata compare nel menu con nome e icona corretti

## Comments

Ultimo ticket del port. Da qui il repository non conosce più GTK. Qt resta dinamico e di sistema: il pacchetto non lo incorpora.

## Esito

Dipendenze:

- `go mod tidy` toglie `fyne.io/systray`, `gotk4` e `gotk4-adwaita`, più
  `golang.org/x/sync`, che solo `systray` portava dentro.
  `github.com/godbus/dbus/v5` resta per `internal/notify`. Nessun file Go
  importa i package tolti. Una seconda esecuzione non cambia niente.
- `git diff go.mod go.sum`: 10 righe rimosse, nessuna aggiunta.

Documentazione e script:

- README: prerequisiti `build-essential`, `pkg-config`, `qt6-base-dev`. Spiega
  che il primo build compila lo shim C++ e costa circa quattro minuti a cache
  vuota, e che `go test ./...` chiede Qt e un compilatore C++. Nessun
  riferimento a GTK, Adwaita, gobject-introspection o gotk4.
- Makefile: intestazione su Qt 6 via miqt con cgo. `make build` resta
  `CGO_ENABLED=1 go build`. `gtk-update-icon-cache` vive nella variabile
  `ICON_CACHE`, con un commento che spiega perché non è una dipendenza di GTK.
  `make test`, `make test-ui` e `make check` non cambiano.

Verifica:

- `GOCACHE=/tmp/opencode/gocache-cold14 make check`, cache vuota: **288,92 s**,
  verde. `go vet ./...`, `go test ./...` su tutti i package, build cgo.
- `make test-ui`: verde.
- `make install PREFIX=/tmp/opencode/prefix DESTDIR=`: installa
  `bin/docker-controller`,
  `share/applications/io.github.cceglia.DockerController.desktop` con
  `Exec=/tmp/opencode/prefix/bin/docker-controller` e
  `StartupWMClass=io.github.cceglia.DockerController`, e le icone `256x256` e
  `512x512` con nome `io.github.cceglia.DockerController.png`.
- `make uninstall PREFIX=/tmp/opencode/prefix DESTDIR=`: rimuove i nomi nuovi e
  quelli della versione precedente (`docker-controller.desktop`,
  `docker-controller.png` nelle due misure). Il database finto nel prefix
  resta.
- Binario installato su xcb: la finestra `0x100000e` ha classe
  `io.github.cceglia.DockerController` e titolo "Docker Controller". Cattura in
  `/tmp/opencode/shot14.png` con `import -window`.

Test del packaging, in `packaging_test.go`, che leggono i file veri:

- `TestReadmeNamesTheQtPrerequisites`
- `TestMakefileBuildsWithQtAndCgo`
- `TestApplicationNameMatchesStartupWMClass`: confronta `appID` di
  `internal/gui/windowconsts.go` con `StartupWMClass` della voce desktop.

Resta manuale l'ultima voce: la comparsa nel menu di KDE dopo un `make install`
reale in `~/.local`. Questa sessione installa solo sotto `/tmp/opencode/prefix`.

