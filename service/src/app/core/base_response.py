from typing import Any
from pydantic import BaseModel, Field


class BaseResponse(BaseModel):
    code: int = Field(..., description="响应码")
    message: str = Field(..., description="响应消息")
    data: Any = Field(None, description="响应数据")
    success: bool = True
