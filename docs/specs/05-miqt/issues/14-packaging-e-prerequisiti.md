# 14: Packaging, prerequisiti Qt e pulizia

**What to build:** Il progetto si prepara e si installa con Qt 6 al posto di GTK. Il README elenca `build-essential`, `pkg-config` e `qt6-base-dev`, spiega che il primo build compila lo shim C++ e che anche `go test` richiede Qt. Il README non nomina più GTK, Adwaita o gobject-introspection. Il `go.mod` non contiene più `gotk4`, `gotk4-adwaita` e `fyne.io/systray`. `make install` e `make uninstall` installano e rimuovono lo stesso binario, la stessa voce di menu e le stesse icone di oggi.

**Blocked by:** 02, 11, 12, 13

**Status:** ready-for-agent

- [ ] `go mod tidy` toglie `gotk4`, `gotk4-adwaita` e `fyne.io/systray`. `godbus` resta per `internal/notify`
- [ ] Il README nomina i prerequisiti Qt 6 e non nomina GTK
- [ ] Il README spiega che il primo build compila lo shim C++ e richiede circa quattro minuti con la cache Go vuota
- [ ] Il README spiega che `go test ./...` richiede `qt6-base-dev` e un compilatore C++
- [ ] Il Makefile non nomina GTK fra le dipendenze di build, e `make build` usa il cgo
- [ ] `gtk-update-icon-cache` o un comando equivalente resta per aggiornare la cache delle icone
- [ ] `make install` installa binario, voce desktop e icone con gli stessi nomi di oggi
- [ ] `make uninstall` rimuove anche i nomi della versione precedente e conserva il database
- [ ] La voce desktop e le icone restano associate all'ID `io.github.cceglia.DockerController`
- [ ] Il test del packaging controlla i nuovi prerequisiti, l'assenza dei vecchi e la corrispondenza fra nome applicazione e `StartupWMClass`
- [ ] `make check` verde da un albero pulito, con la cache di build vuota
- [ ] Verifica manuale su KDE: l'app installata compare nel menu con nome e icona corretti

## Comments

Ultimo ticket del port. Da qui il repository non conosce più GTK. Qt resta dinamico e di sistema: il pacchetto non lo incorpora.
