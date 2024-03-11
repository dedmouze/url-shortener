CREATE TABLE IF NOT EXISTS url
(
    id    INTEGER PRIMARY KEY,
    alias TEXT    NOT NULL UNIQUE,
    url   TEXT    NOT NULL
);
CREATE INDEX IF NOT EXISTS idx_alias ON url(alias);

CREATE TABLE IF NOT EXISTS client
(
    id      INTEGER PRIMARY KEY,
    name    TEXT    NOT NULL UNIQUE,
    apiKey  TEXT    NOT NULL,
    userKey TEXT    NOT NULL
);
CREATE INDEX IF NOT EXISTS idx_name ON client(name);