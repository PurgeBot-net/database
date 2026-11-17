-- Which purges are stored for this server?
-- For a data request, run this together with guild-customization.sql and guild-settings-departure.sql.
-- Only finished purges are here. The table has nothing on who ran a purge, which channels it hit or which
-- filter it used. No rows means nothing is stored: the server never finished a purge, turned stats off, was
-- erased with cmd/deleteguild, or was deleted 30 days after the bot left. recorded_utc is when the purge
-- finished.
-- This is only Postgres. While a purge is queued, running or interrupted, Redis holds its job (including who
-- started it, the filter and channel IDs), and the service logs can hold data about the server too.
-- Edit: guild_id.

WITH params AS (
    SELECT 1412752753348055111::bigint AS guild_id
)
SELECT
    pe.id,
    pe.created_at AT TIME ZONE 'UTC'  AS recorded_utc,
    pe.purge_type,
    pe.target_type,
    pe.deleted,
    round(pe.duration_ms / 1000.0, 1) AS duration_s
FROM purge_events pe
JOIN params p ON pe.guild_id = p.guild_id
ORDER BY pe.created_at DESC, pe.id DESC;
