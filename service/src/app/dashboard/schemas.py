from datetime import datetime
from typing import Literal, Optional
from uuid import UUID
from dataclasses import dataclass
from pydantic import BaseModel, Field

ExecutionStatus = Literal["running", "completed", "failed", "aborted", "calculated"]
CaseStatus = Literal["running", "passed", "failed", "skipped", "broken"]
StepStatus = CaseStatus
AttachmentType = Literal["screenshot", "log", "video", "other"]


@dataclass(frozen=True)
class ExecutionStatusModel:
    """Execution 状态枚举模型"""

    RUNNING: str = "running"
    COMPLETED: str = "completed"
    FAILED: str = "failed"
    ABORTED: str = "aborted"
    CALCULATED: str = "calculated"  # 计算完成，ETL 已完成


@dataclass(frozen=True)
class CaseStatusModel:
    """Case 状态枚举模型"""

    RUNNING: str = "running"
    PASSED: str = "passed"
    FAILED: str = "failed"
    SKIPPED: str = "skipped"
    BROKEN: str = "broken"


@dataclass(frozen=True)
class StepStatusModel:
    """Step 状态枚举模型"""

    RUNNING: str = "running"
    PASSED: str = "passed"
    FAILED: str = "failed"
    SKIPPED: str = "skipped"
    BROKEN: str = "broken"


# ==============================================================================
# 1. Execution (一次 pytest 执行) 模型
# ==============================================================================
class ExecutionCreate(BaseModel):
    """上报一次 execution 开始 (pytest_sessionstart)"""

    build_uid: UUID = Field(..., description="本次 Jenkins build 唯一标识")
    job_name: str = Field(..., max_length=255, description="任务名称")
    job_url: str = Field(..., max_length=255, description="任务地址")
    project_name: str = Field(..., max_length=255, description="项目名称")
    software_name: str = Field(..., max_length=255, description="软件名称")
    software_version: str = Field(..., max_length=255, description="软件版本")
    labels: Optional[list[str]] = Field(None, description="本次执行的标签列表")
    start_time: datetime = Field(..., description="开始时间")
    planned_cases_count: Optional[int] = Field(None, description="预计执行用例总数")


class ExecutionUpdate(BaseModel):
    """上报一次 execution 结束/更新 (pytest_sessionfinish)"""

    status: Optional[ExecutionStatus] = Field(None, description="执行状态")
    end_time: Optional[datetime] = Field(None, description="结束时间")
    duration: Optional[float] = Field(None, description="耗时(秒)")


class ExecutionListQuery(BaseModel):
    """execution 列表查询参数"""

    page: int = Field(default=1, ge=1, description="页码")
    page_size: int = Field(default=10, ge=1, le=100, description="每页条数")
    build_uid: Optional[UUID] = Field(None, description="按 Jenkins 构建 UID 筛选")
    job_name: Optional[str] = Field(None, description="按 Jenkins 任务名称模糊搜索")
    status: Optional[ExecutionStatus] = Field(None, description="按状态筛选")


# ==============================================================================
# 1.1 Jenkins Job (自动化任务) 查询模型
# ==============================================================================
class JobListQuery(BaseModel):
    """自动化任务列表查询参数"""

    page: int = Field(default=1, ge=1, description="页码")
    page_size: int = Field(default=10, ge=1, le=100, description="每页条数")
    job_name: Optional[str] = Field(None, description="按任务名称模糊搜索")
    status: Optional[Literal["active", "inactive"]] = Field(
        None, description="按任务状态筛选"
    )
    last_build_status: Optional[
        Literal["success", "failure", "unstable", "aborted"]
    ] = Field(None, description="按最近一次构建状态筛选")


# ==============================================================================
# 2. ExecutionCase (一次用例的一次 attempt) 模型
# ==============================================================================
class ExecutionCaseCreate(BaseModel):
    """上报一次用例执行开始 (每次重试都调用一次)"""

    build_uid: str = Field(..., description="本次 Jenkins build 唯一标识")
    case_key: str = Field(..., max_length=512, description="JIRA用例的唯一标识")
    case_name: str = Field(..., max_length=255, description="用例名称")
    case_uid: UUID = Field(
        ...,
        description="Pytest 运行时生成的唯一标识符 (pytest_runtest_protocol hook 中的 item.nodeid)",
    )
    labels: Optional[list[str]] = Field(None, description="本次执行的标签列表")
    start_time: Optional[datetime] = Field(None, description="本次 attempt 开始时间")


class ExecutionCaseUpdate(BaseModel):
    """上报一次用例执行结束"""

    status: CaseStatus = Field(..., description="执行结果状态")
    end_time: Optional[datetime] = Field(None, description="结束时间")
    duration: Optional[float] = Field(None, description="耗时(秒)")
    error_message: Optional[str] = Field(None, description="失败概要信息")
    error_traceback: Optional[str] = Field(
        None, description="失败时的完整堆栈/错误信息"
    )
    attachments: Optional[list[dict]] = Field(
        None, description="本次 attempt 的附件列表，每个附件包含 type、name、url 等字段"
    )


class ExecutionCaseListQuery(BaseModel):
    """execution 下用例列表查询参数 (默认只返回每个 case 的最终结果)"""

    page: int = Field(default=1, ge=1, description="页码")
    page_size: int = Field(default=10, ge=1, le=100, description="每页条数")
    status: Optional[CaseStatus] = Field(None, description="按状态筛选")
    case_key: Optional[str] = Field(None, description="按 case_key 精确匹配")
    case_name: Optional[str] = Field(None, description="按用例名称模糊搜索")


# ==============================================================================
# 3. Step (用例步骤) 模型
# ==============================================================================
class StepCreate(BaseModel):
    """单个步骤上报数据"""

    step_index: Optional[int] = Field(
        None, ge=0, description="步骤顺序，缺省则按已有步骤数自动追加"
    )
    parent_step_index: Optional[int] = Field(
        None,
        ge=0,
        description="同 item 内父步骤的 step_index；null 表示根步骤",
    )
    step_path: str = Field(
        ..., max_length=255, description="步骤路径，唯一标识一个步骤"
    )
    step_name: str = Field(..., max_length=255, description="步骤名称")
    status: StepStatus = Field(..., description="步骤结果状态")
    start_time: Optional[datetime] = Field(None, description="步骤开始时间")
    end_time: Optional[datetime] = Field(None, description="步骤结束时间")
    duration: Optional[float] = Field(None, description="耗时(秒)")
    error_message: Optional[str] = Field(None, description="未通过时的错误堆栈/描述")
    error_traceback: Optional[str] = Field(
        None, description="未通过时的完整堆栈/错误信息"
    )
    attachments: Optional[list[dict]] = Field(
        None, description="本步骤的附件列表，每个附件包含 type、name、url 等字段"
    )


class StepsBatchCreate(BaseModel):
    """批量上报步骤请求"""

    steps: list[StepCreate] = Field(..., min_length=1, description="本次上报的步骤列表")


# ==============================================================================
# 5. 跨执行历史查询模型
# ==============================================================================
class CaseHistoryQuery(BaseModel):
    """某个用例跨 execution 的历史结果查询参数"""

    page: int = Field(default=1, ge=1, description="页码")
    page_size: int = Field(default=10, ge=1, le=100, description="每页条数")
