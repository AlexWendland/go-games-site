-- +goose Up

CREATE TABLE users (
    id             INTEGER PRIMARY KEY AUTOINCREMENT,
    user_id        TEXT      NOT NULL UNIQUE,
    display_name   TEXT      NOT NULL,
    created_at     TIMESTAMP NOT NULL,
    is_active      BOOLEAN   NOT NULL CHECK (is_active IN (0, 1))
);

CREATE TABLE sessions (
    id            INTEGER PRIMARY KEY AUTOINCREMENT,
    user_id       INTEGER   NOT NULL REFERENCES users(id),
    session_token TEXT      NOT NULL UNIQUE,
    created_at    TIMESTAMP NOT NULL,
    expires_at    TIMESTAMP NOT NULL
);

CREATE INDEX idx_sessions_session_token ON sessions (session_token);
CREATE INDEX idx_sessions_user_id ON sessions (user_id);

CREATE TABLE ai_users (
    id        INTEGER PRIMARY KEY AUTOINCREMENT,
    game_type TEXT NOT NULL,
    ai_type   TEXT NOT NULL,
    name      TEXT NOT NULL
);

CREATE TABLE games (
    id            INTEGER PRIMARY KEY AUTOINCREMENT,
    game_id       TEXT      NOT NULL UNIQUE,
    game_type     TEXT      NOT NULL,
    status        TEXT      NOT NULL CHECK (status IN ('open', 'finished')),
    created_by    INTEGER   NOT NULL REFERENCES users(id),
    created_at    TIMESTAMP NOT NULL,
    updated_at    TIMESTAMP NOT NULL,
    finished_at   TIMESTAMP,
    state_version INTEGER   NOT NULL,
    state         JSON
);

CREATE TABLE positions (
    id         INTEGER PRIMARY KEY AUTOINCREMENT,
    game_id    INTEGER   NOT NULL REFERENCES games(id),
    user_id    INTEGER            REFERENCES users(id),
    ai_user_id INTEGER            REFERENCES ai_users(id),
    position   INTEGER   NOT NULL,
    joined_at  TIMESTAMP NOT NULL,
    CHECK (
        (user_id IS NOT NULL AND ai_user_id IS NULL) OR
        (user_id IS NULL AND ai_user_id IS NOT NULL)
    )
);

CREATE TABLE game_events (
    id              INTEGER PRIMARY KEY AUTOINCREMENT,
    game_id         INTEGER   NOT NULL REFERENCES games(id),
    position        INTEGER   NOT NULL,
    sequence_number INTEGER   NOT NULL,
    event_version   INTEGER   NOT NULL,
    event_type      TEXT      NOT NULL,
    data            JSON,
    created_at      TIMESTAMP NOT NULL
);

CREATE UNIQUE INDEX idx_game_events_sequence ON game_events (game_id, sequence_number);

-- +goose Down

DROP INDEX idx_game_events_sequence;
DROP INDEX idx_sessions_session_token;
DROP INDEX idx_sessions_user_id;
DROP TABLE game_events;
DROP TABLE positions;
DROP TABLE games;
DROP TABLE ai_users;
DROP TABLE sessions;
DROP TABLE users;
