-- Add password hash to user_profiles for authentication
ALTER TABLE user_profiles ADD COLUMN password_hash TEXT NOT NULL DEFAULT '';

-- Create index on email for fast login lookups
CREATE INDEX idx_user_profiles_email ON user_profiles(email);
