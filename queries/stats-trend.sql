-- How many purges finished, and how many messages did they delete, per UTC day, week or month?
-- Only stored rows count, and the newest period is still in progress. Older periods can shrink later too,
-- because a server that turns stats off loses its rows and a server the bot left is deleted after 30 days.
-- Weeks start on Monday.
-- Edit: bucket (day, week or month) and periods (how many to show, counting back from the current one).

WITH params AS (
    SELECT 'month'::text AS bucket, 12 AS periods
),
series AS (
    SELECT gs AS period_start
    FROM params p,
         generate_series(
             date_trunc(p.bucket, now() AT TIME ZONE 'UTC') - (p.periods - 1) * ('1 ' || p.bucket)::interval,
             date_trunc(p.bucket, now() AT TIME ZONE 'UTC'),
             ('1 ' || p.bucket)::interval
         ) AS gs
),
agg AS (
    SELECT
        date_trunc(p.bucket, pe.created_at AT TIME ZONE 'UTC') AS period_start,
        count(*)                                               AS purges,
        sum(pe.deleted)                                        AS messages_deleted
    FROM purge_events pe
    CROSS JOIN params p
    WHERE pe.created_at >= (SELECT min(period_start) FROM series) AT TIME ZONE 'UTC'
    GROUP BY 1
)
SELECT
    s.period_start::date            AS period_start_utc,
    COALESCE(a.purges, 0)           AS purges,
    COALESCE(a.messages_deleted, 0) AS messages_deleted
FROM series s
LEFT JOIN agg a ON a.period_start = s.period_start
ORDER BY s.period_start;
