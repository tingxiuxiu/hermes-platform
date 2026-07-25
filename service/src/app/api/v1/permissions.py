from typing import Annotated

from fastapi import APIRouter, Depends

from app.dependencies import SessionDep, get_current_active_admin
from app.models.hdp.user import User
from app.core.base_response import NormalResponse
from app.domains.permission.schema import PermissionCreate
from app.domains.permission.repository import PermissionTreeResponse, PermissionListResponse
from app.domains.permission.service import RolePermissionService

router = APIRouter(prefix="/permissions", tags=["权限管理"])


@router.get(
    "/tree",
    response_model=PermissionTreeResponse,
    summary="获取树形结构权限列表 (管理员权限)",
)
async def get_permission_tree(
    session: SessionDep,
    current_admin: Annotated[User, Depends(get_current_active_admin)],
) -> PermissionTreeResponse:
    """获取格式化后的树形菜单与权限节点数据"""
    return await RolePermissionService.get_permission_tree(session=session)


@router.get(
    "",
    response_model=PermissionListResponse,
    summary="获取平铺权限列表 (管理员权限)",
)
async def get_permission_list(
    session: SessionDep,
    current_admin: Annotated[User, Depends(get_current_active_admin)],
) -> PermissionListResponse:
    """获取所有平铺格式的权限项"""
    return await RolePermissionService.get_permission_list(session=session)


@router.post(
    "",
    response_model=NormalResponse,
    summary="创建新权限节点 (管理员权限)",
)
async def create_permission(
    session: SessionDep,
    data: PermissionCreate,
    current_admin: Annotated[User, Depends(get_current_active_admin)],
) -> NormalResponse:
    """新增菜单、按钮或 API 接口权限节点"""
    return await RolePermissionService.create_permission(session=session, data=data)


@router.delete(
    "/{permission_id}",
    response_model=NormalResponse,
    summary="删除指定权限节点 (管理员权限)",
)
async def delete_permission(
    session: SessionDep,
    permission_id: int,
    current_admin: Annotated[User, Depends(get_current_active_admin)],
) -> NormalResponse:
    """删除指定的权限节点"""
    return await RolePermissionService.delete_permission(
        session=session, permission_id=permission_id
    )
