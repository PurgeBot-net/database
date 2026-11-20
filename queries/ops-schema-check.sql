-- Which PurgeBot tables does this database have, and with which columns?
-- Run it first on a database you have not checked. The other files expect customizations, purge_events,
-- guild_settings, access_roles and guild_departures.
-- You get one row per expected table, plus any other table with a guild_id column.
-- present false means the table is missing, and the files that read it will fail.
-- known false means a table has guild_id but the GDPR tools in cmd do not know it, so they refuse to run.
-- schema is where the session looks up table names. It should be the PurgeBot schema.
-- Nothing to edit.

WITH expected(table_name) AS (
    VALUES ('customizations'), ('purge_events'), ('guild_settings'), ('access_roles'), ('guild_departures')
),
tables AS (
    SELECT table_name FROM expected
    UNION
    SELECT table_name::text
    FROM information_schema.columns
    WHERE table_schema = current_schema() AND column_name = 'guild_id'
)
SELECT
    current_schema() AS schema,
    t.table_name,
    t.table_name IN (SELECT table_name FROM expected) AS known,
    to_regclass(quote_ident(current_schema()) || '.' || quote_ident(t.table_name)) IS NOT NULL AS present,
    (SELECT string_agg(c.column_name::text, ', ' ORDER BY c.ordinal_position)
     FROM information_schema.columns c
     WHERE c.table_schema = current_schema() AND c.table_name = t.table_name) AS columns
FROM tables t
ORDER BY known DESC, t.table_name;
