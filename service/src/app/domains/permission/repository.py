from datetime import datetime
from typing import Optional, List
from pydantic import BaseModel, Field

from app.core.base_response import NormalResponse, ErrorResponse
from app.domains.permission.schema import PermissionItem, PermissionTreeItem


class RoleDetailData(BaseModel):
    """角色详情与所拥有的权限关联数据"""
    id: int = Field(..., description="角色ID")
    code: str = Field(..., description="角色标识")
    name: str = Field(..., description="角色名称")
    description: Optional[str] = Field(None, description="角色描述")
    is_system: bool = Field(default=False, description="是否为内置系统角色")
    permissions: List[PermissionItem] = Field(default_factory=list, description="权限列表")
    created_at: Optional[datetime] = Field(None, description="创建时间")
    updated_at: Optional[datetime] = Field(None, description="更新时间")


class RoleListResponse(NormalResponse):
    """角色列表响应"""
    message: str = Field(default="Get Role List Success", description="响应消息")
    data: List[RoleDetailData] = Field(..., description="角色列表数据")


class RoleDetailResponse(NormalResponse):
    """角色详情响应"""
    message: str = Field(default="Get Role Detail Success", description="响应消息")
    data: RoleDetailData = Field(..., description="角色详情数据")


class PermissionTreeResponse(NormalResponse):
    """权限树型结构响应"""
    message: str = Field(default="Get Permission Tree Success", description="响应消息")
    data: List[PermissionTreeItem] = Field(..., description="树形权限列表")


class PermissionListResponse(NormalResponse):
    """权限平铺列表响应"""
    message: str = Field(default="Get Permission List Success", description="响应消息")
    data: List[PermissionItem] = Field(..., description="平铺权限列表")
