from typing import Any
from pydantic import BaseModel, Field


class NormalResponse(BaseModel):
    code: int = Field(..., description="响应码")
    message: str = Field(..., description="响应消息")
    data: Any = Field(None, description="响应数据")
    success: bool = True


class ErrorResponse(BaseModel):
    code: int = Field(..., description="响应码")
    message: str = Field(..., description="响应消息")
    data: Any = Field(None, description="响应数据")
    success: bool = False
