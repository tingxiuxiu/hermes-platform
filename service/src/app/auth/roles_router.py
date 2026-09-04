from typing import Annotated

from fastapi import APIRouter, Depends

from app.dependencies import SessionDep, get_current_active_admin
from .models import User
from .schemas import RoleCreate, RoleUpdate, RolePermissionAssign
from .constants import RoleListResponse, RoleDetailResponse, BaseResponse
from .service import (
    get_roles,
    create_role,
    get_role_detail,
    update_role,
    delete_role,
    assign_role_permissions,
)

roles_router = APIRouter(prefix="/roles", tags=["角色管理"])


@roles_router.get(
    "",
    response_model=RoleListResponse,
    summary="获取角色列表 (管理员权限)",
)
async def get_roles_endpoint(
    session: SessionDep,
    current_admin: Annotated[User, Depends(get_current_active_admin)],
) -> RoleListResponse:
    """获取系统中所有角色及绑定的权限列表"""
    return await get_roles(session=session)


@roles_router.post(
    "",
    response_model=RoleDetailResponse,
    summary="创建新角色 (管理员权限)",
)
async def create_role_endpoint(
    session: SessionDep,
    data: RoleCreate,
    current_admin: Annotated[User, Depends(get_current_active_admin)],
) -> RoleDetailResponse:
    """新建角色并绑定初始权限集合"""
    return await create_role(session=session, data=data)


@roles_router.get(
    "/{role_id}",
    response_model=RoleDetailResponse,
    summary="获取指定角色详情 (管理员权限)",
)
async def get_role_detail_endpoint(
    session: SessionDep,
    role_id: int,
    current_admin: Annotated[User, Depends(get_current_active_admin)],
) -> RoleDetailResponse:
    """查询单个角色的详细数据及其关联权限"""
    return await get_role_detail(session=session, role_id=role_id)


@roles_router.put(
    "/{role_id}",
    response_model=RoleDetailResponse,
    summary="修改角色信息 (管理员权限)",
)
async def update_role_endpoint(
    session: SessionDep,
    role_id: int,
    data: RoleUpdate,
    current_admin: Annotated[User, Depends(get_current_active_admin)],
) -> RoleDetailResponse:
    """更新角色的名称或描述信息"""
    return await update_role(session=session, role_id=role_id, data=data)


@roles_router.delete(
    "/{role_id}",
    response_model=BaseResponse,
    summary="删除指定角色 (管理员权限)",
)
async def delete_role_endpoint(
    session: SessionDep,
    role_id: int,
    current_admin: Annotated[User, Depends(get_current_active_admin)],
) -> BaseResponse:
    """删除自定义角色 (内置系统角色禁止删除)"""
    return await delete_role(session=session, role_id=role_id)


@roles_router.put(
    "/{role_id}/permissions",
    response_model=BaseResponse,
    summary="为角色分配/变更权限 (管理员权限)",
)
async def assign_role_permissions_endpoint(
    session: SessionDep,
    role_id: int,
    data: RolePermissionAssign,
    current_admin: Annotated[User, Depends(get_current_active_admin)],
) -> BaseResponse:
    """重新为角色赋予权限节点集合"""
    return await assign_role_permissions(
        session=session, role_id=role_id, permission_ids=data.permission_ids
    )
