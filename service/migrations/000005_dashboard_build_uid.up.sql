ALTER TABLE automation_dashboard_execution_snapshots
    ADD COLUMN IF NOT EXISTS build_uid UUID NULL;

CREATE INDEX IF NOT EXISTS idx_automation_dashboard_execution_snapshots_build_uid
    ON automation_dashboard_execution_snapshots (build_uid);
