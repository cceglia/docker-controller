# Docker comandato tramite la CLI `docker compose`, non tramite SDK

L'app esegue `docker compose …` e `docker events` come processi esterni invece di usare l'SDK Go di Docker o la libreria di compose. L'SDK non contiene la logica di compose (profili, `.env`, override, `up`). La libreria di compose è pesante e cambia API spesso. Con la CLI il comportamento è identico a quello del terminale. Il prezzo da pagare è il parsing dell'output JSON della CLI, isolato in un unico modulo e coperto da test con output reali.
