-- What customization is stored for this server?
-- For a data request, run this together with guild-purges.sql and guild-settings-departure.sql.
-- No rows means nothing is stored. The avatar image itself is left out. avatar_set and avatar_bytes describe
-- it, and cmd/exportguild exports the file. other_columns shows any column beyond the five known ones, so a
-- column added later still shows up in a data request. {} means there are none.
-- This is only Postgres. Redis keys of a queued, running or interrupted purge and the service logs can also
-- hold data about the server.
-- Edit: guild_id.

WITH params AS (
    SELECT 1412752753348055111::bigint AS guild_id
)
SELECT
    c.guild_id,
    c.bot_name,
    c.bot_avatar IS NOT NULL                     AS avatar_set,
    octet_length(decode(c.bot_avatar, 'base64')) AS avatar_bytes,
    c.remove_branding,
    c.updated_at AT TIME ZONE 'UTC'              AS updated_utc,
    to_jsonb(c) - ARRAY['guild_id', 'bot_name', 'bot_avatar', 'remove_branding', 'updated_at'] AS other_columns
FROM customizations c
JOIN params p ON c.guild_id = p.guild_id;
