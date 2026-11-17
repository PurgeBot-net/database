-- What do purges target, by number of purges and by messages deleted?
-- Targets are server (every channel), category and channel. A target with no stored purge in the period
-- has no row.
-- Only finished purges that are still stored count. Servers with stats off record nothing, and a server is
-- deleted 30 days after the bot leaves, so the shares describe what is stored today.
-- Edit: days. 0 means everything stored, anything else means the last N days.

WITH params AS (
    SELECT 0 AS days
),
e AS (
    SELECT pe.target_type, pe.deleted, pe.duration_ms
    FROM purge_events pe
    CROSS JOIN params p
    WHERE p.days = 0 OR pe.created_at >= now() - p.days * interval '1 day'
)
SELECT
    target_type,
    count(*)                                                                              AS purges,
    round(100.0 * count(*) / sum(count(*)) OVER (), 1)                                    AS pct_of_purges,
    sum(deleted)                                                                          AS messages_deleted,
    round(100.0 * sum(deleted) / NULLIF(sum(sum(deleted)) OVER (), 0), 1)                 AS pct_of_messages,
    round((percentile_cont(0.5) WITHIN GROUP (ORDER BY deleted))::numeric, 1)             AS median_deleted,
    round((percentile_cont(0.5) WITHIN GROUP (ORDER BY duration_ms))::numeric / 1000, 1) AS median_duration_s
FROM e
GROUP BY target_type
ORDER BY purges DESC, target_type;
