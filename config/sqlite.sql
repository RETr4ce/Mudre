PRAGMA foreign_keys = off;
BEGIN TRANSACTION;

-- Table: crackmesLatest
DROP TABLE IF EXISTS crackmesLatest;

CREATE TABLE crackmesLatest (
    id          INTEGER PRIMARY KEY AUTOINCREMENT,
    link        STRING,
    title       STRING,
    description STRING,
    author      STRING,
    category    STRING,
    guid        STRING  UNIQUE ON CONFLICT FAIL,
    pubDate     STRING,
    pushDiscord BOOLEAN
);


-- Table: ctftimeEvents
DROP TABLE IF EXISTS ctftimeEvents;

CREATE TABLE ctftimeEvents (
    id           INT64        UNIQUE ON CONFLICT FAIL,
    title        STRING (100),
    link         STRING,
    start        DATETIME,
    finish       DATETIME,
    description  STRING,
    format       STRING,
    logo         STRING,
    restrictions STRING,
    onsite       BOOLEAN,
    pushDiscord  BOOLEAN
);


-- Table: ctftimeWriteup
DROP TABLE IF EXISTS ctftimeWriteup;

CREATE TABLE ctftimeWriteup (
    id          INTEGER      PRIMARY KEY AUTOINCREMENT,
    link        STRING (50)  UNIQUE ON CONFLICT FAIL,
    title       STRING (100),
    originalUrl STRING,
    lastBuild   DATETIME,
    pushDiscord BOOLEAN
);


COMMIT TRANSACTION;
PRAGMA foreign_keys = on;
PRAGMA journal_mode = WAL;