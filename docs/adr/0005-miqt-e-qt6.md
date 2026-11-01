# Qt 6 con miqt al posto di GTK, systray e del trasporto D-Bus

Sostituiamo `gotk4`, `libadwaita` e `fyne.io/systray` con miqt: Qt 6 Widgets in
Go, con licenza MIT. Sul desktop KDE la finestra deve sembrare parte del
sistema, e il tema Adwaita non lo fa. Oggi l'app ha anche due toolkit, perché la
tray usa una libreria a parte. Qt copre finestra, dialoghi, menu e icona di
sistema con un solo vocabolario di widget.

La build perde tre pacchetti di intestazioni e ne guadagna uno. Restano
`build-essential` e `pkg-config`. Entra `qt6-base-dev`. Escono `libgtk-4-dev`,
`libadwaita-1-dev` e `gobject-introspection`. Il costo dichiarato è il cgo con
un compilatore C++, e un primo build che compila lo shim e richiede qualche
minuto.

Il vincolo di licenza cambia natura. `gotk4-adwaita` è AGPL-3.0 e rende opera
combinata AGPL un binario distribuito. miqt è MIT e Qt 6 è LGPL-3.0: con il
linking dinamico il vincolo AGPL sparisce. Resta l'obbligo di rendere
sostituibile la libreria Qt, che l'installazione di sistema già soddisfa.

La tray passa a `QSystemTrayIcon` e le notifiche a
`QSystemTrayIcon.ShowMessage`. Una sola API per entrambe, nessun trasporto
D-Bus da mantenere. Il prezzo è che senza tray non arriva nessuna notifica
desktop, e che il click sulla notifica non distingue un'Attività da un'altra.
Su Wayland l'icona resta uno StatusNotifierItem e alcuni desktop non la
mostrano: il primo lavoro del port è un prototipo su KDE, prima di toccare il
resto.

Cambia anche il ciclo di vita della finestra. L'ADR 0003 la distruggeva alla
chiusura per liberare la memoria del webview di Wails. Con Qt Widgets non c'è
webview da liberare: la finestra si nasconde e si mostra, e la riapertura
ritrova il Progetto aperto e la cronologia delle Attività. L'ADR 0003 è quindi
sostituito. Gli stream di log e di utilizzo si fermano comunque quando la
finestra si nasconde, perché sono processi esterni. L'output di un'Attività
lanciata dalla tray continua.

Il port segue una sola passata sul branch `miqt`. Il primo ticket è un prototipo
che verifica miqt e la tray su Qt 6.10.2, su X11 e su Wayland. Il secondo porta
la shell Qt, la lista e un dettaglio minimo, e toglie i file GTK. Da lì in poi
ogni ticket riporta un'area e l'app resta compilabile. Il backend di compose,
database e impostazioni non si tocca.

I rischi restano due. Il primo è la tray su Wayland, che il prototipo deve
chiarire. Il secondo è la versione: miqt è generato contro una versione di Qt, e
Qt 6.10.2 può non combaciare con il binding. In quel caso si fissa il commit di
miqt più recente. Il guadagno atteso è una memoria di circa 30-40 MB di PSS,
contro i 60-70 MB attesi da GTK4 e i 185 MB della versione Wails.

L'ADR 0003 è sostituito per il ciclo di vita della finestra. L'ADR 0004 è
sostituito per il framework e per la libreria della tray. La specifica del port
è `05-miqt`.
