-- 000003 · 补充查询索引（缺陷 S-5）
--
-- 现有 schema 缺少支撑高频查询的索引：
--   - cases/history 按 (case_key, start_time) 过滤排序，当前全表扫
--   - Dashboard 时间窗聚合与运行中 execution 查询按 (status, start_time)
--   - 执行详情下的用例列表按 build_uid 过滤 is_latest
--   - Dashboard 活跃任务数按 status 过滤
--   - 按步骤取附件
--
-- 全部 IF NOT EXISTS，对存量库幂等。

CREATE INDEX IF NOT EXISTS idx_execution_items_case_key_start
    ON execution_items (case_key, start_time DESC);

CREATE INDEX IF NOT EXISTS idx_test_executions_status_start
    ON test_executions (status, start_time DESC);

CREATE INDEX IF NOT EXISTS idx_execution_items_build_latest
    ON execution_items (build_uid) WHERE is_latest;

CREATE INDEX IF NOT EXISTS idx_jenkins_pipelines_status
    ON jenkins_pipelines (status);

CREATE INDEX IF NOT EXISTS idx_step_attachments_step_id
    ON execution_case_step_attachments (step_id);

CREATE INDEX IF NOT EXISTS idx_step_attachments_item_id
    ON execution_case_step_attachments (item_id) WHERE item_id IS NOT NULL;
