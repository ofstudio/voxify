-- Drop indexes
DROP INDEX IF EXISTS idx_episodes_created_at;
DROP INDEX IF EXISTS idx_episodes_original_url;
DROP INDEX IF EXISTS idx_episodes_canonical_url;

-- Drop tables
DROP TABLE IF EXISTS episodes;
