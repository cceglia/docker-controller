# Wails v3 (beta) come framework desktop

Usiamo Wails v3, ancora in beta, con la versione fissata in `go.mod` (`v3.0.0-beta.34` al momento della decisione). Il tray con menu annidati e la finestra che si nasconde nel tray sono requisiti centrali dell'app. Wails v2 è stabile ma non ha il tray, e le librerie esterne (`fyne.io/systray`) su Linux entrano in conflitto con il loop GTK di Wails. Gli aggiornamenti di Wails vanno fatti in modo esplicito, verificando tray e finestre su KDE.
