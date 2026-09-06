DROP INDEX IF EXISTS idx_automation_dashboard_execution_snapshots_build_uid;

ALTER TABLE automation_dashboard_execution_snapshots
    DROP COLUMN IF EXISTS build_uid;
