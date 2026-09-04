"""会话期内的当前 case 上下文。步骤真相源是 Allure，`step` 只是 `allure.step` 别名。"""

from __future__ import annotations

import uuid
from contextvars import ContextVar
from dataclasses import dataclass, field
from typing import Any, Optional

import allure
from allure_commons.types import AttachmentType

from hermes_plugin.models import AttachmentTypeEnum, CaseStatusEnum, utc_now

step = allure.step

_ATTACHMENT_TYPE = {
    AttachmentTypeEnum.SCREENSHOT: AttachmentType.PNG,
    AttachmentTypeEnum.LOG: AttachmentType.TEXT,
    AttachmentTypeEnum.VIDEO: AttachmentType.MP4,
    AttachmentTypeEnum.OTHER: AttachmentType.TEXT,
}


@dataclass
class StepFrame:
    path: str
    name: str
    start_time: str
    child_count: int = 0


@dataclass
class CaseRecord:
    build_uid: str
    case_key: str
    case_name: str
    case_uid: str = field(default_factory=lambda: str(uuid.uuid4()))
    start_time: str = field(default_factory=utc_now)
    status: CaseStatusEnum = CaseStatusEnum.RUNNING
    error_message: Optional[str] = None
    error_traceback: Optional[str] = None
    case_labels: list[str] = field(default_factory=list)
    step_stack: list[StepFrame] = field(default_factory=list)
    root_count: int = 0
    create_future: Any = None


_current_case: ContextVar[Optional[CaseRecord]] = ContextVar("_current_case", default=None)


def attachment(name: str, content: Any, attachment_type: AttachmentTypeEnum = AttachmentTypeEnum.LOG) -> None:
    """写入 Allure 附件，不上报 Hermes（v1 无对象存储）。"""
    at = _ATTACHMENT_TYPE.get(attachment_type, AttachmentType.TEXT)
    if isinstance(content, bytes):
        allure.attach(content, name=name, attachment_type=at)
    else:
        allure.attach(str(content), name=name, attachment_type=AttachmentType.TEXT)
