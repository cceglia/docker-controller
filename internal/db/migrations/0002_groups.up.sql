-- Gruppi come entità: il nome e l'ordine non dipendono più dai Progetti che
-- li usano, e un Gruppo può esistere senza Progetti. Il passo di dati in Go
-- riempie la tabella dai nomi già scritti sui Progetti, e la 0003 toglie la
-- vecchia colonna: qui lo schema cambia soltanto.
CREATE TABLE groups (
    id       INTEGER PRIMARY KEY,
    name     TEXT    NOT NULL,
    position INTEGER NOT NULL
);

-- Il nome è unico per come è scritto; l'unicità senza distinzione di maiuscole
-- la garantisce compose, che confronta anche i caratteri fuori dall'ASCII.
CREATE UNIQUE INDEX groups_name ON groups (name);

-- Ogni Progetto punta al suo Gruppo; alla cancellazione del Gruppo il Progetto
-- resta e torna senza Gruppo. La colonna group_name resta finché la 0003 non
-- ha copiato i nomi.
ALTER TABLE projects ADD COLUMN group_id INTEGER REFERENCES groups(id) ON DELETE SET NULL;
