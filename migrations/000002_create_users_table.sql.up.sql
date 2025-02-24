CREATE TABLE IF NOT EXISTS users (
    user_id         integer         PRIMARY KEY GENERATED ALWAYS AS IDENTITY,
    first_name      varchar(50)     NOT NULL,
    last_name       varchar(50)     NOT NULL,
    email           citext          UNIQUE NOT NULL,
    password_hash   bytea           NOT NULL,
    created_at      timestamptz     NOT NULL DEFAULT NOW(),
    updated_at      timestamptz,
    active          boolean         NOT NULL DEFAULT true,
    deleted_at      timestamptz,
    activated       boolean         NOT NULL DEFAULT false
    );
