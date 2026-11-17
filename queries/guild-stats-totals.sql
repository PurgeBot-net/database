-- What does /stats show for this server?
-- This is the same SQL as GetGuildStats in stats.go, so the numbers should match the bot.
-- days 0 is the /stats default and covers every stored row. 7, 14, 30 or 90 counts back from now, and that
-- window uses the session time zone. With days above 0 the totals only match /stats when session_tz is the
-- same as the server default the bot runs with.
-- /stats says nothing is recorded when purges is 0, and only draws a chart when messages_deleted is above 0.
-- With stats off it shows the opt-out message instead, and without Premium it shows an upsell.
-- guild-settings-departure.sql has the opt-out state.
-- The chart (guild-stats-chart.sql) can add up to less. It drops the partial first day of the window, and
-- with days 0 it only covers the last 12 months while these totals cover everything.
-- Edit: guild_id and days.

WITH params AS (
    SELECT 1412752753348055111::bigint AS guild_id, 0 AS days
)
SELECT
    count(*)                              AS purges,
    COALESCE(sum(pe.deleted), 0)          AS messages_deleted,
    max(pe.created_at) AT TIME ZONE 'UTC' AS last_purge_utc,
    current_setting('TimeZone')           AS session_tz
FROM purge_events pe
CROSS JOIN params p
WHERE pe.guild_id = p.guild_id
  AND (p.days = 0 OR pe.created_at >= now() - (p.days * interval '1 day'));
