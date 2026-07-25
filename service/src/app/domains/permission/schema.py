from datetime import datetime
from typing import Optional, List
from pydantic import BaseModel, Field


# ==============================================================================
# 权限实体及请求模型
# ==============================================================================

class PermissionItem(BaseModel):
    """权限基础视图项"""
    id: int = Field(..., description="权限ID")
    parent_id: Optional[int] = Field(None, description="父权限ID")
    code: str = Field(..., description="权限唯一标识 (如: user:create, order:read)")
    name: str = Field(..., description="权限显示名称 (如: 创建用户)")
    resource_type: int = Field(..., description="资源类型 (1-菜单, 2-按钮/功能, 3-API接口)")
    path: Optional[str] = Field(None, description="前端路由或 API 路径")
    method: Optional[str] = Field(None, description="HTTP 请求方式 (GET, POST等)")
    created_at: Optional[datetime] = Field(None, description="创建时间")


class PermissionTreeItem(PermissionItem):
    """树形结构权限数据项"""
    children: List["PermissionTreeItem"] = Field(default_factory=list, description="子权限列表")


class PermissionCreate(BaseModel):
    """创建权限节点请求"""
    parent_id: Optional[int] = Field(None, description="父权限ID")
    code: str = Field(..., min_length=2, max_length=128, description="权限唯一标识 (如 user:create)")
    name: str = Field(..., min_length=2, max_length=64, description="权限名称")
    resource_type: int = Field(default=1, description="资源类型 (1-菜单, 2-按钮/功能, 3-API接口)")
    path: Optional[str] = Field(None, description="前端路径或 API Path")
    method: Optional[str] = Field(None, description="HTTP 请求方式 (GET, POST, PUT, DELETE等)")


class PermissionUpdate(BaseModel):
    """更新权限节点请求"""
    name: Optional[str] = Field(None, description="权限名称")
    path: Optional[str] = Field(None, description="路径")
    method: Optional[str] = Field(None, description="HTTP请求方式")


# ==============================================================================
# 角色实体及请求模型
# ==============================================================================

class RoleCreate(BaseModel):
    """创建角色请求"""
    code: str = Field(..., min_length=2, max_length=64, description="角色唯一标识 (如: editor, admin)")
    name: str = Field(..., min_length=2, max_length=64, description="角色显示名称 (如: 编辑人员)")
    description: Optional[str] = Field(None, description="角色描述信息")
    permission_ids: Optional[List[int]] = Field(default_factory=list, description="初始绑定的权限ID列表")


class RoleUpdate(BaseModel):
    """修改角色基本信息请求"""
    name: Optional[str] = Field(None, description="角色名称")
    description: Optional[str] = Field(None, description="角色描述")


class RolePermissionAssign(BaseModel):
    """配置/分配角色权限请求"""
    role_id: int = Field(..., description="目标角色ID")
    permission_ids: List[int] = Field(..., description="绑定的权限ID列表")
