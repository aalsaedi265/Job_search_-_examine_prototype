-- Add caching fields to jobs table for efficient API quota management
ALTER TABLE jobs ADD COLUMN IF NOT EXISTS search_params_hash VARCHAR(64);
ALTER TABLE jobs ADD COLUMN IF NOT EXISTS cached_at TIMESTAMP DEFAULT NOW();

-- Index for fast cache lookups
CREATE INDEX IF NOT EXISTS idx_jobs_search_hash ON jobs(search_params_hash);
CREATE INDEX IF NOT EXISTS idx_jobs_cached_at ON jobs(cached_at);
