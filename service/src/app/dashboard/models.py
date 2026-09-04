from sqlalchemy import (
    String,
    Text,
    SmallInteger,
    TIMESTAMP,
    ForeignKey,
    UniqueConstraint,
    Index,
    Integer,
    Float,
    Date,
)
from sqlalchemy.dialects.postgresql import JSONB
from sqlalchemy.orm import mapped_column

from app.models import Base


class AutomationDashboardSummarySnapshots(Base):
    __tablename__ = "automation_dashboard_summary_snapshots"

    snapshot_key = mapped_column(String(32), primary_key=True)
    active_jobs_count = mapped_column(Integer, nullable=False, default=0)
    total_executions_count = mapped_column(Integer, nullable=False, default=0)
    running_executions_count = mapped_column(Integer, nullable=False, default=0)
    success_cases_7d = mapped_column(Integer, nullable=False, default=0)
    failure_cases_7d = mapped_column(Integer, nullable=False, default=0)
    skipped_cases_7d = mapped_column(Integer, nullable=False, default=0)
    pass_rate_7d = mapped_column(Float, nullable=False, default=0)
    avg_execution_duration_7d = mapped_column(Float, nullable=True)


class AutomationDashboardDailyTrends(Base):
    __tablename__ = "automation_dashboard_daily_trends"

    stat_date = mapped_column(Date, primary_key=True)
    execution_total = mapped_column(Integer, nullable=False, default=0)
    success_cases = mapped_column(Integer, nullable=False, default=0)
    failure_cases = mapped_column(Integer, nullable=False, default=0)
    skipped_cases = mapped_column(Integer, nullable=False, default=0)
    running_cases = mapped_column(Integer, nullable=False, default=0)


class AutomationDashboardExecutionSnapshots(Base):
    __tablename__ = "automation_dashboard_execution_snapshots"

    execution_id = mapped_column(
        Integer,
        primary_key=True,
    )
    job_id = mapped_column(Integer, nullable=False)
    job_name = mapped_column(String(255), nullable=False)
    job_url = mapped_column(String(255), nullable=False)
    status = mapped_column(
        String(32),
        nullable=False,
    )
    start_time = mapped_column(TIMESTAMP(timezone=True), nullable=False)
    duration = mapped_column(Float, nullable=True)
    pre_cases_count = mapped_column(Integer, nullable=False, default=0)
    completed_cases_count = mapped_column(Integer, nullable=False, default=0)
    success_cases_count = mapped_column(Integer, nullable=False, default=0)
    failure_cases_count = mapped_column(Integer, nullable=False, default=0)
    skipped_cases_count = mapped_column(Integer, nullable=False, default=0)
    running_cases_count = mapped_column(Integer, nullable=False, default=0)
    progress_percent = mapped_column(Float, nullable=False, default=0)
    running_case_names = mapped_column(JSONB, nullable=True)

    __table_args__ = (
        Index(
            "idx_automation_dashboard_execution_snapshots_status",
            "status",
            "start_time",
        ),
    )


class AutomationDashboardExecutionItemSnapshots(Base):
    __tablename__ = "automation_dashboard_execution_item_snapshots"

    id = mapped_column(Integer, primary_key=True, autoincrement=True)
    execution_id = mapped_column(
        Integer,
        nullable=False,
    )
    case_id = mapped_column(
        Integer,
        nullable=False,
    )
    case_key = mapped_column(String(512), nullable=False)
    case_name = mapped_column(String(255), nullable=False)
    attempt_number = mapped_column(SmallInteger, nullable=False, default=1)
    status = mapped_column(
        String(32),
        nullable=False,
    )
    start_time = mapped_column(TIMESTAMP(timezone=True), nullable=True)
    end_time = mapped_column(TIMESTAMP(timezone=True), nullable=True)
    duration = mapped_column(Float, nullable=True)
    error_message = mapped_column(Text, nullable=True)

    __table_args__ = (
        UniqueConstraint(
            "execution_id",
            "case_key",
            name="uq_automation_dashboard_execution_item_snapshots_case",
        ),
        Index(
            "idx_automation_dashboard_execution_item_snapshots_execution",
            "execution_id",
            "start_time",
        ),
    )
