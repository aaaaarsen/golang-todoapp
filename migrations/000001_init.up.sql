CREATE SCHEMA todoapp;

CREATE TABLE todoapp.users (
    id              BIGSERIAL                   PRIMARY KEY,
    version         INT             NOT NULL    DEFAULT 1,
    full_name       VARCHAR(100)    NOT NULL,
    phone_number    VARCHAR(15),
    created_at      TIMESTAMPTZ     NOT NULL    DEFAULT NOW()
);

CREATE TABLE todoapp.tasks (
    id              BIGSERIAL                   PRIMARY KEY,
    version         INT             NOT NULL    DEFAULT 1,
    user_id         BIGINT          NOT NULL    REFERENCES todoapp.users(id) ON DELETE CASCADE,
    title           VARCHAR(255)    NOT NULL,
    description     TEXT,
    deadline        TIMESTAMPTZ,
    importance      INT             NOT NULL    DEFAULT 1,
    category        TEXT            NOT NULL    DEFAULT 'personal',
    completed       BOOLEAN         NOT NULL    DEFAULT FALSE,
    completed_at    TIMESTAMPTZ,
    priority_score  FLOAT,
    priority_level  TEXT,
    priority_reason TEXT,
    last_updated_at TIMESTAMPTZ     NOT NULL    DEFAULT NOW(),
    created_at      TIMESTAMPTZ     NOT NULL    DEFAULT NOW()
);