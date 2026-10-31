# Wails v3.0.0 stabile

Usiamo Wails v3.0.0 definitiva (non più la beta), con CLI e runtime JavaScript del frontend alla stessa versione. L'API usata dall'app non è cambiata fra beta.34 e v3.0.0, quindi l'aggiornamento non ha richiesto modifiche al codice; il build resta su GTK3 finché GTK4 e WebKitGTK 6.0 non sono installati. Restano valide le altre considerazioni dell'ADR 0001.

Sul system tray di Linux due comportamenti di Wails v3.0.0 vanno aggirati, e il codice che li aggira non è ridondante: `SetTooltip` non ha effetto su Linux, dove il testo dell'hover è l'etichetta dell'icona, e un'etichetta non impostata vale "Wails" — perciò l'app imposta l'etichetta con il proprio nome. Il protocollo delle icone di stato non ha un evento di doppio click: su Linux la finestra si riporta in primo piano con un click singolo, mentre il doppio click resta registrato per Windows.
