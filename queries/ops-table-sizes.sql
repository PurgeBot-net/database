-- How many rows does each table hold, how much disk does it use, and how big is the whole database?
-- row_count is exact. dead_rows is an estimate of deleted or updated rows that vacuum has not handled yet.
-- Vacuum only marks that space as reusable. The old bytes from an opt-out or a sweep stay in the files until
-- something overwrites them or VACUUM FULL runs, and WAL and backups keep copies for a while.
-- toast_size is where large values like avatars live. The TOAST table gets vacuumed on its own, so
-- last_vacuum_utc says nothing about it. total_size includes indexes and TOAST.
-- It covers every table in the current schema that this login can read. The whole database row only has
-- sizes.
-- Nothing to edit.

SELECT
    s.relname                                                          AS table_name,
    (xpath('/row/n/text()', query_to_xml(
        format('SELECT count(*) AS n FROM %I.%I', s.schemaname, s.relname), false, true, '')))[1]::text::bigint AS row_count,
    s.n_dead_tup                                                       AS dead_rows,
    pg_size_pretty(pg_total_relation_size(s.relid))                    AS total_size,
    pg_size_pretty(pg_relation_size(s.relid))                          AS table_size,
    pg_size_pretty(pg_indexes_size(s.relid))                           AS index_size,
    pg_size_pretty(COALESCE(pg_total_relation_size(NULLIF(c.reltoastrelid, 0)), 0)) AS toast_size,
    pg_total_relation_size(s.relid)                                    AS total_bytes,
    greatest(s.last_vacuum, s.last_autovacuum) AT TIME ZONE 'UTC'      AS last_vacuum_utc,
    greatest(s.last_analyze, s.last_autoanalyze) AT TIME ZONE 'UTC'    AS last_analyze_utc
FROM pg_stat_user_tables s
JOIN pg_class c ON c.oid = s.relid
WHERE s.schemaname = current_schema()
  AND has_table_privilege(s.relid, 'SELECT')
UNION ALL
SELECT
    '(whole database)', NULL, NULL,
    pg_size_pretty(pg_database_size(current_database())), NULL, NULL, NULL,
    pg_database_size(current_database()), NULL, NULL
ORDER BY total_bytes DESC;
