-- Ricrea la colonna del nome sul Progetto e vi copia i nomi dei Gruppi, così
-- la discesa della 0002 può ricostruire projects e conservarli. Solo questa
-- discesa legge i nomi dalla tabella groups.
ALTER TABLE projects ADD COLUMN group_name TEXT;

UPDATE projects
SET group_name = (SELECT g.name FROM groups g WHERE g.id = projects.group_id);
