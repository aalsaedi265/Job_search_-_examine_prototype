-- Rollback caching fields
DROP INDEX IF EXISTS idx_jobs_cached_at;
DROP INDEX IF EXISTS idx_jobs_search_hash;

ALTER TABLE jobs DROP COLUMN IF EXISTS cached_at;
ALTER TABLE jobs DROP COLUMN IF EXISTS search_params_hash;
