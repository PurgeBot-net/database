CREATE TABLE IF NOT EXISTS customizations (
    guild_id        BIGINT      PRIMARY KEY,
    bot_name        VARCHAR(32),
    bot_avatar      TEXT,
    remove_branding BOOLEAN     NOT NULL DEFAULT false,
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT now()
);

DO $$ BEGIN
    ALTER TABLE customizations ALTER COLUMN updated_by SET DEFAULT 0;
EXCEPTION WHEN undefined_column THEN NULL;
END $$;

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

CREATE TABLE IF NOT EXISTS guild_settings (
    guild_id      BIGINT  PRIMARY KEY,
    stats_enabled BOOLEAN NOT NULL DEFAULT true
);

CREATE TABLE IF NOT EXISTS access_roles (
    guild_id BIGINT NOT NULL,
    role_id  BIGINT NOT NULL,
    PRIMARY KEY (guild_id, role_id)
);

CREATE TABLE IF NOT EXISTS guild_departures (
    guild_id BIGINT      PRIMARY KEY,
    left_at  TIMESTAMPTZ NOT NULL
);
