from datetime import datetime
from typing import Optional

from pydantic import BaseModel, Field
from app.core.base_response import BaseResponse
from .schemas import (
    AttachmentType,
    CaseStatus,
    ExecutionStatus,
    StepStatus,
)


class ExecutionRow(BaseModel):
    id: int = Field(..., description="execution ID")
    build_uid: str = Field(..., description="本次 build 唯一标识")
    job_name: str = Field(..., description="任务名称")
    job_url: Optional[str] = Field(None, description="任务地址")
    project_name: str = Field(..., description="项目名称")
    software_name: str = Field(..., description="软件名称")
    software_version: str = Field(..., description="软件版本")
    labels: Optional[list[str]] = Field(None, description="本次 execution 的标签列表")
    status: ExecutionStatus = Field(..., description="执行状态")
    start_time: datetime = Field(..., description="开始时间")
    end_time: Optional[datetime] = Field(None, description="结束时间")
    duration: Optional[float] = Field(None, description="耗时(秒)")
    planned_cases_count: Optional[int] = Field(None, description="预计执行用例总数")
    pass_count: Optional[int] = Field(None, description="最终成功数")
    failure_count: Optional[int] = Field(None, description="最终失败数")
    skipped_count: Optional[int] = Field(None, description="最终跳过数")
    pass_rate: Optional[float] = Field(None, description="最终通过率")


class ExecutionData(BaseModel):
    id: int = Field(..., description="execution ID")
    build_uid: str = Field(..., description="本次 build 唯一标识")


class AttachmentItem(BaseModel):
    """步骤附件记录"""

    attachment_type: AttachmentType = Field(..., description="附件类型")
    file_name: str = Field(..., description="文件名")
    url: str = Field(..., description="文件外部访问 URL")
    mime_type: Optional[str] = Field(None, description="MIME 类型")


class ExecutionItemRow(BaseModel):
    """用例的一次 attempt 记录"""

    id: int = Field(..., description="item ID")
    build_uid: str = Field(..., description="本次 build 唯一标识")
    case_uid: str = Field(..., description="pytest run id")
    case_key: str = Field(..., description="Jira key，同一用例的唯一标识")
    case_name: str = Field(..., description="用例名称")
    labels: Optional[list[str]] = Field(None, description="用例标签列表")
    attempt_number: int = Field(..., description="第几次执行 (1=首次，2=第一次重试...)")
    is_latest: bool = Field(
        ..., description="是否为该用例在本次 execution 下的最终结果"
    )
    status: CaseStatus = Field(..., description="执行结果状态")
    start_time: datetime = Field(..., description="开始时间")
    end_time: Optional[datetime] = Field(None, description="结束时间")
    duration: Optional[float] = Field(None, description="耗时(秒)")
    error_message: Optional[str] = Field(None, description="失败概要信息")
    error_traceback: Optional[str] = Field(None, description="失败堆栈/错误信息")
    attachments: Optional[list[AttachmentItem]] = Field(
        None,
        description="附件列表：截图/日志等，存储为 JSON 数组 [{'type': 'screenshot', 'url': '...'}]",
    )
    # 第一层步骤列表，包含子步骤的嵌套结构
    steps: Optional[list["StepRow"]] = Field(
        None,
        description="第一层步骤列表，包含子步骤的嵌套结构",
    )


class ExecutionItemData(BaseModel):
    id: int = Field(..., description="item ID")
    build_uid: str = Field(..., description="本次 build 唯一标识")
    case_uid: str = Field(..., description="pytest run id")
    case_key: str = Field(..., description="Jira key，同一用例的唯一标识")


class StepRow(BaseModel):
    """用例步骤记录"""

    id: int = Field(..., description="步骤 ID")
    case_uid: str = Field(..., description="所属用例 attempt ID")
    parent_step_id: Optional[int] = Field(
        None, description="同 item 内父步骤的 step_id；null 表示根步骤"
    )
    parent_step_index: Optional[int] = Field(
        None, description="同 item 内父步骤的 step_index；null 表示根步骤"
    )

    step_index: int = Field(..., description="步骤顺序")
    step_path: str = Field(..., description="步骤路径，类似 '1.2.3' 的层级表示")
    depth: int = Field(..., description="步骤深度，根步骤为 0，子步骤依次加 1")

    step_name: str = Field(..., description="步骤名称")
    status: StepStatus = Field(..., description="步骤结果状态")
    start_time: datetime = Field(..., description="步骤开始时间")
    end_time: Optional[datetime] = Field(None, description="步骤结束时间")
    duration: Optional[float] = Field(None, description="耗时(秒)")
    # 第一层子步骤列表，包含子步骤的嵌套结构
    sub_steps: Optional[list["StepRow"]] = Field(
        None,
        description="第一层子步骤列表，包含子步骤的嵌套结构",
    )
    attachments: Optional[list[AttachmentItem]] = Field(
        None,
        description="附件列表：截图/日志等，存储为 JSON 数组 [{'type': 'screenshot', 'url': '...'}]",
    )


class StepData(BaseModel):
    id: int = Field(..., description="步骤 ID")
    case_uid: str = Field(..., description="所属用例 attempt ID")


class ExecutionItemHistoryRow(BaseModel):
    """用例跨 execution 的历史结果条目"""

    build_uid: str = Field(..., description="所属 execution build 唯一标识")
    job_name: str = Field(..., description="任务名称")
    case_uid: str = Field(..., description="所属用例 attempt ID")
    attempt_number: int = Field(..., description="本次 execution 内的最终 attempt 序号")
    status: CaseStatus = Field(..., description="执行结果状态")
    start_time: datetime = Field(..., description="开始时间")
    end_time: datetime = Field(..., description="结束时间")
    duration: float = Field(..., description="耗时(秒)")


class PipelineRow(BaseModel):
    """Jenkins 自动化任务列表项"""

    id: int = Field(..., description="任务 ID")
    job_name: str = Field(..., description="任务名称")
    job_url: str = Field(..., description="任务地址")
    status: str = Field(..., description="任务状态")
    last_build_number: Optional[int] = Field(None, description="最近一次构建号")
    last_build_uid: Optional[str] = Field(None, description="最近一次构建 UID")
    last_build_status: Optional[str] = Field(None, description="最近一次构建状态")
    last_build_timestamp: Optional[datetime] = Field(
        None, description="最近一次构建时间"
    )
    last_build_duration: Optional[float] = Field(
        None, description="最近一次构建耗时(秒)"
    )
    pipeline_params: Optional[dict] = Field(None, description="任务参数")
    sync_at: Optional[datetime] = Field(None, description="最近同步时间")


# ==============================================================================
# 2. 分页数据结构
# ==============================================================================
class ExecutionListData(BaseModel):
    total: int = Field(..., description="总记录数")
    page: int = Field(..., description="当前页码")
    page_size: int = Field(..., description="每页条数")
    rows: list[ExecutionRow] = Field(..., description="execution 列表项")


class PipelineListData(BaseModel):
    total: int = Field(..., description="总记录数")
    page: int = Field(..., description="当前页码")
    page_size: int = Field(..., description="每页条数")
    rows: list[PipelineRow] = Field(..., description="自动化任务列表项")


class ExecutionItemListData(BaseModel):
    total: int = Field(..., description="总记录数")
    page: int = Field(..., description="当前页码")
    page_size: int = Field(..., description="每页条数")
    rows: list[ExecutionItemRow] = Field(..., description="用例记录列表项")


class ExecutionItemHistoryListData(BaseModel):
    total: int = Field(..., description="总记录数")
    page: int = Field(..., description="当前页码")
    page_size: int = Field(..., description="每页条数")
    rows: list[ExecutionItemHistoryRow] = Field(..., description="历史结果列表项")


class CreateExecutionResponse(BaseResponse):
    data: ExecutionRow = Field(..., description="创建成功的 execution 信息")


class CreateExecutionErrorResponse(BaseResponse): ...


class UpdateExecutionResponse(BaseResponse):
    data: ExecutionRow = Field(..., description="更新后的 execution 信息")


class UpdateExecutionErrorResponse(BaseResponse): ...


class QueryExecutionListResponse(BaseResponse):
    data: ExecutionListData = Field(..., description="分页列表数据")


class QueryExecutionListErrorResponse(BaseResponse): ...


class QueryPipelineListResponse(BaseResponse):
    data: PipelineListData = Field(..., description="分页列表数据")


class QueryPipelineListErrorResponse(BaseResponse): ...


class CreateExecutionItemResponse(BaseResponse):
    data: ExecutionItemRow = Field(..., description="创建/复用的用例 attempt 信息")


class CreateExecutionItemErrorResponse(BaseResponse): ...


class UpdateExecutionItemResponse(BaseResponse):
    data: ExecutionItemRow = Field(..., description="更新后的用例 attempt 信息")


class UpdateExecutionItemErrorResponse(BaseResponse): ...


class QueryExecutionItemListResponse(BaseResponse):
    data: ExecutionItemListData = Field(..., description="分页列表数据")


class QueryExecutionItemListErrorResponse(BaseResponse): ...


class CreateStepResponse(BaseResponse):
    data: list[StepRow] = Field(..., description="创建/更新后的步骤列表")


class CreateStepErrorResponse(BaseResponse): ...


class QueryStepListResponse(BaseResponse):
    data: list[StepRow] = Field(..., description="步骤列表")


class QueryStepListErrorResponse(BaseResponse): ...


class QueryExecutionItemHistoryResponse(BaseResponse):
    data: ExecutionItemHistoryListData = Field(..., description="分页列表数据")


class QueryExecutionItemHistoryErrorResponse(BaseResponse): ...
