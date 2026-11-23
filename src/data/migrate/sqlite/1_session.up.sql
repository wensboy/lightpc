-- Accounts, their profile rows and the session history.
--
-- The row identity is always an auto increment id, never a business value: an insert lets the table choose it,
-- and uid stays a separate unique identifier so a reference survives a rename of the account.
--
-- There is no foreign key on purpose. The link from user_profiles and user_sessions to users is logical, through
-- uid, which keeps a table rebuild cheap and lets a row survive while its account is only soft deleted.

CREATE TABLE IF NOT EXISTS users (
    id         INTEGER PRIMARY KEY AUTOINCREMENT,
    uid        TEXT    NOT NULL UNIQUE, -- hash(account, timestamp, random number) -> 32 char
    uname      TEXT    NOT NULL CHECK (length(trim(uname)) > 0),
    account    TEXT    NOT NULL CHECK (length(trim(account)) > 0),  -- email or phone, current is email
    password   TEXT    NOT NULL CHECK (length(password) > 0),       -- hashed password
    created_at INTEGER NOT NULL CHECK (created_at > 0),
    updated_at INTEGER NOT NULL CHECK (updated_at >= created_at),
    deleted_at INTEGER                                              -- deleted timestamp
);

-- An account is unique among the live rows only: a removed account must be able to register the same address
-- again, which a plain unique column would reject.
CREATE UNIQUE INDEX IF NOT EXISTS idx_users_account_live ON users (account) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_users_uname ON users (uname) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_users_deleted ON users (deleted_at);

CREATE TABLE IF NOT EXISTS user_profiles (
    id         INTEGER PRIMARY KEY AUTOINCREMENT,
    uid        TEXT    NOT NULL,                                    -- logical reference to users.uid
    email      TEXT    NOT NULL DEFAULT '',
    phone      TEXT    NOT NULL DEFAULT '',
    avatar     TEXT    NOT NULL DEFAULT '',                         -- url or local relative path
    nickname   TEXT    NOT NULL DEFAULT '',
    gender     TEXT    NOT NULL DEFAULT '' CHECK (gender IN ('', 'male', 'female', 'other')),
    signature  TEXT    NOT NULL DEFAULT '',                         -- personal signature
    created_at INTEGER NOT NULL CHECK (created_at > 0),
    updated_at INTEGER NOT NULL CHECK (updated_at >= created_at),
    deleted_at INTEGER
);

CREATE INDEX IF NOT EXISTS idx_user_profiles_uid ON user_profiles (uid) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_user_profiles_deleted ON user_profiles (deleted_at);

CREATE TABLE IF NOT EXISTS user_sessions (
    id         INTEGER PRIMARY KEY AUTOINCREMENT,
    uid        TEXT    NOT NULL,                                    -- logical reference to users.uid
    account    TEXT    NOT NULL,                                    -- logical reference to users.account
    action     TEXT    NOT NULL CHECK (action IN ('sign_in', 'sign_out', 'sign_up', 'sign_down')),
    expiration INTEGER NOT NULL DEFAULT 0 CHECK (expiration >= 0),  -- seconds; only meaningful for sign_in
    created_at INTEGER NOT NULL CHECK (created_at > 0)
);

-- The latest row per account answers "when did this account last sign in", which is what the automatic
-- sign in check reads at start up.
CREATE INDEX IF NOT EXISTS idx_user_sessions_uid_created ON user_sessions (uid, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_user_sessions_action_created ON user_sessions (action, created_at DESC);
-- The live window is computed from created_at and expiration of a sign in row, so the pair is indexed.
CREATE INDEX IF NOT EXISTS idx_user_sessions_live ON user_sessions (uid, created_at, expiration) WHERE action = 'sign_in';
