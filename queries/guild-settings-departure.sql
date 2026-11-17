-- Has this server turned stats off, and is it marked as departed and waiting for the sweeper?
-- For a data request, run this together with guild-customization.sql and guild-purges.sql.
-- stats_enabled NULL means there is no settings row, so stats are on (the default). false means opted out.
-- No departure row does not mean the bot is in the server. Only servers with stored data ever get one.
-- waiting: left_at is less than 720 hours (30 days) ago.
-- due: older than that. Every hourly sweep asks Discord about the 100 oldest due rows and only deletes the
-- data if Discord says the server is gone. If the bot is still there, the row is cleared and the data stays.
-- operator delete: left_at is 1970, set by cmd/deleteguild. The next sweep handles it.
-- left_at is when a sweep with a complete server list first noticed the server missing. With the sweeper
-- running normally that is within an hour of leaving, and later if it was down or refusing. For servers that
-- left before the sweeper first ran, it is the time of that first sweep.
-- This is only Postgres. Redis keys of a queued, running or interrupted purge and the service logs can also
-- hold data about the server.
-- Edit: guild_id.

WITH params AS (
    SELECT 1412752753348055111::bigint AS guild_id
)
SELECT
    p.guild_id,
    s.stats_enabled,
    d.left_at AT TIME ZONE 'UTC' AS left_utc,
    CASE WHEN d.left_at > 'epoch' THEN (d.left_at + interval '720 hours') AT TIME ZONE 'UTC' END AS deletable_after_utc,
    CASE
        WHEN d.guild_id IS NULL                      THEN 'no departure row'
        WHEN d.left_at = 'epoch'                     THEN 'operator delete'
        WHEN d.left_at < now() - interval '720 hours' THEN 'due'
        ELSE 'waiting'
    END AS departure_state
FROM params p
LEFT JOIN guild_settings s ON s.guild_id = p.guild_id
LEFT JOIN guild_departures d ON d.guild_id = p.guild_id;
