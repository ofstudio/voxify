-- Drop trigger first
DROP TRIGGER IF EXISTS processes_updated_at;

-- Drop indexes
DROP INDEX IF EXISTS idx_processes_updated_at;
DROP INDEX IF EXISTS idx_processes_created_at;
DROP INDEX IF EXISTS idx_processes_request_url;
DROP INDEX IF EXISTS idx_processes_step;
DROP INDEX IF EXISTS idx_processes_status;
DROP INDEX IF EXISTS idx_episodes_created_at;
DROP INDEX IF EXISTS idx_episodes_original_url;

-- Drop tables (processes first due to foreign key)
DROP TABLE IF EXISTS processes;
DROP TABLE IF EXISTS episodes;
