CREATE TABLE IF NOT EXISTS users (
                                     id SERIAL PRIMARY KEY,
                                     email TEXT UNIQUE NOT NULL,
                                     pass_hash BYTEA NOT NULL,
                                     is_admin BOOLEAN DEFAULT FALSE
);

CREATE TABLE IF NOT EXISTS apps (
                                    id SERIAL PRIMARY KEY,
                                    name TEXT NOT NULL,
                                    secret TEXT NOT NULL
);
