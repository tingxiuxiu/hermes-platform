from app.models import Base
from sqlalchemy import (
    TIMESTAMP,
    Integer,
    SmallInteger,
    String,
    Boolean,
    Text,
    Float,
    UniqueConstraint,
    Index,
    func,
    UUID,
)
from sqlalchemy.dialects.postgresql import JSONB, ARRAY
from sqlalchemy.orm import mapped_column


class JenkinsPipelines(Base):
    __tablename__ = "jenkins_pipelines"

    id = mapped_column(Integer, primary_key=True, autoincrement=True)
    job_name = mapped_column(String(255), nullable=False, unique=True)
    job_url = mapped_column(String(255), nullable=False)
    status = mapped_column(
        String(32),
        nullable=False,
        default="active",
    )
    last_build_number = mapped_column(Integer, nullable=True)
    last_build_uid = mapped_column(UUID(as_uuid=True), nullable=True, unique=True)
    last_build_status = mapped_column(
        String(32),
        nullable=True,
    )
    last_build_timestamp = mapped_column(TIMESTAMP(timezone=True), nullable=True)
    last_build_duration = mapped_column(Float, nullable=True)  # Duration in seconds
    pipeline_params = mapped_column(JSONB, nullable=True)
    sync_at = mapped_column(
        TIMESTAMP(timezone=True), server_default=func.now(), onupdate=func.now()
    )


class TestExecutions(Base):
    __tablename__ = "test_executions"

    id = mapped_column(Integer, primary_key=True, autoincrement=True)
    build_uid = mapped_column(UUID(as_uuid=True), nullable=False, unique=True)
    job_name = mapped_column(String(255), nullable=False)
    job_url = mapped_column(String(255), nullable=True, default=None)
    project_name = mapped_column(String(255), nullable=False)
    software_name = mapped_column(String(100), nullable=False)
    software_version = mapped_column(String(100), nullable=False)
    labels = mapped_column(ARRAY(String(64)), nullable=True)
    status = mapped_column(
        String(32),
        nullable=False,
        default="running",
    )
    start_time = mapped_column(TIMESTAMP(timezone=True), nullable=False)
    end_time = mapped_column(TIMESTAMP(timezone=True), nullable=True)
    duration = mapped_column(Float, nullable=True)  # Duration in seconds
    planned_cases_count = mapped_column(
        Integer, nullable=False, default=0, server_default="0"
    )
    pass_count = mapped_column(Integer, nullable=False, default=0, server_default="0")
    failure_count = mapped_column(
        Integer, nullable=False, default=0, server_default="0"
    )
    skipped_count = mapped_column(
        Integer, nullable=False, default=0, server_default="0"
    )
    pass_rate = mapped_column(Float, nullable=True)


class ExecutionItems(Base):

    __tablename__ = "execution_items"

    id = mapped_column(Integer, primary_key=True, autoincrement=True)
    build_uid = mapped_column(
        UUID(as_uuid=True),
        nullable=False,
    )
    # 用于标识 pytest 中的 id
    case_uid = mapped_column(UUID(as_uuid=True), nullable=False, unique=True)
    # 用于标识 JIRA等测试管理系统中的用例 ID
    case_key = mapped_column(String(128), nullable=False)

    case_name = mapped_column(Text, nullable=False)
    labels = mapped_column(
        ARRAY(String(64)), nullable=True
    )  # JIRA labels / pytest markers / 自定义标签等

    # 第几次执行：1=首次运行，2=第一次重试(rerun)，以此类推；重试不覆盖旧行
    attempt_number = mapped_column(
        SmallInteger, nullable=False, default=1, server_default="1"
    )
    # 是否是该 case 在本次 execution 中当前生效的最终结果（最后一次 attempt）
    is_latest = mapped_column(
        Boolean, nullable=False, default=True, server_default="true"
    )

    status = mapped_column(
        String(32),
        nullable=False,
        default="running",
    )
    start_time = mapped_column(TIMESTAMP(timezone=True), nullable=True)
    end_time = mapped_column(TIMESTAMP(timezone=True), nullable=True)
    duration = mapped_column(Float, nullable=True)  # Duration in seconds

    # 失败时的概要错误信息，用于列表页直接展示，避免为了显示概要而 join steps 表
    error_message = mapped_column(Text, nullable=True)
    error_traceback = mapped_column(Text, nullable=True)  # 失败时的完整堆栈/错误信息

    __table_args__ = (
        # 同一 execution 下，同一 case 的同一次 attempt 只会有一行；
        # pytest 事件重复上报同一 attempt 时可据此做幂等 upsert（ON CONFLICT DO UPDATE）
        UniqueConstraint(
            "build_uid",
            "case_key",
            "attempt_number",
            name="uq_execution_items_attempt",
        ),
    )


class ExecutionItemSteps(Base):
    __tablename__ = "execution_item_steps"

    id = mapped_column(Integer, primary_key=True, autoincrement=True)
    case_uid = mapped_column(
        UUID(as_uuid=True),
        nullable=False,
    )
    parent_step_id = mapped_column(
        Integer,
        nullable=True,
    )
    parent_step_index = mapped_column(Integer, nullable=True)

    step_index = mapped_column(Integer, nullable=False, default=0)
    step_path = mapped_column(String(255), nullable=False)  # 0, 0.1, 0.1.2
    depth = mapped_column(Integer, nullable=False, default=0)

    step_name = mapped_column(Text, nullable=False)
    status = mapped_column(String(32), nullable=False)
    start_time = mapped_column(TIMESTAMP(timezone=True), nullable=True)
    end_time = mapped_column(TIMESTAMP(timezone=True), nullable=True)
    duration = mapped_column(Float, nullable=True)

    __table_args__ = (
        UniqueConstraint("case_uid", "step_path", name="uq_execution_item_steps_path"),
        Index("idx_execution_item_steps_case_parent", "case_uid", "parent_step_id"),
        Index("idx_execution_item_steps_case_path", "case_uid", "step_path"),
    )


class ExecutionCaseStepAttachments(Base):
    __tablename__ = "execution_case_step_attachments"

    id = mapped_column(Integer, primary_key=True, autoincrement=True)
    step_id = mapped_column(
        Integer,
        nullable=False,
    )
    attachment_type = mapped_column(String(32), nullable=False)  # screenshot/log/file
    file_name = mapped_column(String(255), nullable=True)
    url = mapped_column(String(1024), nullable=False)
    mime_type = mapped_column(String(128), nullable=True)
