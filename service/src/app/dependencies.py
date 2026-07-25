from collections.abc import AsyncGenerator
from typing import Annotated, List, Callable
from dataclasses import dataclass

import jwt
from sqlalchemy import select
from sqlalchemy.orm import selectinload
from sqlalchemy.ext.asyncio import AsyncSession
from fastapi import Depends, HTTPException, status
from fastapi.security import OAuth2PasswordBearer
from jwt.exceptions import InvalidTokenError
from pydantic import ValidationError

from app.infrastructure.database.hdp.session import SessionLocal
from app.core import security, redis_util
from app.config import settings
from app.models.hdp.user import User, Role, Permission, RolePermissions

reusable_oauth2 = OAuth2PasswordBearer(
    tokenUrl=f"{settings.API_V1_STR}/login/access-token"
)


async def get_db() -> AsyncGenerator[AsyncSession, None]:
    async with SessionLocal() as session:
        async with session.begin():
            yield session


SessionDep = Annotated[AsyncSession, Depends(get_db)]
TokenDep = Annotated[str, Depends(reusable_oauth2)]


@dataclass
class TokenPayload:
    sub: str | None = None


async def get_current_user(session: SessionDep, token: TokenDep) -> User:
    """获取当前认证用户 (包含 Redis 缓存支持与角色数据预加载)"""
    try:
        payload = jwt.decode(
            token, settings.SECRET_KEY, algorithms=[security.ALGORITHM]
        )
        token_data = TokenPayload(**payload)
    except (InvalidTokenError, ValidationError):
        raise HTTPException(
            status_code=status.HTTP_403_FORBIDDEN,
            detail="Could not validate credentials",
        )

    if not token_data.sub:
        raise HTTPException(
            status_code=status.HTTP_403_FORBIDDEN,
            detail="Could not validate credentials",
        )

    user_id = int(token_data.sub) if token_data.sub.isdigit() else token_data.sub
    cache_key = f"user:cache:{token_data.sub}"

    # 1. 尝试从 Redis 缓存中获取用户信息
    try:
        cached_user_data = await redis_util.get_json(cache_key)
        if cached_user_data:
            roles = [
                Role(
                    id=r["id"],
                    code=r["code"],
                    name=r["name"],
                    description=r.get("description"),
                )
                for r in cached_user_data.get("roles", [])
            ]
            user = User(
                id=cached_user_data.get("id"),
                username=cached_user_data.get("username"),
                email=cached_user_data.get("email"),
                status=cached_user_data.get("status", 0),
                password_hash=cached_user_data.get("password_hash", ""),
                metadata_=cached_user_data.get("metadata_", {}),
                roles=roles,
            )
            if user.status != 0:
                raise HTTPException(status_code=400, detail="Inactive user")
            return user
    except HTTPException:
        raise
    except Exception:
        # Redis 异常降级，回源数据库
        pass

    # 2. Redis 未命中，回源数据库查询 (预加载角色关联)
    stmt = select(User).options(selectinload(User.roles)).where(User.id == user_id)
    result = await session.execute(stmt)
    user = result.scalar_one_or_none()

    if not user:
        raise HTTPException(status_code=404, detail="User not found")
    if user.status != 0:
        raise HTTPException(status_code=400, detail="Inactive user")

    # 3. 写入 Redis 缓存 (设置 TTL 300 秒)
    try:
        roles_data = [
            {"id": r.id, "code": r.code, "name": r.name, "description": r.description}
            for r in user.roles
        ]
        user_dict = {
            "id": user.id,
            "username": user.username,
            "email": user.email,
            "status": user.status,
            "password_hash": user.password_hash,
            "metadata_": user.metadata_ or {},
            "roles": roles_data,
        }
        await redis_util.set_json(cache_key, user_dict, expire=300)
    except Exception:
        pass

    return user


async def get_current_active_admin(
    current_user: Annotated[User, Depends(get_current_user)],
) -> User:
    """权限校验依赖：确保当前用户为管理员账号 (拥有 admin/superadmin 角色或用户名 admin)"""
    admin_role_codes = {"admin", "superadmin", "administrator"}
    user_role_codes = (
        {r.code.lower() for r in current_user.roles} if current_user.roles else set()
    )

    if current_user.username == "admin" or bool(user_role_codes & admin_role_codes):
        return current_user

    raise HTTPException(
        status_code=status.HTTP_403_FORBIDDEN,
        detail="权限不足：只有管理员才允许执行此操作 (Requires Admin Permission)",
    )


def require_permission(permission_code: str) -> Callable:
    """细粒度权限校验依赖工厂 (如 require_permission('user:delete'))"""

    async def permission_checker(
        session: SessionDep,
        current_user: Annotated[User, Depends(get_current_user)],
    ) -> User:
        admin_role_codes = {"admin", "superadmin", "administrator"}
        user_role_codes = (
            {r.code.lower() for r in current_user.roles}
            if current_user.roles
            else set()
        )

        # 超级管理员直接放行
        if current_user.username == "admin" or bool(user_role_codes & admin_role_codes):
            return current_user

        role_ids = [r.id for r in current_user.roles]
        if not role_ids:
            raise HTTPException(
                status_code=status.HTTP_403_FORBIDDEN,
                detail=f"权限不足，缺失权限: [{permission_code}]",
            )

        stmt = (
            select(Permission.code)
            .join(RolePermissions, RolePermissions.permission_id == Permission.id)
            .where(RolePermissions.role_id.in_(role_ids))
        )
        res = await session.execute(stmt)
        user_permissions = set(res.scalars().all())

        if permission_code not in user_permissions:
            raise HTTPException(
                status_code=status.HTTP_403_FORBIDDEN,
                detail=f"权限不足，缺失权限: [{permission_code}]",
            )

        return current_user

    return permission_checker


async def clear_user_cache(user_id: int | str):
    """清除指定用户的 Redis 缓存"""
    try:
        cache_key = f"user:cache:{user_id}"
        await redis_util.delete(cache_key)
    except Exception:
        pass
