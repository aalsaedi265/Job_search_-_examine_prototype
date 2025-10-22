-- Rollback authentication changes
DROP INDEX IF EXISTS idx_user_profiles_email;
ALTER TABLE user_profiles DROP COLUMN IF EXISTS password_hash;
