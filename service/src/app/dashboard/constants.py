from datetime import date, datetime
from typing import Optional

from pydantic import BaseModel, Field

from app.core.base_response import ErrorResponse, NormalResponse
from app.domains.automation.schema import CaseStatus, ExecutionStatus


class DashboardSummaryItem(BaseModel):
    active_jobs_count: int = Field(..., description="启用中的自动化任务数")
    total_executions_count: int = Field(..., description="execution 总数")
    running_executions_count: int = Field(..., description="运行中的 execution 数")
    success_cases_7d: int = Field(..., description="近 7 天成功用例数")
    failure_cases_7d: int = Field(..., description="近 7 天失败用例数")
    skipped_cases_7d: int = Field(..., description="近 7 天跳过用例数")
    pass_rate_7d: float = Field(..., description="近 7 天最终通过率")
    avg_execution_duration_7d: Optional[float] = Field(
        None, description="近 7 天平均 execution 耗时(秒)"
    )
    updated_at: datetime = Field(..., description="快照更新时间")


class DashboardTrendItem(BaseModel):
    stat_date: date = Field(..., description="统计日期")
    execution_total: int = Field(..., description="当日 execution 数")
    success_cases: int = Field(..., description="当日成功用例数")
    failure_cases: int = Field(..., description="当日失败用例数")
    skipped_cases: int = Field(..., description="当日跳过用例数")
    running_cases: int = Field(..., description="当日运行中用例数")


class DashboardRunningExecutionItem(BaseModel):
    execution_id: int = Field(..., description="execution ID")
    job_id: int = Field(..., description="Jenkins 任务 ID")
    job_name: str = Field(..., description="任务名称")
    job_url: str = Field(..., description="任务地址")
    status: ExecutionStatus = Field(..., description="execution 状态")
    started_at: datetime = Field(..., description="开始时间")
    duration: Optional[float] = Field(None, description="当前已运行时长(秒)")
    pre_cases_count: int = Field(..., description="预计用例总数")
    completed_cases_count: int = Field(..., description="已完成用例数")
    success_cases_count: int = Field(..., description="已通过用例数")
    failure_cases_count: int = Field(..., description="已失败用例数")
    skipped_cases_count: int = Field(..., description="已跳过用例数")
    running_cases_count: int = Field(..., description="运行中用例数")
    progress_percent: float = Field(..., description="当前进度百分比")
    running_case_names: list[str] = Field(
        default_factory=list, description="当前运行中用例名称"
    )
    updated_at: datetime = Field(..., description="快照更新时间")


class DashboardRunningCaseItem(BaseModel):
    execution_id: int = Field(..., description="execution ID")
    item_id: int = Field(..., description="当前生效 attempt ID")
    case_key: str = Field(..., description="pytest nodeid")
    case_name: str = Field(..., description="用例名称")
    attempt_number: int = Field(..., description="当前 attempt 序号")
    status: CaseStatus = Field(..., description="当前状态")
    start_at: Optional[datetime] = Field(None, description="开始时间")
    end_at: Optional[datetime] = Field(None, description="结束时间")
    duration: Optional[float] = Field(None, description="耗时(秒)")
    error_message: Optional[str] = Field(None, description="失败概要")
    updated_at: datetime = Field(..., description="快照更新时间")


class DashboardOverviewData(BaseModel):
    summary: DashboardSummaryItem = Field(..., description="统计卡片数据")
    trends: list[DashboardTrendItem] = Field(..., description="趋势图数据")
    running_executions: list[DashboardRunningExecutionItem] = Field(
        ..., description="运行中的 execution 卡片列表"
    )


class DashboardRunningCaseListData(BaseModel):
    execution_id: int = Field(..., description="execution ID")
    items: list[DashboardRunningCaseItem] = Field(..., description="case 明细列表")


class DashboardOverviewResponse(NormalResponse):
    message: str = Field(
        default="Get Dashboard Overview Success", description="响应消息"
    )
    data: DashboardOverviewData = Field(..., description="dashboard 概览数据")


class DashboardOverviewErrorResponse(ErrorResponse):
    message: str = Field(
        default="Get Dashboard Overview Failed", description="响应消息"
    )


class DashboardRunningCaseListResponse(NormalResponse):
    message: str = Field(
        default="Get Running Execution Cases Success", description="响应消息"
    )
    data: DashboardRunningCaseListData = Field(
        ..., description="运行中 execution 的 case 快照"
    )


class DashboardRunningCaseListErrorResponse(ErrorResponse):
    message: str = Field(
        default="Get Running Execution Cases Failed", description="响应消息"
    )
