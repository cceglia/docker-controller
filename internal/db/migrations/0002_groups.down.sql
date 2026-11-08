-- Torna al Gruppo come stringa sul Progetto, conservando i nomi e le
-- associazioni: i Gruppi senza Progetti spariscono, perché senza la tabella non
-- hanno dove vivere. La discesa della 0003 ha già copiato i nomi nella colonna
-- group_name; questa discesa ricostruisce projects senza group_id e vi conserva
-- group_name. Le righe figlie passano da tabelle temporanee, perché la
-- ricostruzione elimina projects e con esso le righe che la referenziano.
CREATE TEMP TABLE project_files_backup AS SELECT * FROM project_files;
CREATE TEMP TABLE active_profiles_backup AS SELECT * FROM active_profiles;

CREATE TABLE projects_without_groups (
    name         TEXT PRIMARY KEY,
    folder       TEXT NOT NULL,
    display_name TEXT,
    group_name   TEXT
);

INSERT INTO projects_without_groups (name, folder, display_name, group_name)
SELECT name, folder, display_name, group_name FROM projects;

DROP TABLE projects;
ALTER TABLE projects_without_groups RENAME TO projects;

INSERT INTO project_files (project_name, position, path)
SELECT project_name, position, path FROM project_files_backup;
INSERT INTO active_profiles (project_name, profile)
SELECT project_name, profile FROM active_profiles_backup;

DROP TABLE project_files_backup;
DROP TABLE active_profiles_backup;

DROP INDEX groups_name;
DROP TABLE groups;
