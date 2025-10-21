-- Remove pause/resume columns
DROP INDEX IF EXISTS idx_applications_paused;

ALTER TABLE applications
    DROP COLUMN IF EXISTS paused_at,
    DROP COLUMN IF EXISTS current_url;
