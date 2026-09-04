from __future__ import annotations

from datetime import UTC, datetime, timedelta
from typing import Any

from celery.exceptions import CeleryError
from sqlalchemy import case, delete, func, select
from sqlalchemy.dialects.postgresql import insert as pg_insert

from app.config import settings
from app.celery_app.celery_main import celery_app
from app.infrastructure.database.hdp.sync_session import SyncSessionLocal
from app.models.hdp.automation import (
    AutomationDashboardDailyTrends,
    AutomationDashboardExecutionCaseSnapshots,
    AutomationDashboardExecutionSnapshots,
    AutomationDashboardSummarySnapshots,
    JenkinsPipelines,
    JobExecutions,
    ExecutionCases,
)

SUMMARY_KEY = "global"


class AutomationDashboardProjector:
    @staticmethod
    def refresh_all() -> None:
        with SyncSessionLocal() as session:
            AutomationDashboardProjector._refresh_global_summary(session)
            AutomationDashboardProjector._refresh_daily_trends(
                session, days=settings.AUTOMATION_DASHBOARD_TREND_DAYS
            )
            running_execution_ids = (
                session.execute(
                    select(JobExecutions.id).where(JobExecutions.status == "running")
                )
                .scalars()
                .all()
            )
            for execution_id in running_execution_ids:
                AutomationDashboardProjector._refresh_execution_snapshot(
                    session, execution_id
                )
            session.commit()

    @staticmethod
    def refresh_execution(execution_id: int) -> None:
        with SyncSessionLocal() as session:
            AutomationDashboardProjector._refresh_execution_snapshot(
                session, execution_id
            )
            AutomationDashboardProjector._refresh_global_summary(session)
            AutomationDashboardProjector._refresh_daily_trends(
                session, days=settings.AUTOMATION_DASHBOARD_TREND_DAYS
            )
            session.commit()

    @staticmethod
    def _refresh_global_summary(session: Any) -> None:
        window_start = datetime.now(UTC) - timedelta(days=7)

        active_jobs_count = (
            session.execute(
                select(func.count())
                .select_from(JenkinsPipelines)
                .where(JenkinsPipelines.status == "active")
            ).scalar_one()
            or 0
        )
        total_executions_count = (
            session.execute(
                select(func.count()).select_from(JobExecutions)
            ).scalar_one()
            or 0
        )
        running_executions_count = (
            session.execute(
                select(func.count())
                .select_from(JobExecutions)
                .where(JobExecutions.status == "running")
            ).scalar_one()
            or 0
        )

        latest_cases_stmt = select(
            func.coalesce(
                func.sum(case((ExecutionCases.status == "success", 1), else_=0)), 0
            ),
            func.coalesce(
                func.sum(
                    case(
                        (
                            ExecutionCases.status.in_(["failure", "error"]),
                            1,
                        ),
                        else_=0,
                    )
                ),
                0,
            ),
            func.coalesce(
                func.sum(case((ExecutionCases.status == "skipped", 1), else_=0)), 0
            ),
        ).where(
            ExecutionCases.is_latest.is_(True),
            ExecutionCases.start_time >= window_start,
        )
        success_cases_7d, failure_cases_7d, skipped_cases_7d = session.execute(
            latest_cases_stmt
        ).one()

        duration_stmt = select(func.avg(JobExecutions.duration)).where(
            JobExecutions.start_time >= window_start,
            JobExecutions.duration.is_not(None),
            JobExecutions.status.in_(["completed", "failed"]),
        )
        avg_execution_duration_7d = session.execute(duration_stmt).scalar_one()

        finished_total = success_cases_7d + failure_cases_7d + skipped_cases_7d
        pass_rate_7d = (
            round((success_cases_7d / finished_total) * 100, 2)
            if finished_total > 0
            else 0.0
        )

        stmt = pg_insert(AutomationDashboardSummarySnapshots).values(
            snapshot_key=SUMMARY_KEY,
            active_jobs_count=active_jobs_count,
            total_executions_count=total_executions_count,
            running_executions_count=running_executions_count,
            success_cases_7d=success_cases_7d,
            failure_cases_7d=failure_cases_7d,
            skipped_cases_7d=skipped_cases_7d,
            pass_rate_7d=pass_rate_7d,
            avg_execution_duration_7d=avg_execution_duration_7d,
        )
        session.execute(
            stmt.on_conflict_do_update(
                index_elements=[AutomationDashboardSummarySnapshots.snapshot_key],
                set_={
                    "active_jobs_count": stmt.excluded.active_jobs_count,
                    "total_executions_count": stmt.excluded.total_executions_count,
                    "running_executions_count": stmt.excluded.running_executions_count,
                    "success_cases_7d": stmt.excluded.success_cases_7d,
                    "failure_cases_7d": stmt.excluded.failure_cases_7d,
                    "skipped_cases_7d": stmt.excluded.skipped_cases_7d,
                    "pass_rate_7d": stmt.excluded.pass_rate_7d,
                    "avg_execution_duration_7d": stmt.excluded.avg_execution_duration_7d,
                    "updated_at": func.now(),
                },
            )
        )

    @staticmethod
    def _refresh_daily_trends(session: Any, days: int) -> None:
        today = datetime.now(UTC).date()
        start_date = today - timedelta(days=days - 1)
        execution_day = func.date(JobExecutions.start_time)
        case_day = func.date(ExecutionCases.start_time)

        execution_rows = session.execute(
            select(
                execution_day.label("stat_date"),
                func.count(JobExecutions.id).label("execution_total"),
            )
            .where(JobExecutions.start_time >= start_date)
            .group_by(execution_day)
        ).all()
        execution_map = {row.stat_date: row.execution_total for row in execution_rows}

        case_rows = session.execute(
            select(
                case_day.label("stat_date"),
                func.coalesce(
                    func.sum(case((ExecutionCases.status == "success", 1), else_=0)),
                    0,
                ).label("success_cases"),
                func.coalesce(
                    func.sum(
                        case(
                            (
                                ExecutionCases.status.in_(["failure", "error"]),
                                1,
                            ),
                            else_=0,
                        )
                    ),
                    0,
                ).label("failure_cases"),
                func.coalesce(
                    func.sum(case((ExecutionCases.status == "skipped", 1), else_=0)),
                    0,
                ).label("skipped_cases"),
                func.coalesce(
                    func.sum(case((ExecutionCases.status == "running", 1), else_=0)),
                    0,
                ).label("running_cases"),
            )
            .where(
                ExecutionCases.is_latest.is_(True),
                ExecutionCases.start_time >= start_date,
            )
            .group_by(case_day)
        ).all()
        case_map = {
            row.stat_date: {
                "success_cases": row.success_cases,
                "failure_cases": row.failure_cases,
                "skipped_cases": row.skipped_cases,
                "running_cases": row.running_cases,
            }
            for row in case_rows
        }

        for index in range(days):
            stat_date = start_date + timedelta(days=index)
            case_values = case_map.get(stat_date, {})
            stmt = pg_insert(AutomationDashboardDailyTrends).values(
                stat_date=stat_date,
                execution_total=execution_map.get(stat_date, 0),
                success_cases=case_values.get("success_cases", 0),
                failure_cases=case_values.get("failure_cases", 0),
                skipped_cases=case_values.get("skipped_cases", 0),
                running_cases=case_values.get("running_cases", 0),
            )
            session.execute(
                stmt.on_conflict_do_update(
                    index_elements=[AutomationDashboardDailyTrends.stat_date],
                    set_={
                        "execution_total": stmt.excluded.execution_total,
                        "success_cases": stmt.excluded.success_cases,
                        "failure_cases": stmt.excluded.failure_cases,
                        "skipped_cases": stmt.excluded.skipped_cases,
                        "running_cases": stmt.excluded.running_cases,
                        "updated_at": func.now(),
                    },
                )
            )

    @staticmethod
    def _refresh_execution_snapshot(session: Any, execution_id: int) -> None:
        execution = session.get(JobExecutions, execution_id)
        if not execution:
            session.execute(
                delete(AutomationDashboardExecutionCaseSnapshots).where(
                    AutomationDashboardExecutionCaseSnapshots.execution_id
                    == execution_id
                )
            )
            session.execute(
                delete(AutomationDashboardExecutionSnapshots).where(
                    AutomationDashboardExecutionSnapshots.execution_id == execution_id
                )
            )
            return

        latest_items = (
            session.execute(
                select(ExecutionCases)
                .where(
                    ExecutionCases.execution_id == execution_id,
                    ExecutionCases.is_latest.is_(True),
                )
                .order_by(
                    ExecutionCases.start_at.desc().nullslast(),
                    ExecutionCases.id.desc(),
                )
            )
            .scalars()
            .all()
        )

        success_cases_count = sum(
            1 for item in latest_items if item.status == "success"
        )
        failure_cases_count = sum(
            1 for item in latest_items if item.status in {"failure", "error"}
        )
        skipped_cases_count = sum(
            1 for item in latest_items if item.status == "skipped"
        )
        running_items = [item for item in latest_items if item.status == "running"]
        completed_cases_count = (
            success_cases_count + failure_cases_count + skipped_cases_count
        )
        pre_cases_count = execution.pre_cases_count or len(latest_items)
        progress_base = pre_cases_count or len(latest_items) or 1
        progress_percent = min(
            100.0, round((completed_cases_count / progress_base) * 100, 2)
        )
        duration = execution.duration
        if duration is None and execution.start_at is not None:
            duration = (datetime.now(UTC) - execution.start_at).total_seconds()

        snapshot_stmt = pg_insert(AutomationDashboardExecutionSnapshots).values(
            execution_id=execution.id,
            job_id=execution.job_id,
            job_name=execution.job_name,
            job_url=execution.job_url,
            status=execution.status,
            started_at=execution.start_at,
            duration=duration,
            pre_cases_count=pre_cases_count,
            completed_cases_count=completed_cases_count,
            success_cases_count=success_cases_count,
            failure_cases_count=failure_cases_count,
            skipped_cases_count=skipped_cases_count,
            running_cases_count=len(running_items),
            progress_percent=progress_percent,
            running_case_names=[item.case_name for item in running_items[:5]],
        )
        session.execute(
            snapshot_stmt.on_conflict_do_update(
                index_elements=[AutomationDashboardExecutionSnapshots.execution_id],
                set_={
                    "job_id": snapshot_stmt.excluded.job_id,
                    "job_name": snapshot_stmt.excluded.job_name,
                    "job_url": snapshot_stmt.excluded.job_url,
                    "status": snapshot_stmt.excluded.status,
                    "started_at": snapshot_stmt.excluded.started_at,
                    "duration": snapshot_stmt.excluded.duration,
                    "pre_cases_count": snapshot_stmt.excluded.pre_cases_count,
                    "completed_cases_count": snapshot_stmt.excluded.completed_cases_count,
                    "success_cases_count": snapshot_stmt.excluded.success_cases_count,
                    "failure_cases_count": snapshot_stmt.excluded.failure_cases_count,
                    "skipped_cases_count": snapshot_stmt.excluded.skipped_cases_count,
                    "running_cases_count": snapshot_stmt.excluded.running_cases_count,
                    "progress_percent": snapshot_stmt.excluded.progress_percent,
                    "running_case_names": snapshot_stmt.excluded.running_case_names,
                    "updated_at": func.now(),
                },
            )
        )

        active_case_keys = {item.case_key for item in latest_items}
        if active_case_keys:
            session.execute(
                delete(AutomationDashboardExecutionCaseSnapshots).where(
                    AutomationDashboardExecutionCaseSnapshots.execution_id
                    == execution_id,
                    AutomationDashboardExecutionCaseSnapshots.case_key.not_in(
                        active_case_keys
                    ),
                )
            )
        else:
            session.execute(
                delete(AutomationDashboardExecutionCaseSnapshots).where(
                    AutomationDashboardExecutionCaseSnapshots.execution_id
                    == execution_id
                )
            )

        for item in latest_items:
            case_stmt = pg_insert(AutomationDashboardExecutionCaseSnapshots).values(
                execution_id=execution_id,
                item_id=item.id,
                case_key=item.case_key,
                case_name=item.case_name,
                attempt_number=item.attempt_number,
                status=item.status,
                start_at=item.start_at,
                end_at=item.end_at,
                duration=item.duration,
                error_message=item.error_message,
            )
            session.execute(
                case_stmt.on_conflict_do_update(
                    constraint="uq_automation_dashboard_execution_case_snapshots_case",
                    set_={
                        "item_id": case_stmt.excluded.item_id,
                        "case_name": case_stmt.excluded.case_name,
                        "attempt_number": case_stmt.excluded.attempt_number,
                        "status": case_stmt.excluded.status,
                        "start_at": case_stmt.excluded.start_at,
                        "end_at": case_stmt.excluded.end_at,
                        "duration": case_stmt.excluded.duration,
                        "error_message": case_stmt.excluded.error_message,
                        "updated_at": func.now(),
                    },
                )
            )


@celery_app.task(name="automation.dashboard.refresh_all")
def refresh_automation_dashboard_all_task() -> None:
    AutomationDashboardProjector.refresh_all()


@celery_app.task(name="automation.dashboard.refresh_execution")
def refresh_automation_dashboard_for_execution_task(execution_id: int) -> None:
    AutomationDashboardProjector.refresh_execution(execution_id)


async def ensure_dashboard_snapshot_fresh(execution_id: int | None = None) -> None:
    from anyio import to_thread

    if execution_id is None:
        await to_thread.run_sync(AutomationDashboardProjector.refresh_all)
    else:
        await to_thread.run_sync(
            AutomationDashboardProjector.refresh_execution, execution_id
        )


def queue_dashboard_refresh(execution_id: int) -> None:
    try:
        refresh_automation_dashboard_for_execution_task.delay(execution_id)
    except (CeleryError, OSError, RuntimeError):
        # worker / broker 不可用时降级：读接口会在首次访问时同步补算快照
        return
