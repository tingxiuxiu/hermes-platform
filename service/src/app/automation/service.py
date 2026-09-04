from datetime import UTC, datetime
from typing import Optional

from sqlalchemy import func, select, update
from sqlalchemy.dialects.postgresql import insert as pg_insert
from sqlalchemy.ext.asyncio import AsyncSession
from sqlalchemy.orm import Query, selectinload

from .exceptions import (
    ExecutionNotFoundException,
    ExecutionItemNotFoundException,
    StepNotFoundException,
)
from .constants import (
    ExecutionRow,
    AttachmentItem,
    ExecutionItemRow,
    StepRow,
    ExecutionItemHistoryRow,
    PipelineRow,
    ExecutionListData,
    PipelineListData,
    ExecutionItemListData,
    ExecutionItemHistoryListData,
    CreateExecutionResponse,
    CreateExecutionErrorResponse,
    UpdateExecutionResponse,
    UpdateExecutionErrorResponse,
    QueryExecutionListResponse,
    QueryExecutionListErrorResponse,
    QueryPipelineListResponse,
    QueryPipelineListErrorResponse,
    CreateExecutionItemResponse,
    CreateExecutionItemErrorResponse,
    UpdateExecutionItemResponse,
    UpdateExecutionItemErrorResponse,
    QueryExecutionItemListResponse,
    QueryExecutionItemListErrorResponse,
    CreateStepResponse,
    CreateStepErrorResponse,
    QueryStepListResponse,
    QueryStepListErrorResponse,
    ExecutionItemHistoryListData,
    QueryExecutionItemHistoryResponse,
    QueryExecutionItemHistoryErrorResponse,
)
from .schemas import (
    QueryExecutionItemHistory,
    CreateExecution,
    QueryExecutionList,
    UpdateExecution,
    CreateExecutionItem,
    QueryExecutionItemList,
    UpdateExecutionItem,
    QueryPipelineList,
    CreateStep,
)
from .models import (
    JenkinsPipelines,
    TestExecutions,
    ExecutionItems,
    ExecutionItemSteps,
    ExecutionCaseStepAttachments,
)


def _to_execution_row(execution: TestExecutions) -> ExecutionRow:
    return ExecutionRow(
        id=execution.id,
        build_uid=str(execution.build_uid),
        job_name=execution.job_name,
        job_url=execution.job_url,
        project_name=execution.project_name,
        software_name=execution.software_name,
        software_version=execution.software_version,
        status=execution.status,
        start_time=execution.start_time,
        end_time=execution.end_time,
        duration=execution.duration,
        planned_cases_count=execution.planned_cases_count,
        pass_count=execution.pass_count,
        failure_count=execution.failure_count,
        skipped_count=execution.skipped_count,
        created_at=execution.created_at,
    )


def _to_pipeline_row(job: JenkinsPipelines) -> PipelineRow:
    return PipelineRow(
        id=job.id,
        job_name=job.job_name,
        job_url=job.job_url,
        status=job.status,
        last_build_number=job.last_build_number,
        last_build_uid=str(job.last_build_uid) if job.last_build_uid else None,
        last_build_status=job.last_build_status,
        last_build_timestamp=job.last_build_timestamp,
        last_build_duration=job.last_build_duration,
        pipeline_params=job.pipeline_params,
        sync_at=job.sync_at,
    )


def _to_execution_item_row(item: ExecutionItems) -> ExecutionItemRow:
    return ExecutionItemRow(
        id=item.id,
        build_uid=str(item.build_uid),
        case_key=item.case_key,
        case_uid=str(item.case_uid),
        case_name=item.case_name,
        attempt_number=item.attempt_number,
        is_latest=item.is_latest,
        status=item.status,
        start_time=item.start_time,
        end_time=item.end_time,
        duration=item.duration,
        error_message=item.error_message,
        error_traceback=item.error_traceback,
    )


def _to_step_row(step: ExecutionItemSteps) -> StepRow:
    return StepRow(
        id=step.id,
        case_uid=str(step.case_uid),
        step_index=step.step_index,
        parent_step_id=step.parent_step_id,
        parent_step_index=step.parent_step_index,
        step_path=step.step_path,
        step_name=step.step_name,
        depth=step.depth,
        status=step.status,
        start_time=step.start_time,
        end_time=step.end_time,
        duration=step.duration,
        attachments=[],
        sub_steps=[],  # This will be populated later if needed
    )


async def list_pipelines(
    session: AsyncSession, query: QueryPipelineList
) -> QueryPipelineListResponse:
    """获取自动化任务分页列表"""
    stmt = select(JenkinsPipelines)

    if query.job_name:
        stmt = stmt.where(JenkinsPipelines.job_name.ilike(f"%{query.job_name}%"))
    if query.status is not None:
        stmt = stmt.where(JenkinsPipelines.status == query.status)
    if query.last_build_status is not None:
        stmt = stmt.where(JenkinsPipelines.last_build_status == query.last_build_status)

    count_stmt = select(func.count()).select_from(stmt.subquery())
    total = (await session.execute(count_stmt)).scalar_one() or 0

    offset = (query.page - 1) * query.page_size
    pipelines = (
        (
            await session.execute(
                stmt.order_by(JenkinsPipelines.id.desc())
                .offset(offset)
                .limit(query.page_size)
            )
        )
        .scalars()
        .all()
    )

    return QueryPipelineListResponse(
        code=200,
        message="获取自动化任务列表成功",
        data=[_to_pipeline_row(pipeline) for pipeline in pipelines],
    )


async def create_execution(
    session: AsyncSession, data: CreateExecution
) -> CreateExecutionResponse:
    """创建一次 execution (pytest_sessionstart 上报)"""
    execution = TestExecutions(
        build_uid=data.build_uid,
        job_name=data.job_name,
        job_url=data.job_url,
        project_name=data.project_name,
        software_name=data.software_name,
        software_version=data.software_version,
        status="running",
        start_time=data.start_time,
        planned_cases_count=data.planned_cases_count,
    )
    session.add(execution)
    await session.flush()
    return CreateExecutionResponse(
        code=200, message="创建 execution 成功", data=_to_execution_row(execution)
    )


async def _get_execution_by_build_uid(
    session: AsyncSession, build_uid: str
) -> TestExecutions | None:
    return (
        await session.execute(
            select(TestExecutions).where(TestExecutions.build_uid == build_uid)
        )
    ).scalar_one_or_none()


async def update_execution(
    session: AsyncSession, build_uid: str, data: UpdateExecution
) -> UpdateExecutionResponse:
    """更新/结束一次 execution (pytest_sessionfinish 上报)"""
    execution = await _get_execution_by_build_uid(session, build_uid)
    if not execution:
        raise ExecutionNotFoundException()

    if data.status is not None:
        execution.status = data.status
    if data.end_time is not None:
        execution.end_time = data.end_time
    if data.duration is not None:
        execution.duration = data.duration

    await session.flush()
    return UpdateExecutionResponse(
        code=200, message="更新 execution 成功", data=_to_execution_row(execution)
    )


async def get_execution_row(session: AsyncSession, build_uid: str) -> ExecutionRow:
    """获取 execution 详情"""
    execution = await _get_execution_by_build_uid(session, build_uid)
    if not execution:
        raise ExecutionNotFoundException()

    return _to_execution_row(execution)


async def list_executions(
    session: AsyncSession, query: QueryExecutionList
) -> QueryExecutionListResponse:
    """获取 execution 分页列表"""
    stmt = select(TestExecutions)

    if query.build_uid is not None:
        stmt = stmt.where(TestExecutions.build_uid == query.build_uid)
    if query.job_name:
        stmt = stmt.where(TestExecutions.job_name.ilike(f"%{query.job_name}%"))
    if query.status is not None:
        stmt = stmt.where(TestExecutions.status == query.status)

    count_stmt = select(func.count()).select_from(stmt.subquery())
    total = (await session.execute(count_stmt)).scalar_one() or 0

    offset = (query.page - 1) * query.page_size
    stmt = stmt.order_by(TestExecutions.id.desc()).offset(offset).limit(query.page_size)
    executions = (await session.execute(stmt)).scalars().all()

    list_data = ExecutionListData(
        total=total,
        page=query.page,
        page_size=query.page_size,
        items=[_to_execution_row(e) for e in executions],
    )
    return QueryExecutionListResponse(
        code=200, message="获取 execution 列表成功", data=list_data
    )


async def create_execution_item(
    session: AsyncSession, build_uid: str, data: CreateExecutionItem
) -> CreateExecutionItemResponse:
    max_attempt = (
        await session.execute(
            select(func.max(ExecutionItems.attempt_number)).where(
                ExecutionItems.build_uid == build_uid,
                ExecutionItems.case_key == data.case_key,
            )
        )
    ).scalar() or 0
    attempt_number = max_attempt + 1
    execution_item = ExecutionItems(
        build_uid=build_uid,
        case_key=data.case_key,
        case_name=data.case_name,
        case_uid=data.case_uid,
        attempt_number=attempt_number,
        is_latest=True,
        status="running",
        start_time=data.start_time or datetime.now(UTC),
    )

    session.add(execution_item)
    await session.flush()

    return CreateExecutionItemResponse(
        code=200,
        message="上报用例执行开始成功",
        data=_to_execution_item_row(execution_item),
    )


async def update_execution_item(
    session: AsyncSession,
    build_uid: str,
    case_key: str,
    case_uid: str,
    data: UpdateExecutionItem,
) -> UpdateExecutionItemResponse:
    item = (
        await session.execute(
            select(ExecutionItems).where(
                ExecutionItems.build_uid == build_uid,
                ExecutionItems.case_key == case_key,
                ExecutionItems.case_uid == case_uid,
            )
        )
    ).scalar_one_or_none()
    if not item or str(item.build_uid) != build_uid:
        raise ExecutionItemNotFoundException()
    item.status = data.status
    if data.end_time is not None:
        item.end_time = data.end_time
    if data.duration is not None:
        item.duration = data.duration
    if data.error_message is not None:
        item.error_message = data.error_message

    await session.flush()
    return UpdateExecutionItemResponse(
        code=200, message="上报用例执行结束成功", data=_to_execution_item_row(item)
    )


async def list_execution_items(
    session: AsyncSession,
    build_uid: str,
    query: QueryExecutionItemList,
    only_latest: bool = True,
) -> QueryExecutionItemListResponse:
    stmt = select(ExecutionItems).where(ExecutionItems.build_uid == build_uid)
    if only_latest:
        stmt = stmt.where(ExecutionItems.is_latest.is_(True))
    if query.status is not None:
        stmt = stmt.where(ExecutionItems.status == query.status)
    if query.case_key is not None:
        stmt = stmt.where(ExecutionItems.case_key == query.case_key)
    if query.case_name is not None:
        stmt = stmt.where(ExecutionItems.case_name.ilike(f"%{query.case_name}%"))

    count_stmt = select(func.count()).select_from(stmt.subquery())
    total = (await session.execute(count_stmt)).scalar_one() or 0

    page_offset = (query.page - 1) * query.page_size
    stmt = (
        stmt.order_by(ExecutionItems.id.asc())
        .offset(page_offset)
        .limit(query.page_size)
    )
    items = (await session.execute(stmt)).scalars().all()

    list_data = ExecutionItemListData(
        total=total,
        page=query.page,
        page_size=query.page_size,
        rows=[_to_execution_item_row(i) for i in items],
    )
    return QueryExecutionItemListResponse(
        code=200, message="获取用例列表成功", data=list_data
    )


async def get_execution_item_attempts(
    session: AsyncSession, build_uid: str, case_key: str
) -> QueryExecutionItemListResponse:
    stmt = (
        select(ExecutionItems)
        .where(
            ExecutionItems.build_uid == build_uid,
            ExecutionItems.case_key == case_key,
            ExecutionItems.is_latest.is_(False),
        )
        .order_by(ExecutionItems.attempt_number.asc())
    )
    items = (await session.execute(stmt)).scalars().all()

    list_data = ExecutionItemListData(
        total=len(items),
        page=1,
        page_size=len(items) or 1,
        rows=[_to_execution_item_row(i) for i in items],
    )
    return QueryExecutionItemListResponse(
        code=200, message="获取用例重试历史成功", data=list_data
    )


async def get_case_detail(session: AsyncSession, item_id: int) -> QueryStepListResponse:
    """获取单次 attempt 详情 (含步骤与附件)"""
    item = await session.get(ExecutionItems, item_id)
    if not item:
        raise ExecutionItemNotFoundException()

    stmt = (
        select(ExecutionItemSteps)
        .where(ExecutionItemSteps.case_uid == item.case_uid)
        .order_by(ExecutionItemSteps.step_index.asc(), ExecutionItemSteps.id.asc())
    )
    steps = (await session.execute(stmt)).scalars().all()

    children_by_parent: dict[int | None, list[StepRow]] = {}
    step_rows: dict[int, StepRow] = {}

    for step in steps:
        row = StepRow(
            id=step.id,
            case_uid=str(step.case_uid),
            parent_step_id=step.parent_step_id,
            parent_step_index=step.parent_step_index,
            step_index=step.step_index,
            step_path=step.step_path,
            depth=step.depth,
            step_name=step.step_name,
            status=step.status,
            start_time=step.start_time or datetime.now(UTC),
            end_time=step.end_time,
            duration=step.duration,
            attachments=[],
            sub_steps=[],
        )
        step_rows[step.id] = row
        children_by_parent.setdefault(step.parent_step_id, []).append(row)

    root_steps: list[StepRow] = []
    for step in steps:
        row = step_rows[step.id]
        parent_id = step.parent_step_id
        if parent_id is None:
            root_steps.append(row)
            continue

        parent_row = step_rows.get(parent_id)
        if parent_row is not None:
            parent_row.sub_steps = parent_row.sub_steps or []
            parent_row.sub_steps.append(row)

    return QueryStepListResponse(
        code=200,
        message="获取用例执行详情成功",
        data=root_steps,
    )


async def create_steps(
    session: AsyncSession, item_id: int, data: CreateStep
) -> CreateStepResponse:
    """单步上报步骤，使用 step_path 作为幂等键，避免同一个 item 在高频上传下插错顺序或重复写入。"""
    item = await session.get(ExecutionItems, item_id)
    if not item:
        raise ExecutionItemNotFoundException()

    parent_step_id = None
    if data.parent_step_index is not None:
        parent_step = (
            await session.execute(
                select(ExecutionItemSteps).where(
                    ExecutionItemSteps.case_uid == item.case_uid,
                    ExecutionItemSteps.step_index == data.parent_step_index,
                )
            )
        ).scalar_one_or_none()
        if parent_step is not None:
            parent_step_id = parent_step.id

    step_index = data.step_index
    if step_index is None:
        existing_max = (
            await session.execute(
                select(
                    func.coalesce(func.max(ExecutionItemSteps.step_index), -1)
                ).where(ExecutionItemSteps.case_uid == item.case_uid)
            )
        ).scalar_one()
        step_index = existing_max + 1

    parent_step_index = data.parent_step_index
    if parent_step_id is not None:
        parent_step_index = (
            await session.execute(
                select(ExecutionItemSteps.step_index).where(
                    ExecutionItemSteps.id == parent_step_id
                )
            )
        ).scalar_one()

    depth = (
        0
        if parent_step_id is None
        else 1
        + (
            await session.execute(
                select(ExecutionItemSteps.depth).where(
                    ExecutionItemSteps.id == parent_step_id
                )
            )
        ).scalar_one()
    )

    payload = {
        "case_uid": item.case_uid,
        "parent_step_id": parent_step_id,
        "parent_step_index": parent_step_index,
        "step_index": step_index,
        "step_path": data.step_path,
        "depth": depth,
        "step_name": data.step_name,
        "status": data.status,
        "start_time": data.start_time or datetime.now(UTC),
        "end_time": data.end_time,
        "duration": data.duration,
    }

    stmt = pg_insert(ExecutionItemSteps).values(payload)
    stmt = stmt.on_conflict_do_update(
        constraint="uq_execution_item_steps_path",
        set_={
            "parent_step_id": stmt.excluded.parent_step_id,
            "parent_step_index": stmt.excluded.parent_step_index,
            "step_index": stmt.excluded.step_index,
            "depth": stmt.excluded.depth,
            "step_name": stmt.excluded.step_name,
            "status": stmt.excluded.status,
            "start_time": stmt.excluded.start_time,
            "end_time": stmt.excluded.end_time,
            "duration": stmt.excluded.duration,
        },
    ).returning(ExecutionItemSteps)

    row = (await session.execute(stmt)).scalar_one()
    await session.flush()

    return CreateStepResponse(
        code=200,
        message="上报用例步骤成功",
        data=[_to_step_row(row)],
    )


async def list_steps(session: AsyncSession, item_id: int) -> QueryStepListResponse:
    stmt = (
        select(ExecutionItemSteps)
        .where(ExecutionItemSteps.item_id == item_id)
        .order_by(ExecutionItemSteps.step_index.asc())
    )
    rows = (await session.execute(stmt)).all()

    return QueryStepListResponse(
        code=200,
        message="获取步骤列表成功",
        data=[_to_step_row(step, attachment_count=count) for step, count in rows],
    )


# =========================================================================
# 5. 用例跨 execution 历史趋势
# =========================================================================
async def get_case_history(
    session: AsyncSession, case_key: str, query: QueryExecutionItemHistory
) -> QueryExecutionItemHistoryResponse:
    """获取某个用例跨 execution 的最终结果历史 (用于稳定性/flaky 分析)"""
    stmt = (
        select(ExecutionItems, TestExecutions.job_name)
        .join(TestExecutions, TestExecutions.build_uid == ExecutionItems.build_uid)
        .where(
            ExecutionItems.case_key == case_key,
            ExecutionItems.is_latest.is_(True),
        )
    )

    count_stmt = select(func.count()).select_from(stmt.subquery())
    total = (await session.execute(count_stmt)).scalar_one() or 0

    offset = (query.page - 1) * query.page_size
    stmt = (
        stmt.order_by(ExecutionItems.start_at.desc().nullslast())
        .offset(offset)
        .limit(query.page_size)
    )
    rows = (await session.execute(stmt)).all()

    items = [
        ExecutionItemHistoryRow(
            build_uid=item.build_uid,
            job_name=job_name,
            item_id=item.id,
            attempt_number=item.attempt_number,
            status=item.status,
            start_at=item.start_at,
            duration=item.duration,
        )
        for item, job_name in rows
    ]

    list_data = ExecutionItemHistoryListData(
        total=total, page=query.page, page_size=query.page_size, rows=items
    )
    return QueryExecutionItemHistoryResponse(
        code=200, message="获取用例历史成功", data=list_data
    )
