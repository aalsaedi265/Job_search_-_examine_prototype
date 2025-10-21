-- Add columns for pause/resume functionality
ALTER TABLE applications
    ADD COLUMN IF NOT EXISTS paused_at TIMESTAMPTZ,
    ADD COLUMN IF NOT EXISTS current_url TEXT;

-- Update status column to support new states
-- Valid states: pending, in_progress, paused, submitted, failed, cancelled, timeout
COMMENT ON COLUMN applications.status IS 'Application state: pending, in_progress, paused, submitted, failed, cancelled, timeout';

-- Add index for paused applications
CREATE INDEX IF NOT EXISTS idx_applications_paused ON applications(status, paused_at) WHERE status = 'paused';
