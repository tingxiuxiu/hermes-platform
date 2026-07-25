from typing import Annotated

from fastapi import APIRouter, Depends

from app.dependencies import SessionDep, get_current_active_admin
from app.models.hdp.user import User
from app.core.base_response import NormalResponse
from app.domains.permission.schema import RoleCreate, RoleUpdate, RolePermissionAssign
from app.domains.permission.repository import RoleListResponse, RoleDetailResponse
from app.domains.permission.service import RolePermissionService

router = APIRouter(prefix="/roles", tags=["角色管理"])


@router.get(
    "",
    response_model=RoleListResponse,
    summary="获取角色列表 (管理员权限)",
)
async def get_roles(
    session: SessionDep,
    current_admin: Annotated[User, Depends(get_current_active_admin)],
) -> RoleListResponse:
    """获取系统中所有角色及绑定的权限列表"""
    return await RolePermissionService.get_roles(session=session)


@router.post(
    "",
    response_model=RoleDetailResponse,
    summary="创建新角色 (管理员权限)",
)
async def create_role(
    session: SessionDep,
    data: RoleCreate,
    current_admin: Annotated[User, Depends(get_current_active_admin)],
) -> RoleDetailResponse:
    """新建角色并绑定初始权限集合"""
    return await RolePermissionService.create_role(session=session, data=data)


@router.get(
    "/{role_id}",
    response_model=RoleDetailResponse,
    summary="获取指定角色详情 (管理员权限)",
)
async def get_role_detail(
    session: SessionDep,
    role_id: int,
    current_admin: Annotated[User, Depends(get_current_active_admin)],
) -> RoleDetailResponse:
    """查询单个角色的详细数据及其关联权限"""
    return await RolePermissionService.get_role_detail(session=session, role_id=role_id)


@router.put(
    "/{role_id}",
    response_model=RoleDetailResponse,
    summary="修改角色信息 (管理员权限)",
)
async def update_role(
    session: SessionDep,
    role_id: int,
    data: RoleUpdate,
    current_admin: Annotated[User, Depends(get_current_active_admin)],
) -> RoleDetailResponse:
    """更新角色的名称或描述信息"""
    return await RolePermissionService.update_role(
        session=session, role_id=role_id, data=data
    )


@router.delete(
    "/{role_id}",
    response_model=NormalResponse,
    summary="删除指定角色 (管理员权限)",
)
async def delete_role(
    session: SessionDep,
    role_id: int,
    current_admin: Annotated[User, Depends(get_current_active_admin)],
) -> NormalResponse:
    """删除自定义角色 (内置系统角色禁止删除)"""
    return await RolePermissionService.delete_role(session=session, role_id=role_id)


@router.put(
    "/{role_id}/permissions",
    response_model=NormalResponse,
    summary="为角色分配/变更权限 (管理员权限)",
)
async def assign_role_permissions(
    session: SessionDep,
    role_id: int,
    data: RolePermissionAssign,
    current_admin: Annotated[User, Depends(get_current_active_admin)],
) -> NormalResponse:
    """重新为角色赋予权限节点集合"""
    return await RolePermissionService.assign_role_permissions(
        session=session, role_id=role_id, permission_ids=data.permission_ids
    )
