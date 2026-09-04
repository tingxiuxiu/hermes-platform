from datetime import datetime, timezone
from enum import Enum
from typing import Optional

from pydantic import BaseModel, Field


def utc_now() -> str:
    """RFC3339 UTC，对齐 Go `time.Time` JSON。"""
    return datetime.now(timezone.utc).strftime("%Y-%m-%dT%H:%M:%SZ")


class ExecutionStatusEnum(str, Enum):
    RUNNING = "running"
    COMPLETED = "completed"
    FAILED = "failed"
    ABORTED = "aborted"
    CALCULATED = "calculated"


class CaseStatusEnum(str, Enum):
    RUNNING = "running"
    PASSED = "passed"
    FAILED = "failed"
    SKIPPED = "skipped"
    BROKEN = "broken"


class StepStatusEnum(str, Enum):
    RUNNING = "running"
    PASSED = "passed"
    FAILED = "failed"
    SKIPPED = "skipped"
    BROKEN = "broken"


class AttachmentTypeEnum(str, Enum):
    SCREENSHOT = "screenshot"
    LOG = "log"
    VIDEO = "video"
    OTHER = "other"


class CreateExecution(BaseModel):
    build_uid: str
    job_name: str
    job_url: str
    project_name: str
    software_name: str
    software_version: str
    labels: list[str] = Field(default_factory=list)
    start_time: str
    planned_cases_count: int | None = None

    def to_dict(self) -> dict:
        return {
            "build_uid": self.build_uid,
            "job_name": self.job_name,
            "job_url": self.job_url,
            "project_name": self.project_name,
            "software_name": self.software_name,
            "software_version": self.software_version,
            "labels": self.labels,
            "start_time": self.start_time,
            "planned_cases_count": self.planned_cases_count,
        }


class UpdateExecution(BaseModel):
    status: ExecutionStatusEnum
    end_time: str
    duration: float

    def to_dict(self) -> dict:
        return {
            "status": self.status.value,
            "end_time": self.end_time,
            "duration": self.duration,
        }


class CreateCase(BaseModel):
    build_uid: str
    case_key: str
    case_name: str
    case_uid: str
    labels: list[str] = Field(default_factory=list)
    start_time: str

    def to_dict(self) -> dict:
        return {
            "build_uid": self.build_uid,
            "case_key": self.case_key,
            "case_name": self.case_name,
            "case_uid": self.case_uid,
            "labels": self.labels,
            "start_time": self.start_time,
        }


class UpdateCase(BaseModel):
    status: CaseStatusEnum
    end_time: str
    duration: float
    error_message: str | None = None
    error_traceback: str | None = None

    def to_dict(self) -> dict:
        return {
            "status": self.status.value,
            "end_time": self.end_time,
            "duration": self.duration,
            "error_message": self.error_message or "",
            "error_traceback": self.error_traceback or "",
        }


class UpsertStep(BaseModel):
    step_path: str
    step_name: str
    status: StepStatusEnum
    start_time: str
    end_time: Optional[str] = None
    duration: Optional[float] = None

    def to_dict(self) -> dict:
        body: dict = {
            "step_path": self.step_path,
            "step_name": self.step_name,
            "status": self.status.value,
            "start_time": self.start_time,
        }
        if self.end_time is not None:
            body["end_time"] = self.end_time
        if self.duration is not None:
            body["duration"] = self.duration
        return body
