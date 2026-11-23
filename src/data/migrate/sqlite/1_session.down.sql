DROP INDEX IF EXISTS idx_user_sessions_live;
DROP INDEX IF EXISTS idx_user_sessions_action_created;
DROP INDEX IF EXISTS idx_user_sessions_uid_created;
DROP TABLE IF EXISTS user_sessions;

DROP INDEX IF EXISTS idx_user_profiles_deleted;
DROP INDEX IF EXISTS idx_user_profiles_uid;
DROP TABLE IF EXISTS user_profiles;

DROP INDEX IF EXISTS idx_users_deleted;
DROP INDEX IF EXISTS idx_users_uname;
DROP INDEX IF EXISTS idx_users_account_live;
DROP TABLE IF EXISTS users;
