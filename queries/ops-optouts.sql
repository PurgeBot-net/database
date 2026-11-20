-- How many servers have turned stats off, and does any of them still have purge rows?
-- These are the opt-outs stored now, not every opt-out ever. A server that opted out and then left is
-- deleted 30 days later, and cmd/deleteguild removes a turned-back-on row.
-- turned_back_on counts servers that opted out once and later ran /settings stats state:On.
-- opted_out_with_purge_rows should be 0. If it is not, something is still recording purges for an opted-out
-- server, an old purger image for example. leaking_guild_ids lists them.
-- Nothing to edit.

WITH s AS (
    SELECT
        gs.guild_id,
        gs.stats_enabled,
        EXISTS (SELECT 1 FROM purge_events pe WHERE pe.guild_id = gs.guild_id) AS has_purge_rows
    FROM guild_settings gs
)
SELECT
    count(*)                                                                  AS settings_rows,
    count(*) FILTER (WHERE NOT stats_enabled)                                 AS opted_out,
    count(*) FILTER (WHERE stats_enabled)                                     AS turned_back_on,
    count(*) FILTER (WHERE NOT stats_enabled AND has_purge_rows)              AS opted_out_with_purge_rows,
    array_agg(guild_id ORDER BY guild_id) FILTER (WHERE NOT stats_enabled AND has_purge_rows) AS leaking_guild_ids
FROM s;
