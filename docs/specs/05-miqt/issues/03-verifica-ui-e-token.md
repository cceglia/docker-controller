# 03: Verifica della UI e token di design

**What to build:** Il repository verifica la UI senza display. `make test-ui` costruisce la finestra con `QT_QPA_PLATFORM=offscreen`, la distrugge e controlla i token di design: la scala tipografica, il corpo minimo di 16 px, i rapporti di contrasto 4,5:1 e 3:1 e il bersaglio minimo di 24x24 px. I ticket successivi estendono il test quando aggiungono un ruolo al foglio di stile.

**Blocked by:** 02

**Status:** resolved

- [x] `make test-ui` esiste ed esegue i test della UI con `QT_QPA_PLATFORM=offscreen`
- [x] `make check` resta senza display: i test offscreen e gremlins non entrano nel target
- [x] Lo smoke test costruisce la finestra vuota, la mostra con il backend offscreen e la distrugge
- [x] Lo smoke test fallisce quando un costruttore o un cablaggio è rotto
- [x] Il test dei token controlla la scala tipografica e il corpo minimo di 16 px
- [x] Il test dei token controlla i contrasti 4,5:1 per il testo e 3:1 per componenti e bordi
- [x] Il test dei token controlla il bersaglio minimo di 24x24 px
- [x] Il test legge i valori da una sola sorgente, la stessa che il codice applica
- [x] Un ruolo nuovo senza controllo fa fallire il test
- [x] Il README spiega `make test-ui` e il suo prerequisito Qt

## Esito

I token hanno una sola sorgente: `internal/gui/designtokens.go`. Il foglio di
stile in `internal/gui/qtstyle.go` si costruisce da quei valori. Il test legge
la stessa sorgente.

**Scala tipografica**: rapporto 1.125, corpo 16 px, titolo di sezione 18 px,
titolo del dettaglio 20 px. Il test deriva i due passi dal corpo e dal rapporto.
Il corpo minimo resta 16 px.

**Spaziatura**: 4, 8, 12, 16, 24 px. I padding del foglio e i margini dei
layout usano questi passi.

**Bersaglio minimo**: 24x24 px.

**Ruoli del tema chiaro**, fondo `#ffffff`:

| Ruolo | Valore | Tipo | Contrasto | Soglia |
| --- | --- | --- | --- | --- |
| `statusActive` | `#1a7f37` | testo | 5,08:1 | 4,5:1 |
| `statusPartial` | `#8a6508` | testo | 5,32:1 | 4,5:1 |
| `tackFault` | `#e05561` | componente | 3,72:1 | 3:1 |

**Ruoli del tema scuro**, fondo `#232629`:

| Ruolo | Valore | Tipo | Contrasto | Soglia |
| --- | --- | --- | --- | --- |
| `statusActive` | `#51b564` | testo | 5,90:1 | 4,5:1 |
| `statusPartial` | `#e0a800` | testo | 7,08:1 | 4,5:1 |
| `tackFault` | `#e05561` | componente | 4,09:1 | 3:1 |

I vecchi valori non passavano la soglia sul bianco: `#2f9e44` valeva 3,45:1 e
`#b8860b` valeva 3,25:1. La soglia del test resta 4,5:1. Un colore che passa sul
bianco non passa sul fondo scuro, quindi il foglio ha due palette. All'avvio la
luminosità del colore finestra della palette Qt sceglie la palette: sotto 128 la
palette è scura. `make check` non applica il tag `ui`, quindi resta senza
display.

I ruoli del tema e il foglio di stile sono controllati da
`internal/gui/designtokens_test.go`. Un ruolo senza tipo di controllo fallisce.
Un colore esadecimale nel foglio che non appartiene alla sorgente fallisce.

**Comandi e risultato**:

```
make test-ui
QT_QPA_PLATFORM=offscreen go vet -tags ui ./internal/gui
QT_QPA_PLATFORM=offscreen go test -tags ui ./internal/gui
ok  dockercontroller/internal/gui
```

Lo smoke test controlla 4 controlli interattivi contro il bersaglio di 24 px:
la lista e i tre pulsanti. La shell alza il minimo di ogni controllo al token
con `keepTarget`. `gofmt -l .`, `go vet ./...` e `make check` passano.

**Prova di rottura**: ho tolto per prova `s.window.SetCentralWidget` da
`qtshell.go`. `make test-ui` è fallito con "the window has no central widget".
Poi ho ripristinato il codice. Ho anche messo `color: #123456` nel foglio e
tolto il tipo a `tackFault`: i due test sono falliti, poi ho ripristinato.

## Comments

Il test dei token nasce con i ruoli della lista e delle etichette di stato. Il ticket 07 aggiunge i colori dei Servizi nel log.
