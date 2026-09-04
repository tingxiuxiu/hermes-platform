-- 000004 · 回滚心跳列

DROP INDEX IF EXISTS idx_test_executions_stale_running;

ALTER TABLE test_executions
    DROP COLUMN IF EXISTS last_heartbeat_at;
