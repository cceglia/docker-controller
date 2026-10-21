-- Progetti salvati
CREATE TABLE projects (
    name         TEXT PRIMARY KEY,
    folder       TEXT NOT NULL,
    display_name TEXT,
    group_name   TEXT
);

-- File compose di un Progetto, in ordine (-f)
CREATE TABLE project_files (
    project_name TEXT    NOT NULL REFERENCES projects(name) ON DELETE CASCADE,
    position     INTEGER NOT NULL,
    path         TEXT    NOT NULL,
    PRIMARY KEY (project_name, position)
);

-- Profili attivi per Progetto
CREATE TABLE active_profiles (
    project_name TEXT NOT NULL REFERENCES projects(name) ON DELETE CASCADE,
    profile      TEXT NOT NULL,
    PRIMARY KEY (project_name, profile)
);

-- Impostazioni dell'app (chiave/valore)
CREATE TABLE settings (
    key   TEXT PRIMARY KEY,
    value TEXT NOT NULL
);
