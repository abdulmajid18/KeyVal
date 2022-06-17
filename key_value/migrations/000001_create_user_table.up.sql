CREATE TABLE IF NOT EXISTS users (
    id bigserial PRIMARY KEY,
    created_at timestamp(0) with time zone NOT NULL DEFAULT NOW(),
    username text NOT NULL,
    email citext UNIQUE NOT NULL,
    dbname text NOT NULL,
    password_hash bytea NOT NULL,
    key     text  NOT NuLL,
    activated bool NOT NULL,
    version integer NOT NULL DEFAULT 1
);