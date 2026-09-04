-- 000001 · 回滚：删除全部 15 张表
-- 顺序与 up 相反，先删依赖方（此处无 FK 依赖，按上下文倒序即可）

DROP INDEX IF EXISTS idx_sys_dict_type;
DROP TABLE IF EXISTS sys_dict;

DROP INDEX IF EXISTS idx_automation_dashboard_execution_item_snapshots_execution;
DROP TABLE IF EXISTS automation_dashboard_execution_item_snapshots;

DROP INDEX IF EXISTS idx_automation_dashboard_execution_snapshots_status;
DROP TABLE IF EXISTS automation_dashboard_execution_snapshots;

DROP TABLE IF EXISTS automation_dashboard_daily_trends;
DROP TABLE IF EXISTS automation_dashboard_summary_snapshots;

DROP TABLE IF EXISTS execution_case_step_attachments;

DROP INDEX IF EXISTS idx_execution_item_steps_case_path;
DROP INDEX IF EXISTS idx_execution_item_steps_case_parent;
DROP TABLE IF EXISTS execution_item_steps;

DROP INDEX IF EXISTS uq_execution_items_case_uid;
DROP TABLE IF EXISTS execution_items;

DROP INDEX IF EXISTS uq_test_executions_build_uid;
DROP TABLE IF EXISTS test_executions;

DROP INDEX IF EXISTS uq_jenkins_pipelines_last_build_uid;
DROP INDEX IF EXISTS uq_jenkins_pipelines_job_name;
DROP TABLE IF EXISTS jenkins_pipelines;

DROP INDEX IF EXISTS idx_role_permissions_perm_id;
DROP TABLE IF EXISTS role_permissions;

DROP INDEX IF EXISTS idx_user_roles_role_id;
DROP TABLE IF EXISTS user_roles;

DROP INDEX IF EXISTS uq_permissions_code;
DROP TABLE IF EXISTS permissions;

DROP INDEX IF EXISTS uq_roles_code;
DROP TABLE IF EXISTS roles;

DROP INDEX IF EXISTS ix_users_id;
DROP INDEX IF EXISTS uq_users_email;
DROP INDEX IF EXISTS uq_users_username;
DROP TABLE IF EXISTS users;
