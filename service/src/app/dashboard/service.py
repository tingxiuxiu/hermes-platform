from __future__ import annotations

from collections import defaultdict
from collections.abc import Iterable
from datetime import UTC, date, datetime, timedelta
from typing import Any

from sqlalchemy import case, delete, func, select
from sqlalchemy.ext.asyncio import AsyncSession

from app.celery_app.tasks.automation_dashboard import ensure_dashboard_snapshot_fresh
from app.config import settings
from .constants import (
    DashboardOverviewData,
    DashboardOverviewResponse,
    DashboardRunningCaseItem,
    DashboardRunningCaseListData,
    DashboardRunningCaseListResponse,
    DashboardRunningExecutionItem,
    DashboardSummaryItem,
    DashboardTrendItem,
)
from .models import (
    AutomationDashboardDailyTrends,
    AutomationDashboardExecutionCaseSnapshots,
    AutomationDashboardExecutionSnapshots,
    AutomationDashboardSummarySnapshots,
    JenkinsPipelines,
    JobExecutions,
    ExecutionCases,
)

SUMMARY_KEY = "global"


class AutomationDashboardService:
    @staticmethod
    def _build_job_filters(job_id: int | None, job_name: str | None) -> list[Any]:
        filters: list[Any] = []
        if job_id is not None:
            filters.append(JobExecutions.job_id == job_id)
        if job_name:
            filters.append(JobExecutions.job_name.ilike(f"%{job_name.strip()}%"))
        return filters

    @staticmethod
    def _build_pipeline_filters(job_id: int | None, job_name: str | None) -> list[Any]:
        filters: list[Any] = []
        if job_id is not None:
            filters.append(JenkinsPipelines.id == job_id)
        if job_name:
            filters.append(JenkinsPipelines.job_name.ilike(f"%{job_name.strip()}%"))
        return filters

    @staticmethod
    def _window_start(days: int) -> datetime:
        start_date = datetime.now(UTC).date() - timedelta(days=days - 1)
        return datetime.combine(start_date, datetime.min.time(), tzinfo=UTC)

    @staticmethod
    def _build_running_execution_item(
        execution: JobExecutions,
        latest_items: list[ExecutionCases],
    ) -> DashboardRunningExecutionItem:
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

        return DashboardRunningExecutionItem(
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
            updated_at=execution.updated_at,
        )

    @staticmethod
    async def _get_filtered_overview(
        session: AsyncSession,
        *,
        days: int,
        job_id: int | None,
        job_name: str | None,
    ) -> DashboardOverviewResponse:
        window_start = AutomationDashboardService._window_start(days)
        execution_filters = AutomationDashboardService._build_job_filters(
            job_id, job_name
        )
        pipeline_filters = AutomationDashboardService._build_pipeline_filters(
            job_id, job_name
        )

        active_jobs_count = (
            await session.execute(
                select(func.count())
                .select_from(JenkinsPipelines)
                .where(
                    JenkinsPipelines.status == "active",
                    *pipeline_filters,
                )
            )
        ).scalar_one() or 0
        total_executions_count = (
            await session.execute(
                select(func.count())
                .select_from(JobExecutions)
                .where(*execution_filters)
            )
        ).scalar_one() or 0
        running_executions_count = (
            await session.execute(
                select(func.count())
                .select_from(JobExecutions)
                .where(JobExecutions.status == "running", *execution_filters)
            )
        ).scalar_one() or 0

        latest_cases_stmt = (
            select(
                func.coalesce(
                    func.sum(case((ExecutionCases.status == "success", 1), else_=0)),
                    0,
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
                    func.sum(case((ExecutionCases.status == "skipped", 1), else_=0)),
                    0,
                ),
            )
            .select_from(ExecutionCases)
            .join(JobExecutions, JobExecutions.id == ExecutionCases.execution_id)
            .where(
                ExecutionCases.is_latest.is_(True),
                ExecutionCases.start_at >= window_start,
                *execution_filters,
            )
        )
        success_cases, failure_cases, skipped_cases = (
            await session.execute(latest_cases_stmt)
        ).one()

        avg_execution_duration = (
            await session.execute(
                select(func.avg(JobExecutions.duration)).where(
                    JobExecutions.start_at >= window_start,
                    JobExecutions.duration.is_not(None),
                    JobExecutions.status.in_(["completed", "failed"]),
                    *execution_filters,
                )
            )
        ).scalar_one()

        finished_total = success_cases + failure_cases + skipped_cases
        pass_rate = (
            round((success_cases / finished_total) * 100, 2)
            if finished_total > 0
            else 0.0
        )

        execution_day = func.date(JobExecutions.start_at)
        execution_rows = (
            await session.execute(
                select(
                    execution_day.label("stat_date"),
                    func.count(JobExecutions.id).label("execution_total"),
                )
                .where(JobExecutions.start_at >= window_start, *execution_filters)
                .group_by(execution_day)
            )
        ).all()
        execution_map = {row.stat_date: row.execution_total for row in execution_rows}

        case_day = func.date(ExecutionCases.start_at)
        case_rows = (
            await session.execute(
                select(
                    case_day.label("stat_date"),
                    func.coalesce(
                        func.sum(
                            case((ExecutionCases.status == "success", 1), else_=0)
                        ),
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
                        func.sum(
                            case((ExecutionCases.status == "skipped", 1), else_=0)
                        ),
                        0,
                    ).label("skipped_cases"),
                    func.coalesce(
                        func.sum(
                            case((ExecutionCases.status == "running", 1), else_=0)
                        ),
                        0,
                    ).label("running_cases"),
                )
                .select_from(ExecutionCases)
                .join(JobExecutions, JobExecutions.id == ExecutionCases.execution_id)
                .where(
                    ExecutionCases.is_latest.is_(True),
                    ExecutionCases.start_at >= window_start,
                    *execution_filters,
                )
                .group_by(case_day)
            )
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

        start_date = window_start.date()
        trends = [
            DashboardTrendItem(
                stat_date=start_date + timedelta(days=index),
                execution_total=execution_map.get(
                    start_date + timedelta(days=index), 0
                ),
                success_cases=case_map.get(start_date + timedelta(days=index), {}).get(
                    "success_cases", 0
                ),
                failure_cases=case_map.get(start_date + timedelta(days=index), {}).get(
                    "failure_cases", 0
                ),
                skipped_cases=case_map.get(start_date + timedelta(days=index), {}).get(
                    "skipped_cases", 0
                ),
                running_cases=case_map.get(start_date + timedelta(days=index), {}).get(
                    "running_cases", 0
                ),
            )
            for index in range(days)
        ]

        running_executions = (
            (
                await session.execute(
                    select(JobExecutions)
                    .where(JobExecutions.status == "running", *execution_filters)
                    .order_by(JobExecutions.start_at.desc())
                )
            )
            .scalars()
            .all()
        )
        running_execution_ids = [execution.id for execution in running_executions]
        running_items_map: dict[int, list[ExecutionCases]] = defaultdict(list)
        if running_execution_ids:
            running_items = (
                (
                    await session.execute(
                        select(ExecutionCases)
                        .where(
                            ExecutionCases.execution_id.in_(running_execution_ids),
                            ExecutionCases.is_latest.is_(True),
                        )
                        .order_by(
                            ExecutionCases.start_at.desc().nullslast(),
                            ExecutionCases.id.desc(),
                        )
                    )
                )
                .scalars()
                .all()
            )
            for item in running_items:
                running_items_map[item.execution_id].append(item)

        now = datetime.now(UTC)
        data = DashboardOverviewData(
            summary=DashboardSummaryItem(
                active_jobs_count=active_jobs_count,
                total_executions_count=total_executions_count,
                running_executions_count=running_executions_count,
                success_cases_7d=success_cases,
                failure_cases_7d=failure_cases,
                skipped_cases_7d=skipped_cases,
                pass_rate_7d=pass_rate,
                avg_execution_duration_7d=avg_execution_duration,
                updated_at=now,
            ),
            trends=trends,
            running_executions=[
                AutomationDashboardService._build_running_execution_item(
                    execution, running_items_map.get(execution.id, [])
                )
                for execution in running_executions
            ],
        )
        return DashboardOverviewResponse(
            code=200,
            message="获取 dashboard 概览成功",
            data=data,
        )

    @staticmethod
    async def _source_has_newer_dashboard_data(
        session: AsyncSession, updated_at: datetime
    ) -> bool:
        latest_execution_update = (
            await session.execute(select(func.max(JobExecutions.updated_at)))
        ).scalar_one()
        latest_case_update = (
            await session.execute(select(func.max(ExecutionCases.updated_at)))
        ).scalar_one()

        candidates = [
            value for value in [latest_execution_update, latest_case_update] if value
        ]
        return any(value > updated_at for value in candidates)

    @staticmethod
    async def _execution_source_is_newer(
        session: AsyncSession, execution_id: int, updated_at: datetime
    ) -> bool:
        latest_execution_update = (
            await session.execute(
                select(JobExecutions.updated_at).where(JobExecutions.id == execution_id)
            )
        ).scalar_one_or_none()
        latest_case_update = (
            await session.execute(
                select(func.max(ExecutionCases.updated_at)).where(
                    ExecutionCases.execution_id == execution_id
                )
            )
        ).scalar_one()

        candidates = [
            value for value in [latest_execution_update, latest_case_update] if value
        ]
        return any(value > updated_at for value in candidates)

    @staticmethod
    async def ensure_fresh_data(execution_id: int | None = None) -> None:
        await ensure_dashboard_snapshot_fresh(execution_id)

    @staticmethod
    async def get_overview(session: AsyncSession) -> DashboardOverviewResponse:
        return await AutomationDashboardService.get_overview(
            session=session,
            days=settings.AUTOMATION_DASHBOARD_TREND_DAYS,
            job_id=None,
            job_name=None,
        )

    @staticmethod
    async def get_overview(
        session: AsyncSession,
        *,
        days: int,
        job_id: int | None,
        job_name: str | None,
    ) -> DashboardOverviewResponse:
        if (
            days != settings.AUTOMATION_DASHBOARD_TREND_DAYS
            or job_id is not None
            or job_name
        ):
            return await AutomationDashboardService._get_filtered_overview(
                session,
                days=days,
                job_id=job_id,
                job_name=job_name,
            )

        summary_snapshot = await session.get(
            AutomationDashboardSummarySnapshots, SUMMARY_KEY
        )
        needs_refresh = summary_snapshot is None
        if summary_snapshot is not None:
            now = datetime.now(UTC)
            needs_refresh = (
                now - summary_snapshot.updated_at
            ).total_seconds() > settings.AUTOMATION_DASHBOARD_REFRESH_SECONDS
            if not needs_refresh:
                needs_refresh = (
                    await AutomationDashboardService._source_has_newer_dashboard_data(
                        session, summary_snapshot.updated_at
                    )
                )

        if needs_refresh:
            await AutomationDashboardService.ensure_fresh_data()

        summary_snapshot = await session.get(
            AutomationDashboardSummarySnapshots, SUMMARY_KEY
        )
        if summary_snapshot is None:
            await AutomationDashboardService.ensure_fresh_data()
            summary_snapshot = await session.get(
                AutomationDashboardSummarySnapshots, SUMMARY_KEY
            )
        assert summary_snapshot is not None

        trends = (
            (
                await session.execute(
                    select(AutomationDashboardDailyTrends)
                    .order_by(AutomationDashboardDailyTrends.stat_date.asc())
                    .limit(settings.AUTOMATION_DASHBOARD_TREND_DAYS)
                )
            )
            .scalars()
            .all()
        )
        running_snapshots = (
            (
                await session.execute(
                    select(AutomationDashboardExecutionSnapshots)
                    .where(AutomationDashboardExecutionSnapshots.status == "running")
                    .order_by(AutomationDashboardExecutionSnapshots.started_at.desc())
                )
            )
            .scalars()
            .all()
        )

        data = DashboardOverviewData(
            summary=DashboardSummaryItem(
                active_jobs_count=summary_snapshot.active_jobs_count,
                total_executions_count=summary_snapshot.total_executions_count,
                running_executions_count=summary_snapshot.running_executions_count,
                success_cases_7d=summary_snapshot.success_cases_7d,
                failure_cases_7d=summary_snapshot.failure_cases_7d,
                skipped_cases_7d=summary_snapshot.skipped_cases_7d,
                pass_rate_7d=summary_snapshot.pass_rate_7d,
                avg_execution_duration_7d=summary_snapshot.avg_execution_duration_7d,
                updated_at=summary_snapshot.updated_at,
            ),
            trends=[
                DashboardTrendItem(
                    stat_date=trend.stat_date,
                    execution_total=trend.execution_total,
                    success_cases=trend.success_cases,
                    failure_cases=trend.failure_cases,
                    skipped_cases=trend.skipped_cases,
                    running_cases=trend.running_cases,
                )
                for trend in trends
            ],
            running_executions=[
                DashboardRunningExecutionItem(
                    execution_id=snapshot.execution_id,
                    job_id=snapshot.job_id,
                    job_name=snapshot.job_name,
                    job_url=snapshot.job_url,
                    status=snapshot.status,
                    started_at=snapshot.started_at,
                    duration=snapshot.duration,
                    pre_cases_count=snapshot.pre_cases_count,
                    completed_cases_count=snapshot.completed_cases_count,
                    success_cases_count=snapshot.success_cases_count,
                    failure_cases_count=snapshot.failure_cases_count,
                    skipped_cases_count=snapshot.skipped_cases_count,
                    running_cases_count=snapshot.running_cases_count,
                    progress_percent=snapshot.progress_percent,
                    running_case_names=snapshot.running_case_names or [],
                    updated_at=snapshot.updated_at,
                )
                for snapshot in running_snapshots
            ],
        )
        return DashboardOverviewResponse(
            code=200, message="获取 dashboard 概览成功", data=data
        )

    @staticmethod
    async def get_running_execution_cases(
        session: AsyncSession, execution_id: int
    ) -> DashboardRunningCaseListResponse:
        snapshot = await session.get(
            AutomationDashboardExecutionSnapshots, execution_id
        )
        if (
            snapshot is None
            or (datetime.now(UTC) - snapshot.updated_at).total_seconds()
            > settings.AUTOMATION_DASHBOARD_REFRESH_SECONDS
        ):
            await AutomationDashboardService.ensure_fresh_data(execution_id)
        elif await AutomationDashboardService._execution_source_is_newer(
            session, execution_id, snapshot.updated_at
        ):
            await AutomationDashboardService.ensure_fresh_data(execution_id)

        rows = (
            (
                await session.execute(
                    select(AutomationDashboardExecutionCaseSnapshots)
                    .where(
                        AutomationDashboardExecutionCaseSnapshots.execution_id
                        == execution_id
                    )
                    .order_by(
                        AutomationDashboardExecutionCaseSnapshots.start_at.desc().nullslast(),
                        AutomationDashboardExecutionCaseSnapshots.item_id.desc(),
                    )
                )
            )
            .scalars()
            .all()
        )
        return DashboardRunningCaseListResponse(
            code=200,
            message="获取运行中 execution 的 case 快照成功",
            data=DashboardRunningCaseListData(
                execution_id=execution_id,
                items=[
                    DashboardRunningCaseItem(
                        execution_id=row.execution_id,
                        item_id=row.item_id,
                        case_key=row.case_key,
                        case_name=row.case_name,
                        attempt_number=row.attempt_number,
                        status=row.status,
                        start_at=row.start_at,
                        end_at=row.end_at,
                        duration=row.duration,
                        error_message=row.error_message,
                        updated_at=row.updated_at,
                    )
                    for row in rows
                ],
            ),
        )
