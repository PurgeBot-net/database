CREATE TABLE IF NOT EXISTS customizations (
    guild_id        BIGINT      PRIMARY KEY,
    bot_name        VARCHAR(32),
    bot_avatar      TEXT,
    remove_branding BOOLEAN     NOT NULL DEFAULT false,
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_by      BIGINT      NOT NULL
);

CREATE TABLE IF NOT EXISTS purge_events (
    id          BIGSERIAL   PRIMARY KEY,
    guild_id    BIGINT      NOT NULL,
    purge_type  VARCHAR(16) NOT NULL,
    target_type VARCHAR(16) NOT NULL,
    deleted     INT         NOT NULL,
    duration_ms INT         NOT NULL,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS purge_events_guild_id_idx ON purge_events (guild_id);
