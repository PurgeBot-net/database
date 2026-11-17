-- How much is /customize used?
-- Each row is a server that saved a customization and still has it. Clearing it, cmd/deleteguild or the bot
-- being gone for 30 days removes the row. Saving needs Premium, but this counts stored rows, not servers that
-- have Premium right now.
-- Nothing to edit.

SELECT
    count(*)                                                         AS stored_customizations,
    count(*) FILTER (WHERE bot_name IS NOT NULL)                     AS with_name,
    count(*) FILTER (WHERE bot_avatar IS NOT NULL)                   AS with_avatar,
    count(*) FILTER (WHERE remove_branding)                          AS branding_removed,
    count(*) FILTER (WHERE updated_at >= now() - interval '30 days') AS changed_last_30_days
FROM customizations;
