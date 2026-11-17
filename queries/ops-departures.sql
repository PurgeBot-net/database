-- How big is the sweeper backlog, and is the sweeper keeping up?
-- One row per state. A missing state has nothing in it.
-- waiting: marked as departed less than 720 hours (30 days) ago. next_due_utc is when the first one is due.
-- due: older than that. Every hourly sweep asks Discord about up to 100 due or operator delete rows, oldest
-- first, and deletes the servers Discord says are gone. If the bot is still in one, its row is cleared and
-- the data stays. hourly_sweeps_to_clear counts both states.
-- oldest_overdue is how far the oldest due row is past its 720 hours. With 100 due rows or fewer, anything
-- over about 2 hours means trouble: the sweeper is down, it is refusing an incomplete server list (it logs
-- an error), or Discord keeps returning an unexpected error for that server. With more than 100, rows marked
-- at the same time wait their turn and oldest_overdue grows by about an hour per sweep. That is fine as long
-- as departure_rows drops by about 100 an hour.
-- operator delete: left_at is 1970, set by cmd/deleteguild. The next sweep clears or deletes it, so if one is
-- still here an hour after deleteguild, the sweeper is not running or its server list keeps failing.
-- Nothing to edit.

WITH d AS (
    SELECT
        left_at,
        CASE
            WHEN left_at = 'epoch'                      THEN 'operator delete'
            WHEN left_at < now() - interval '720 hours' THEN 'due'
            ELSE 'waiting'
        END AS state
    FROM guild_departures
)
SELECT
    state,
    count(*)                         AS departure_rows,
    min(left_at) AT TIME ZONE 'UTC'  AS oldest_left_utc,
    max(left_at) AT TIME ZONE 'UTC'  AS newest_left_utc,
    CASE WHEN state = 'waiting' THEN (min(left_at) + interval '720 hours') AT TIME ZONE 'UTC' END AS next_due_utc,
    CASE WHEN state = 'due' THEN now() - (min(left_at) + interval '720 hours') END                AS oldest_overdue,
    CASE WHEN state = 'due' THEN ceil(sum(count(*)) FILTER (WHERE state <> 'waiting') OVER () / 100.0) END AS hourly_sweeps_to_clear
FROM d
GROUP BY state
ORDER BY CASE state WHEN 'operator delete' THEN 1 WHEN 'due' THEN 2 ELSE 3 END;
