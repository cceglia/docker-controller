# Finestra distrutta nel tray

*Sostituito dall'ADR 0005: con Qt la finestra si nasconde e si mostra, non si distrugge.*

Chiudere la finestra la fa sparire dal desktop, ma non libera la memoria del webview. Wails nasconde la finestra: il processo WebKit e la sua memoria restano allocati per tutta la vita dell'app. Su una macchina che tiene l'app nel tray per giorni, quella memoria è sprecata.

La finestra ora si distrugge davvero. Prima di distruggerla l'app ferma lo stream di Utilizzo risorse e ogni stream dei log, perché `docker stats` e `compose logs` sono processi esterni e non devono sopravvivere alla finestra. Il clic sul tray ricrea la finestra con le stesse opzioni e riavvia lo stream di utilizzo.

Wails v3.0.0 non espone un comando pubblico di distruzione: `WebviewWindow.Close()` emette solo l'evento di chiusura, e la distruzione avviene quando nessun hook lo annulla. L'app lascia quindi passare l'evento invece di annullarlo e nascondere. `Linux.DisableQuitOnLastWindowClosed` tiene in vita il processo quando l'ultima finestra sparisce.

Il costo è dichiarato e accettato: la riapertura dal tray richiede circa un secondo, il frontend riparte da zero e perde la cronologia delle Attività della sessione, il Progetto aperto e i filtri dei log. I Progetti, i Gruppi e le impostazioni restano nel database, quindi non si perde nulla di persistente. Il testo dell'impostazione nelle Impostazioni e nel dialogo di chiusura dice questo.
