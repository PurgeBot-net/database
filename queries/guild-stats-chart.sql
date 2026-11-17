-- What do the bars on the /stats chart show for this server?
-- Grouping and gap-filling copy fillDailyBars and fillMonthlyBars in stats.go, so the bars should match the bot.
-- days 7, 14, 30 or 90 gives one bar per UTC day, ending today. days 0 (the /stats default) gives 12 monthly
-- bars, ending with this month.
-- Purges from the partial first day of the window, or from the month 12 months back, count in
-- guild-stats-totals.sql but get no bar. /stats does the same. With days 0 the totals also include everything
-- older than 12 months, so the bars can add up to a lot less.
-- /stats draws no chart when the total messages deleted is 0, and shows the opt-out message when stats are
-- off. guild-settings-departure.sql has the opt-out state.
-- Near a daylight saving change, the first daily bar can differ from /stats if the session time zone is not
-- the server default the bot runs with. guild-stats-totals.sql shows session_tz.
-- Edit: guild_id and days.

WITH params AS (
    SELECT 1412752753348055111::bigint AS guild_id, 0 AS days
),
daily AS (
    SELECT date_trunc('day', pe.created_at AT TIME ZONE 'UTC')::date AS bar, sum(pe.deleted) AS deleted
    FROM purge_events pe
    CROSS JOIN params p
    WHERE p.days > 0 AND pe.guild_id = p.guild_id AND pe.created_at >= now() - (p.days * interval '1 day')
    GROUP BY 1
),
monthly AS (
    SELECT date_trunc('month', pe.created_at AT TIME ZONE 'UTC')::date AS bar, sum(pe.deleted) AS deleted
    FROM purge_events pe
    CROSS JOIN params p
    WHERE p.days = 0 AND pe.guild_id = p.guild_id AND pe.created_at >= now() - interval '12 months'
    GROUP BY 1
),
bars AS (
    SELECT gs::date AS bar
    FROM params p,
         generate_series(
             date_trunc('day', now() AT TIME ZONE 'UTC') - (p.days - 1) * interval '1 day',
             date_trunc('day', now() AT TIME ZONE 'UTC'),
             interval '1 day'
         ) AS gs
    WHERE p.days > 0
    UNION ALL
    SELECT gs::date
    FROM params p,
         generate_series(
             date_trunc('month', now() AT TIME ZONE 'UTC') - interval '11 months',
             date_trunc('month', now() AT TIME ZONE 'UTC'),
             interval '1 month'
         ) AS gs
    WHERE p.days = 0
)
SELECT
    b.bar                                AS bar_start_utc,
    COALESCE(d.deleted, m.deleted, 0)    AS messages_deleted
FROM bars b
LEFT JOIN daily d ON d.bar = b.bar
LEFT JOIN monthly m ON m.bar = b.bar
ORDER BY b.bar;
