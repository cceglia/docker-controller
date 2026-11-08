-- Il nome del Gruppo vive solo nella tabella groups: la colonna vecchia esce
-- dallo schema. L'apertura esegue questa versione solo quando la colonna esiste
-- ancora; un database che non ce l'ha passa la versione senza eseguirla.
ALTER TABLE projects DROP COLUMN group_name;
