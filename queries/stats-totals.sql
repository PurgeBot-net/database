-- How many finished purges are stored, how many messages did they delete, and what does a typical one look like?
-- These are the rows stored right now, not lifetime totals. A server that turns stats off loses its old rows,
-- and a server the bot left more than 30 days ago is gone too. Cancelled purges never get a row, and neither
-- do purges that never finish.
-- Durations count from the first start, so a purge that resumed after a restart includes the downtime.
-- Edit: days. 0 means everything stored, anything else means the last N days.

WITH params AS (
    SELECT 0 AS days
),
e AS (
    SELECT pe.deleted, pe.duration_ms, pe.created_at
    FROM purge_events pe
    CROSS JOIN params p
    WHERE p.days = 0 OR pe.created_at >= now() - p.days * interval '1 day'
)
SELECT
    count(*)                                                                               AS purges,
    COALESCE(sum(deleted), 0)                                                              AS messages_deleted,
    round(avg(deleted), 1)                                                                 AS avg_deleted,
    round((percentile_cont(0.5) WITHIN GROUP (ORDER BY deleted))::numeric, 1)              AS median_deleted,
    max(deleted)                                                                           AS largest_purge,
    count(*) FILTER (WHERE deleted = 0)                                                    AS purges_deleting_nothing,
    round((percentile_cont(0.5) WITHIN GROUP (ORDER BY duration_ms))::numeric / 1000, 1)  AS median_duration_s,
    round((percentile_cont(0.95) WITHIN GROUP (ORDER BY duration_ms))::numeric / 1000, 1) AS p95_duration_s,
    min(created_at) AT TIME ZONE 'UTC'                                                     AS first_stored_utc,
    max(created_at) AT TIME ZONE 'UTC'                                                     AS last_stored_utc
FROM e;
