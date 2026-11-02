# Sottopackage di compose e gui

`internal/compose` e `internal/gui` sono cresciuti in un package piatto: 27 file
e circa 5.200 righe il primo, 75 file e circa 16.700 righe il secondo, di cui
circa 10.000 di adattatori Qt. Un package unico mescola responsabilità diverse
(parsing dell'output Docker, persistenza, porta verso la CLI, logica di
presentazione, widget) e rende il test dei pezzi puri dipendente dal build con
cgo. Dividiamo i due package in sottopackage dentro `internal/`, lasciando in
radice il punto di composizione: `Manager` per compose, la shell Qt per gui.

`internal/compose` si divide in `model` (tipi di dominio, helper puri e
sentinelle), `docker` (porta e adattatore della CLI, con `fake` escluso dal
build di produzione), `store` (persistenza SQLite) e `parse` (parsing
dell'output della CLI). `Manager` resta l'unica interfaccia del modulo e resta
in radice, con i metodi che oggi possiede: `Do`, `Logs` e `Usage` non diventano
componenti separati, perché sono metodi su `*Manager` e spostarli richiederebbe
un ridisegno del modulo senza un guadagno di profondità. I tipi usati fuori da
compose restano raggiungibili con alias (`type Project = model.Project`), così
`gui`, `tray` e `main` non cambiano.

`internal/gui` si divide in `design`, `view`, `project`, `forms`, `activity`,
`logs`, `usage` e `traymenu`, più `guitest` per gli helper dei test. Le viste Qt
restano in radice: sono adattatori, e spostarle per schermata richiederebbe di
esportare l'intera superficie dei widget usata dalla shell. I sottopackage
contengono solo codice senza Qt, già testabile senza display. `view` è il
package foglia e tiene i testi condivisi (`ProjectName`, `DisplayNameOf`,
`MissingComposeText`), `design` tiene il foglio di stile, e i messaggi degli
errori dei Log stanno in `logs`: così i tre cicli fra `view` e i package di
feature si chiudono. La regola di dipendenza è `design`/`view` ← package di
feature ← `internal/gui`; nessun sottopackage importa la radice.

Le alternative scartate: un package per schermata (più profondo, ma esporta
tutta la superficie delle viste); estrarre `activity`, `logs` e `usage` come
componenti di compose (richiede un ridisegno di `Manager`); una cartella
`packages/` fuori da `internal/` (perde la garanzia che il codice non sia
importabile da fuori). Ogni nuovo package di codice di produzione entra in
Gremlins con il proprio file `.gremlins/<path>.yaml`; `docker/fake` e `guitest`
restano fuori perché sono supporto ai test. Lo split cambia ciò che Gremlins
misura, quindi le soglie di `compose` e `gui` vanno rimisurate e `make mutate`
passa a quindici package. `store`, `parse` e `model` ricevono test di package,
perché oggi la loro copertura passa solo da `Manager` e senza test Gremlins non
misura nulla.
