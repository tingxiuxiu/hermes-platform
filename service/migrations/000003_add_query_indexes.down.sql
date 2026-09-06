-- 000003 · 回滚：删除补充索引

DROP INDEX IF EXISTS idx_step_attachments_item_id;
DROP INDEX IF EXISTS idx_step_attachments_step_id;
DROP INDEX IF EXISTS idx_jenkins_pipelines_status;
DROP INDEX IF EXISTS idx_execution_items_build_latest;
DROP INDEX IF EXISTS idx_test_executions_status_start;
DROP INDEX IF EXISTS idx_execution_items_case_key_start;
