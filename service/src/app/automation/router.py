from typing import Annotated

from app.auth.models import User
from fastapi import APIRouter, Depends, Query
from structlog import get_logger

from app.dependencies import SessionDep, get_current_user, get_service_token
from .constants import (
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
    QueryExecutionItemHistoryResponse,
    QueryExecutionItemHistoryErrorResponse,
)
from .schemas import (
    CreateExecution,
    UpdateExecution,
    QueryExecutionList,
    QueryPipelineList,
    CreateExecutionItem,
    UpdateExecutionItem,
    QueryExecutionItemList,
    AttachmentItem,
    CreateStep,
    QueryExecutionItemHistory,
    ExecutionStatus,
    CaseStatus,
    AttachmentType,
)
from .service import (
    list_pipelines,
    create_execution,
    update_execution,
    get_execution_row,
    list_executions,
    create_execution_item,
    update_execution_item,
    list_execution_items,
    get_execution_item_attempts,
    get_case_detail,
    create_steps,
    list_steps,
    get_case_history,
)

logger = get_logger(__name__)
automation_router = APIRouter(prefix="/automation", tags=["自动化用例执行"])

# 查询类接口 (供前端 Dashboard 使用)：复用现有用户 JWT 鉴权
CurrentUser = Annotated[User, Depends(get_current_user)]
# 上报类接口 (供 pytest/CI Runner 调用)：使用独立的服务令牌鉴权 (X-Service-Token)，与用户登录态区分
ServiceAuth = Annotated[str, Depends(get_service_token)]


# ==============================================================================
# 1. Job 查询
# ==============================================================================
@automation_router.get(
    "/pipelines",
    response_model=QueryPipelineListResponse,
    summary="获取自动化任务列表",
)
async def list_pipelines_endpoint(
    session: SessionDep,
    current_user: CurrentUser,
    page: int = Query(default=1, ge=1, description="页码"),
    page_size: int = Query(default=10, ge=1, le=100, description="每页条数"),
    job_name: str | None = Query(default=None, description="按任务名称模糊搜索"),
    status: str | None = Query(default=None, description="按任务状态筛选"),
    last_build_status: str | None = Query(
        default=None, description="按最近一次构建状态筛选"
    ),
) -> QueryPipelineListResponse:
    query = QueryPipelineList(
        page=page,
        page_size=page_size,
        job_name=job_name,
        status=status,
        last_build_status=last_build_status,
    )
    return await list_pipelines(session=session, query=query)


# ==============================================================================
# 2. Execution 上报与查询
# ==============================================================================
@automation_router.post(
    "/executions",
    response_model=CreateExecutionResponse,
    summary="创建一次执行 (pytest session 开始上报)",
)
async def create_execution_endpoint(
    session: SessionDep,
    data: CreateExecution,
    service_auth: ServiceAuth,
) -> CreateExecutionResponse:
    """pytest_sessionstart 时调用，创建一次 job execution 记录"""
    try:
        return await create_execution(session=session, data=data)
    except Exception as e:
        logger.exception(f"Failed to create execution: {e}")
        return CreateExecutionErrorResponse(success=False, message=str(e))


@automation_router.patch(
    "/executions/{build_uid}",
    response_model=UpdateExecutionResponse,
    summary="结束/更新一次执行 (pytest session 结束上报)",
)
async def update_execution_endpoint(
    session: SessionDep,
    build_uid: str,
    data: UpdateExecution,
    service_auth: ServiceAuth,
) -> UpdateExecutionResponse:
    """pytest_sessionfinish 时调用，更新执行状态与最终统计数据"""
    return await update_execution(session=session, build_uid=build_uid, data=data)


@automation_router.get(
    "/executions",
    response_model=QueryExecutionListResponse,
    summary="获取执行列表",
)
async def list_executions_endpoint(
    session: SessionDep,
    current_user: CurrentUser,
    page: int = Query(default=1, ge=1, description="页码"),
    page_size: int = Query(default=10, ge=1, le=100, description="每页条数"),
    build_uid: str | None = Query(default=None, description="按最近一次构建 UID 筛选"),
    job_name: str | None = Query(
        default=None, description="按 Jenkins 任务名称模糊搜索"
    ),
    status: ExecutionStatus | None = Query(default=None, description="按状态筛选"),
) -> QueryExecutionListResponse:
    """分页查询 execution 列表"""
    query = QueryExecutionList(
        page=page,
        page_size=page_size,
        build_uid=build_uid,
        job_name=job_name,
        status=status,
    )
    return await list_executions(session=session, query=query)


@automation_router.get(
    "/executions/{build_uid}",
    response_model=QueryExecutionItemList,
    summary="获取执行详情",
)
async def get_execution_endpoint(
    session: SessionDep,
    build_uid: str,
    current_user: CurrentUser,
) -> QueryExecutionItemList:
    """获取单次 execution 的详情与统计汇总"""
    return await get_execution_row(session=session, build_uid=build_uid)


# ==============================================================================
# 2. Item (用例 attempt) 上报与查询
# ==============================================================================
@automation_router.post(
    "/executions/{build_uid}/items",
    response_model=CreateExecutionItemResponse,
    summary="上报一次用例执行开始 (含失败重试)",
)
async def create_item_endpoint(
    session: SessionDep,
    build_uid: str,
    data: CreateExecutionItem,
    service_auth: ServiceAuth,
) -> CreateExecutionItemResponse:
    """每个用例开始执行时调用一次 (包括每次自动重试)，自动分配 attempt_number 且不覆盖历史记录"""
    return await create_execution_item(session=session, build_uid=build_uid, data=data)


@automation_router.patch(
    "/executions/{build_uid}/items/{case_uid}",
    response_model=UpdateExecutionItemResponse,
    summary="上报一次用例执行结束",
)
async def update_item_endpoint(
    session: SessionDep,
    build_uid: str,
    case_uid: str,
    data: UpdateExecutionItem,
    service_auth: ServiceAuth,
) -> UpdateExecutionItemResponse:
    """用例执行完成后调用，更新本次 attempt 的最终结果"""
    return await update_execution_item(
        session=session, build_uid=build_uid, case_uid=case_uid, data=data
    )


@automation_router.get(
    "/executions/{build_uid}/items/attempts",
    response_model=QueryExecutionItemHistoryResponse,
    summary="获取某个用例在本次执行下的全部重试历史",
)
async def get_case_attempts_endpoint(
    session: SessionDep,
    build_uid: str,
    current_user: CurrentUser,
    case_key: str = Query(..., description="pytest nodeid"),
) -> QueryExecutionItemHistoryResponse:
    """按 attempt_number 升序返回该用例的全部执行记录 (首次运行 + 每次重试)"""
    return await get_execution_item_attempts(
        session=session, build_uid=build_uid, case_key=case_key
    )


# ==============================================================================
# 3. Step 上报与查询
# ==============================================================================
@automation_router.post(
    "/executions/{build_uid}/items/{item_id}/steps",
    response_model=CreateStepResponse,
    summary="批量上报用例步骤",
)
async def add_steps_endpoint(
    session: SessionDep,
    build_uid: str,
    item_id: int,
    data: CreateStep,
    service_auth: ServiceAuth,
) -> CreateStepResponse:
    return await create_steps(session=session, item_id=item_id, data=data)


# ==============================================================================
# 5. 用例跨执行历史趋势
# ==============================================================================
@automation_router.get(
    "/cases/history",
    response_model=QueryExecutionItemHistoryResponse,
    summary="获取用例跨执行的历史结果 (稳定性/flaky 分析)",
)
async def get_case_history_endpoint(
    session: SessionDep,
    current_user: CurrentUser,
    case_key: str = Query(..., description="pytest nodeid"),
    page: int = Query(default=1, ge=1, description="页码"),
    page_size: int = Query(default=10, ge=1, le=100, description="每页条数"),
) -> QueryExecutionItemHistoryResponse:
    """按时间倒序返回该用例在各次 execution 中的最终结果，用于识别间歇性失败(flaky)用例"""
    query = QueryExecutionItemHistory(page=page, page_size=page_size)
    return await get_case_history(session=session, case_key=case_key, query=query)
