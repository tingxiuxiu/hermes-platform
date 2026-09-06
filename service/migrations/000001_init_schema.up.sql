-- 000001 · 从零建库：15 张表
--
-- schema 基准 = service/src/app/**/models.py（SQLAlchemy 模型），见 ADR-0003。
-- 注意：本文件刻意不包含 alembic 迁移里那个错误的 sys_dict 单列 UNIQUE(dict_type)。
-- 所有表统一带 created_at / updated_at（来自 SQLAlchemy Base），
-- updated_at 由应用层在 UPDATE 时显式写 now()，不使用数据库触发器（S-4）。

-- ============================================================================
-- identity 上下文
-- ============================================================================

CREATE TABLE users (
    id             BIGSERIAL    PRIMARY KEY,
    username       VARCHAR(64)  NOT NULL,
    email          VARCHAR(255) NOT NULL,
    password_hash  TEXT         NOT NULL,
    -- 0=正常 1=禁用 2=已删除(软删)；models.py 注释里的 -1/-2 是错的（缺陷 D-01）
    status         SMALLINT     NOT NULL,
    metadata       JSONB        NOT NULL DEFAULT '{}',
    last_login_at  TIMESTAMPTZ  NULL,
    last_login_ip  VARCHAR(45)  NULL,
    created_at     TIMESTAMPTZ  NOT NULL DEFAULT now(),
    updated_at     TIMESTAMPTZ  NOT NULL DEFAULT now(),
    CONSTRAINT chk_users_status CHECK (status IN (0, 1, 2))
);

CREATE UNIQUE INDEX uq_users_username ON users (username);
CREATE UNIQUE INDEX uq_users_email ON users (email);
-- 与 Python 侧的 ix_users_id 保持一致（与 PK 冗余，保留以兼容）
CREATE INDEX ix_users_id ON users (id);

COMMENT ON COLUMN users.status IS '0=正常 1=禁用 2=已删除(软删)';
COMMENT ON COLUMN users.password_hash IS 'Argon2id PHC 字符串';

CREATE TABLE roles (
    id           BIGSERIAL    PRIMARY KEY,
    code         VARCHAR(64)  NOT NULL,
    name         VARCHAR(64)  NOT NULL,
    description  TEXT         NULL,
    is_system    BOOLEAN      NOT NULL DEFAULT false,
    created_at   TIMESTAMPTZ  NOT NULL DEFAULT now(),
    updated_at   TIMESTAMPTZ  NOT NULL DEFAULT now()
);

CREATE UNIQUE INDEX uq_roles_code ON roles (code);

CREATE TABLE permissions (
    id             BIGSERIAL     PRIMARY KEY,
    -- 自引用，未建外键（与 models.py 一致）
    parent_id      BIGINT        NULL,
    code           VARCHAR(128)  NOT NULL,
    name           VARCHAR(64)   NOT NULL,
    -- 1=菜单 2=按钮/功能 3=API 接口
    resource_type  SMALLINT      NOT NULL DEFAULT 1,
    path           TEXT          NULL,
    method         VARCHAR(16)   NULL,
    created_at     TIMESTAMPTZ   NOT NULL DEFAULT now(),
    updated_at     TIMESTAMPTZ   NOT NULL DEFAULT now(),
    CONSTRAINT chk_permissions_type CHECK (resource_type IN (1, 2, 3))
);

CREATE UNIQUE INDEX uq_permissions_code ON permissions (code);

CREATE TABLE user_roles (
    id          BIGSERIAL   PRIMARY KEY,
    user_id     BIGINT      NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    role_id     BIGINT      NOT NULL REFERENCES roles (id) ON DELETE CASCADE,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_user_roles_role_id ON user_roles (role_id);

CREATE TABLE role_permissions (
    id             BIGSERIAL   PRIMARY KEY,
    role_id        BIGINT      NOT NULL REFERENCES roles (id) ON DELETE CASCADE,
    permission_id  BIGINT      NOT NULL REFERENCES permissions (id) ON DELETE CASCADE,
    created_at     TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at     TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_role_permissions_perm_id ON role_permissions (permission_id);

-- ============================================================================
-- automation 上下文
--
-- 4 张业务表之间刻意不建外键（与 models.py 一致）：
-- pytest 插件可能乱序上报（先报 item 后补 execution），外键会让合法上报被拒。
-- ============================================================================

CREATE TABLE jenkins_pipelines (
    id                   SERIAL       PRIMARY KEY,
    job_name             VARCHAR(255) NOT NULL,
    job_url              VARCHAR(255) NOT NULL,
    status               VARCHAR(32)  NOT NULL DEFAULT 'active',
    last_build_number    INTEGER      NULL,
    last_build_uid       UUID         NULL,
    last_build_status    VARCHAR(32)  NULL,
    last_build_timestamp TIMESTAMPTZ  NULL,
    last_build_duration  DOUBLE PRECISION NULL,
    pipeline_params      JSONB        NULL,
    sync_at              TIMESTAMPTZ  NULL DEFAULT now(),
    created_at           TIMESTAMPTZ  NOT NULL DEFAULT now(),
    updated_at           TIMESTAMPTZ  NOT NULL DEFAULT now()
);

CREATE UNIQUE INDEX uq_jenkins_pipelines_job_name ON jenkins_pipelines (job_name);
CREATE UNIQUE INDEX uq_jenkins_pipelines_last_build_uid ON jenkins_pipelines (last_build_uid);

CREATE TABLE test_executions (
    id                   SERIAL       PRIMARY KEY,
    build_uid            UUID         NOT NULL,
    job_name             VARCHAR(255) NOT NULL,
    job_url              VARCHAR(255) NULL,
    project_name         VARCHAR(255) NOT NULL,
    software_name        VARCHAR(100) NOT NULL,
    software_version     VARCHAR(100) NOT NULL,
    labels               VARCHAR(64)[] NULL,
    status               VARCHAR(32)  NOT NULL DEFAULT 'running',
    start_time           TIMESTAMPTZ  NOT NULL,
    end_time             TIMESTAMPTZ  NULL,
    duration             DOUBLE PRECISION NULL,
    planned_cases_count  INTEGER      NOT NULL DEFAULT 0,
    pass_count           INTEGER      NOT NULL DEFAULT 0,
    failure_count        INTEGER      NOT NULL DEFAULT 0,
    skipped_count        INTEGER      NOT NULL DEFAULT 0,
    pass_rate            DOUBLE PRECISION NULL,
    created_at           TIMESTAMPTZ  NOT NULL DEFAULT now(),
    updated_at           TIMESTAMPTZ  NOT NULL DEFAULT now()
);

CREATE UNIQUE INDEX uq_test_executions_build_uid ON test_executions (build_uid);

CREATE TABLE execution_items (
    id              SERIAL       PRIMARY KEY,
    build_uid       UUID         NOT NULL,
    -- 每次 attempt 生成新的 case_uid，全局唯一
    case_uid        UUID         NOT NULL,
    case_key        VARCHAR(128) NOT NULL,
    case_name       TEXT         NOT NULL,
    labels          VARCHAR(64)[] NULL,
    attempt_number  SMALLINT     NOT NULL DEFAULT 1,
    is_latest       BOOLEAN      NOT NULL DEFAULT true,
    status          VARCHAR(32)  NOT NULL DEFAULT 'running',
    start_time      TIMESTAMPTZ  NULL,
    end_time        TIMESTAMPTZ  NULL,
    duration        DOUBLE PRECISION NULL,
    error_message   TEXT         NULL,
    error_traceback TEXT         NULL,
    created_at      TIMESTAMPTZ  NOT NULL DEFAULT now(),
    updated_at      TIMESTAMPTZ  NOT NULL DEFAULT now(),
    CONSTRAINT uq_execution_items_attempt UNIQUE (build_uid, case_key, attempt_number)
);

CREATE UNIQUE INDEX uq_execution_items_case_uid ON execution_items (case_uid);

CREATE TABLE execution_item_steps (
    id                 SERIAL       PRIMARY KEY,
    case_uid           UUID         NOT NULL,
    parent_step_id     INTEGER      NULL,
    parent_step_index  INTEGER      NULL,
    step_index         INTEGER      NOT NULL DEFAULT 0,
    -- 步骤上报的幂等键，形如 0 / 0.1 / 0.1.2
    step_path          VARCHAR(255) NOT NULL,
    depth              INTEGER      NOT NULL DEFAULT 0,
    step_name          TEXT         NOT NULL,
    status             VARCHAR(32)  NOT NULL,
    start_time         TIMESTAMPTZ  NULL,
    end_time           TIMESTAMPTZ  NULL,
    duration           DOUBLE PRECISION NULL,
    created_at         TIMESTAMPTZ  NOT NULL DEFAULT now(),
    updated_at         TIMESTAMPTZ  NOT NULL DEFAULT now(),
    CONSTRAINT uq_execution_item_steps_path UNIQUE (case_uid, step_path)
);

CREATE INDEX idx_execution_item_steps_case_parent ON execution_item_steps (case_uid, parent_step_id);
CREATE INDEX idx_execution_item_steps_case_path ON execution_item_steps (case_uid, step_path);

CREATE TABLE execution_case_step_attachments (
    id               SERIAL        PRIMARY KEY,
    -- case 级附件（无对应 step）记 item_id，step 级附件记 step_id，两者互斥（评审确认项 Q1）
    item_id          INTEGER       NULL,
    step_id          INTEGER       NULL,
    attachment_type  VARCHAR(32)   NOT NULL,
    file_name        VARCHAR(255)  NULL,
    url              VARCHAR(1024) NOT NULL,
    mime_type        VARCHAR(128)  NULL,
    created_at       TIMESTAMPTZ   NOT NULL DEFAULT now(),
    updated_at       TIMESTAMPTZ   NOT NULL DEFAULT now()
);

-- ============================================================================
-- dashboard 上下文（CQRS 读模型，全部可重算）
-- ============================================================================

CREATE TABLE automation_dashboard_summary_snapshots (
    snapshot_key                 VARCHAR(32) PRIMARY KEY,
    active_jobs_count            INTEGER     NOT NULL DEFAULT 0,
    total_executions_count       INTEGER     NOT NULL DEFAULT 0,
    running_executions_count     INTEGER     NOT NULL DEFAULT 0,
    success_cases_7d             INTEGER     NOT NULL DEFAULT 0,
    failure_cases_7d             INTEGER     NOT NULL DEFAULT 0,
    skipped_cases_7d             INTEGER     NOT NULL DEFAULT 0,
    pass_rate_7d                 DOUBLE PRECISION NOT NULL DEFAULT 0,
    avg_execution_duration_7d    DOUBLE PRECISION NULL,
    created_at                   TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at                   TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE automation_dashboard_daily_trends (
    stat_date        DATE     PRIMARY KEY,
    execution_total  INTEGER  NOT NULL DEFAULT 0,
    success_cases    INTEGER  NOT NULL DEFAULT 0,
    failure_cases    INTEGER  NOT NULL DEFAULT 0,
    skipped_cases    INTEGER  NOT NULL DEFAULT 0,
    running_cases    INTEGER  NOT NULL DEFAULT 0,
    created_at       TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at       TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE automation_dashboard_execution_snapshots (
    execution_id           INTEGER      PRIMARY KEY,
    job_id                 INTEGER      NOT NULL,
    job_name               VARCHAR(255) NOT NULL,
    job_url                VARCHAR(255) NOT NULL,
    status                 VARCHAR(32)  NOT NULL,
    -- 数据库列名是 start_time，API DTO 字段是 started_at（缺陷 D-14，Repository 层映射）
    start_time             TIMESTAMPTZ  NOT NULL,
    duration               DOUBLE PRECISION NULL,
    pre_cases_count        INTEGER      NOT NULL DEFAULT 0,
    completed_cases_count  INTEGER      NOT NULL DEFAULT 0,
    success_cases_count    INTEGER      NOT NULL DEFAULT 0,
    failure_cases_count    INTEGER      NOT NULL DEFAULT 0,
    skipped_cases_count    INTEGER      NOT NULL DEFAULT 0,
    running_cases_count    INTEGER      NOT NULL DEFAULT 0,
    progress_percent       DOUBLE PRECISION NOT NULL DEFAULT 0,
    running_case_names     JSONB        NULL,
    created_at             TIMESTAMPTZ  NOT NULL DEFAULT now(),
    updated_at             TIMESTAMPTZ  NOT NULL DEFAULT now()
);

CREATE INDEX idx_automation_dashboard_execution_snapshots_status
    ON automation_dashboard_execution_snapshots (status, start_time);

CREATE TABLE automation_dashboard_execution_item_snapshots (
    id              SERIAL       PRIMARY KEY,
    execution_id    INTEGER      NOT NULL,
    -- 语义是 execution_items.id；API DTO 字段名为 item_id（缺陷 D-15，Repository 层映射）
    case_id         INTEGER      NOT NULL,
    case_key        VARCHAR(512) NOT NULL,
    case_name       VARCHAR(255) NOT NULL,
    attempt_number  SMALLINT     NOT NULL DEFAULT 1,
    status          VARCHAR(32)  NOT NULL,
    start_time      TIMESTAMPTZ  NULL,
    end_time        TIMESTAMPTZ  NULL,
    duration        DOUBLE PRECISION NULL,
    error_message   TEXT         NULL,
    created_at      TIMESTAMPTZ  NOT NULL DEFAULT now(),
    updated_at      TIMESTAMPTZ  NOT NULL DEFAULT now(),
    CONSTRAINT uq_automation_dashboard_execution_item_snapshots_case
        UNIQUE (execution_id, case_key)
);

CREATE INDEX idx_automation_dashboard_execution_item_snapshots_execution
    ON automation_dashboard_execution_item_snapshots (execution_id, start_time);

-- ============================================================================
-- system 上下文
-- ============================================================================

CREATE TABLE sys_dict (
    id           SERIAL       PRIMARY KEY,
    dict_type    VARCHAR(64)  NOT NULL,
    code         SMALLINT     NOT NULL,
    label        VARCHAR(128) NOT NULL,
    dict_name    VARCHAR(128) NOT NULL,
    description  TEXT         NULL,
    category     VARCHAR(64)  NULL,
    sort_order   INTEGER      NOT NULL DEFAULT 0,
    color        VARCHAR(32)  NULL,
    status       VARCHAR(32)  NOT NULL DEFAULT 'active',
    created_at   TIMESTAMPTZ  NOT NULL DEFAULT now(),
    updated_at   TIMESTAMPTZ  NOT NULL DEFAULT now(),
    CONSTRAINT uq_sys_dict_type_code UNIQUE (dict_type, code)
);

CREATE INDEX idx_sys_dict_type ON sys_dict (dict_type);
