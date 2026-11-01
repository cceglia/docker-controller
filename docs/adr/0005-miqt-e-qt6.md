# Qt 6 con miqt al posto di GTK e systray

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

La tray passa a `QSystemTrayIcon`. Le notifiche restano su D-Bus, con il
mittente `internal/notify` e `github.com/godbus/dbus/v5`. `ShowMessage` non può
riportare il click, perché manda `Notify` con l'elenco delle azioni vuoto: KDE
non ha modo di segnalarlo. Il mittente D-Bus passa l'azione `default` e ascolta
`ActionInvoked`, quindi il click mostra la finestra. La notifica arriva anche
senza tray, perché passa dal bus di sessione. Il prezzo è un trasporto da
mantenere, che però esiste già in `internal/notify`.
Su Wayland l'icona resta uno StatusNotifierItem. Il prototipo su KDE mostra che
funziona, e che il menu lo disegna il desktop tramite `com.canonical.dbusmenu`:
il popup Qt non serve.

Cambia anche il ciclo di vita della finestra. L'ADR 0003 la distruggeva alla
chiusura per liberare la memoria del webview di Wails. Con Qt Widgets non c'è
webview da liberare: la finestra si nasconde e si mostra, e la riapertura
ritrova il Progetto aperto e la cronologia delle Attività. L'ADR 0003 è quindi
sostituito. Gli stream di log e di utilizzo si fermano comunque quando la
finestra si nasconde, perché sono processi esterni. L'output di un'Attività
lanciata dalla tray continua.

Il port segue una sola passata sul branch `miqt`. Il primo ticket è un prototipo
che verifica miqt e la tray su Qt 6.10.2, su Wayland con KDE. Il secondo porta
la shell Qt, la lista e un dettaglio minimo, e toglie i file GTK. Da lì in poi
ogni ticket riporta un'area e l'app resta compilabile. Il backend di compose,
database e impostazioni non si tocca.

I rischi sono chiusi dal prototipo. miqt `v0.15.0` compila e gira contro
Qt 6.10.2, con Go 1.27.3 e g++ 15.2.0. La tray funziona su Wayland con KDE:
`QSystemTrayIcon_IsSystemTrayAvailable()` risponde `true` e
`org.kde.StatusNotifierWatcher` è presente. Il primo build dura 256,53 s con la
cache Go vuota, circa quattro minuti, e 7,3 s con la cache calda. Il guadagno
misurato è 27,0 MB di PSS a riposo con la finestra aperta, contro i 60-70 MB
attesi da GTK4 e i 185 MB della versione Wails.

L'ADR 0003 è sostituito per il ciclo di vita della finestra. L'ADR 0004 è
sostituito per il framework e per la libreria della tray. La specifica del port
è `05-miqt`.
