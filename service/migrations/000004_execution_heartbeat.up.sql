-- 000004 · execution 进程心跳（live 观察者存活探测）
--
-- last_heartbeat_at 由插件 10–15s 上报一次。scheduler 扫描
-- status=running 且心跳超过 HEARTBEAT_TIMEOUT_SECONDS（默认 60s）的行，标 aborted。

ALTER TABLE test_executions
    ADD COLUMN IF NOT EXISTS last_heartbeat_at TIMESTAMPTZ NULL;

UPDATE test_executions
   SET last_heartbeat_at = COALESCE(last_heartbeat_at, updated_at, start_time)
 WHERE last_heartbeat_at IS NULL;

CREATE INDEX IF NOT EXISTS idx_test_executions_stale_running
    ON test_executions (last_heartbeat_at)
    WHERE status = 'running';
