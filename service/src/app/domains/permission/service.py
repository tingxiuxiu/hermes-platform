from typing import List, Dict, Optional
from sqlalchemy import select
from sqlalchemy.orm import selectinload
from sqlalchemy.ext.asyncio import AsyncSession
from fastapi import HTTPException, status

from app.models.hdp.user import Role, Permission
from app.core.base_response import NormalResponse
from app.domains.permission.schema import (
    PermissionItem, PermissionTreeItem, PermissionCreate, PermissionUpdate,
    RoleCreate, RoleUpdate,
)
from app.domains.permission.repository import (
    RoleDetailData, RoleListResponse, RoleDetailResponse,
    PermissionTreeResponse, PermissionListResponse,
)


class RolePermissionService:
    """角色与权限领域服务"""

    # =========================================================================
    # 1. 角色管理 (Role Management)
    # =========================================================================

    @staticmethod
    async def get_roles(session: AsyncSession) -> RoleListResponse:
        """获取系统所有角色明细列表 (包含关联合查的权限数据)"""
        stmt = select(Role).options(selectinload(Role.permissions)).order_by(Role.id.asc())
        result = await session.execute(stmt)
        roles = result.scalars().all()

        role_list: List[RoleDetailData] = []
        for r in roles:
            perm_items = [
                PermissionItem(
                    id=p.id,
                    parent_id=p.parent_id,
                    code=p.code,
                    name=p.name,
                    resource_type=p.resource_type,
                    path=p.path,
                    method=p.method,
                    created_at=p.created_at,
                )
                for p in r.permissions
            ]
            role_list.append(
                RoleDetailData(
                    id=r.id,
                    code=r.code,
                    name=r.name,
                    description=r.description,
                    is_system=r.is_system,
                    permissions=perm_items,
                    created_at=r.created_at,
                    updated_at=r.updated_at,
                )
            )

        return RoleListResponse(
            code=200,
            message="获取角色列表成功",
            data=role_list,
            success=True,
        )

    @staticmethod
    async def get_role_detail(session: AsyncSession, role_id: int) -> RoleDetailResponse:
        """获取单个角色详情"""
        stmt = select(Role).options(selectinload(Role.permissions)).where(Role.id == role_id)
        result = await session.execute(stmt)
        role = result.scalar_one_or_none()

        if not role:
            raise HTTPException(status_code=404, detail="指定角色不存在")

        perm_items = [
            PermissionItem(
                id=p.id,
                parent_id=p.parent_id,
                code=p.code,
                name=p.name,
                resource_type=p.resource_type,
                path=p.path,
                method=p.method,
                created_at=p.created_at,
            )
            for p in role.permissions
        ]

        detail = RoleDetailData(
            id=role.id,
            code=role.code,
            name=role.name,
            description=role.description,
            is_system=role.is_system,
            permissions=perm_items,
            created_at=role.created_at,
            updated_at=role.updated_at,
        )

        return RoleDetailResponse(
            code=200,
            message="获取角色详情成功",
            data=detail,
            success=True,
        )

    @staticmethod
    async def create_role(session: AsyncSession, data: RoleCreate) -> RoleDetailResponse:
        """新建角色"""
        # 检查角色 code 是否冲突
        stmt = select(Role).where(Role.code == data.code)
        if (await session.execute(stmt)).scalar_one_or_none():
            raise HTTPException(status_code=400, detail="角色标识 code 已存在")

        new_role = Role(
            code=data.code,
            name=data.name,
            description=data.description,
            is_system=False,
        )

        if data.permission_ids:
            perms_stmt = select(Permission).where(Permission.id.in_(data.permission_ids))
            perms_result = await session.execute(perms_stmt)
            new_role.permissions = list(perms_result.scalars().all())

        session.add(new_role)
        await session.flush()

        return await RolePermissionService.get_role_detail(session, new_role.id)

    @staticmethod
    async def update_role(
        session: AsyncSession, role_id: int, data: RoleUpdate
    ) -> RoleDetailResponse:
        """更新角色基本信息"""
        stmt = select(Role).where(Role.id == role_id)
        result = await session.execute(stmt)
        role = result.scalar_one_or_none()

        if not role:
            raise HTTPException(status_code=404, detail="指定角色不存在")

        if data.name:
            role.name = data.name
        if data.description is not None:
            role.description = data.description

        await session.flush()
        return await RolePermissionService.get_role_detail(session, role_id)

    @staticmethod
    async def delete_role(session: AsyncSession, role_id: int) -> NormalResponse:
        """删除角色 (系统内置角色禁止删除)"""
        stmt = select(Role).where(Role.id == role_id)
        result = await session.execute(stmt)
        role = result.scalar_one_or_none()

        if not role:
            raise HTTPException(status_code=404, detail="指定角色不存在")

        if role.is_system:
            raise HTTPException(status_code=400, detail="系统内置角色，无法删除")

        await session.delete(role)
        await session.flush()

        return NormalResponse(code=200, message="角色删除成功", data=None, success=True)

    @staticmethod
    async def assign_role_permissions(
        session: AsyncSession, role_id: int, permission_ids: List[int]
    ) -> NormalResponse:
        """为角色关联并更新权限集合"""
        stmt = select(Role).options(selectinload(Role.permissions)).where(Role.id == role_id)
        result = await session.execute(stmt)
        role = result.scalar_one_or_none()

        if not role:
            raise HTTPException(status_code=404, detail="指定角色不存在")

        perms_stmt = select(Permission).where(Permission.id.in_(permission_ids))
        perms_result = await session.execute(perms_stmt)
        role.permissions = list(perms_result.scalars().all())

        await session.flush()
        return NormalResponse(code=200, message="配置角色权限成功", data=None, success=True)

    # =========================================================================
    # 2. 权限树与菜单/API节点管理 (Permission Management)
    # =========================================================================

    @staticmethod
    async def get_permission_tree(session: AsyncSession) -> PermissionTreeResponse:
        """获取所有权限构成的层次树结构"""
        stmt = select(Permission).order_by(Permission.id.asc())
        result = await session.execute(stmt)
        permissions = result.scalars().all()

        # 映射为节点字典以递归构建树
        nodes: Dict[int, PermissionTreeItem] = {}
        root_items: List[PermissionTreeItem] = []

        for p in permissions:
            nodes[p.id] = PermissionTreeItem(
                id=p.id,
                parent_id=p.parent_id,
                code=p.code,
                name=p.name,
                resource_type=p.resource_type,
                path=p.path,
                method=p.method,
                created_at=p.created_at,
                children=[],
            )

        for p in permissions:
            item = nodes[p.id]
            if p.parent_id and p.parent_id in nodes:
                nodes[p.parent_id].children.append(item)
            else:
                root_items.append(item)

        return PermissionTreeResponse(
            code=200,
            message="获取权限树成功",
            data=root_items,
            success=True,
        )

    @staticmethod
    async def get_permission_list(session: AsyncSession) -> PermissionListResponse:
        """获取平铺列表形式的权限列表"""
        stmt = select(Permission).order_by(Permission.id.asc())
        result = await session.execute(stmt)
        permissions = result.scalars().all()

        items = [
            PermissionItem(
                id=p.id,
                parent_id=p.parent_id,
                code=p.code,
                name=p.name,
                resource_type=p.resource_type,
                path=p.path,
                method=p.method,
                created_at=p.created_at,
            )
            for p in permissions
        ]

        return PermissionListResponse(
            code=200,
            message="获取权限列表成功",
            data=items,
            success=True,
        )

    @staticmethod
    async def create_permission(
        session: AsyncSession, data: PermissionCreate
    ) -> NormalResponse:
        """创建新的权限节点"""
        stmt = select(Permission).where(Permission.code == data.code)
        if (await session.execute(stmt)).scalar_one_or_none():
            raise HTTPException(status_code=400, detail="权限标识 code 已存在")

        new_perm = Permission(
            parent_id=data.parent_id,
            code=data.code,
            name=data.name,
            resource_type=data.resource_type,
            path=data.path,
            method=data.method,
        )
        session.add(new_perm)
        await session.flush()

        return NormalResponse(code=200, message="创建权限节点成功", data={"id": new_perm.id}, success=True)

    @staticmethod
    async def delete_permission(session: AsyncSession, permission_id: int) -> NormalResponse:
        """删除指定的权限节点"""
        stmt = select(Permission).where(Permission.id == permission_id)
        result = await session.execute(stmt)
        perm = result.scalar_one_or_none()

        if not perm:
            raise HTTPException(status_code=404, detail="指定权限节点不存在")

        await session.delete(perm)
        await session.flush()

        return NormalResponse(code=200, message="权限节点删除成功", data=None, success=True)
