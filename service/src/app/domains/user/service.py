from datetime import datetime, UTC, timedelta
from typing import Optional, List

from sqlalchemy import func, or_
from sqlalchemy.future import select
from sqlalchemy.orm import selectinload
from sqlalchemy.ext.asyncio import AsyncSession

from app.config import settings
from app.core import security
from app.dependencies import clear_user_cache
from app.models.hdp.user import User, Role, UserRoles
from app.domains.user.schema import (
    UserLogin,
    UserCreate,
    UserListQuery,
    UserUpdate,
    UserItem,
    RoleItem,
    UserDetailData,
    UserMetadata,
)
from app.domains.user.repository import (
    UserLoginData,
    UserLoginResponse,
    UserLogoutResponse,
    UserCreateResponse,
    UserListResponse,
    UserListData,
    UserDetailResponse,
    UserUpdateResponse,
    UserRoleAssignResponse,
    UserStatusUpdateResponse,
    UserResetPasswordResponse,
    UserChangePasswordResponse,
    UserBatchResponse,
    RoleOptionResponse,
)
from app.domains.user.exceptions import (
    InvalidCredentialsException,
    UserInactiveException,
    UserNotFoundException,
    UserAlreadyExistsException,
)


class UserService:
    """用户领域服务 (包含登录、认证、用户 CRUD、角色配置、状态变更及缓存清理等业务逻辑)"""

    # =========================================================================
    # 1. 认证与登录登出 (Authentication)
    # =========================================================================

    @staticmethod
    async def authenticate_user(
        session: AsyncSession, username: str, password: str
    ) -> User:
        """根据用户名或邮箱与密码认证用户身份"""
        stmt = (
            select(User)
            .options(selectinload(User.roles))
            .where(or_(User.username == username, User.email == username))
        )
        result = await session.execute(stmt)
        user = result.scalar_one_or_none()

        if not user:
            raise InvalidCredentialsException()

        is_valid, updated_hash = security.verify_password(password, user.password_hash)
        if not is_valid:
            raise InvalidCredentialsException()

        if updated_hash:
            user.password_hash = updated_hash
            await session.flush()

        if user.status != 0:
            raise UserInactiveException()

        return user

    @staticmethod
    async def login(
        session: AsyncSession,
        login_data: UserLogin,
        client_ip: Optional[str] = None,
    ) -> UserLoginResponse:
        """用户登录业务主流程"""
        user = await UserService.authenticate_user(
            session=session,
            username=login_data.username,
            password=login_data.password,
        )

        user.last_login_at = datetime.now(UTC)
        if client_ip:
            user.last_login_ip = client_ip
        await session.flush()

        expires_delta = timedelta(seconds=settings.ACCESS_TOKEN_EXPIRE_SECONDS)
        access_token = security.create_access_token(
            subject=user.id, expires_delta=expires_delta
        )

        role_items: List[RoleItem] = [
            RoleItem(
                id=role.id,
                code=role.code,
                name=role.name,
                description=role.description,
            )
            for role in user.roles
        ]

        user_item = UserItem(
            id=user.id,
            username=user.username,
            email=user.email,
            status=user.status,
            roles=role_items,
            last_login_at=user.last_login_at,
            created_at=user.created_at if hasattr(user, "created_at") else None,
        )

        login_result = UserLoginData(
            access_token=access_token,
            token_type="bearer",
            expires_in=settings.ACCESS_TOKEN_EXPIRE_SECONDS,
            user=user_item,
        )

        return UserLoginResponse(
            code=200,
            message="登录成功",
            data=login_result,
            success=True,
        )

    @staticmethod
    async def logout(user_id: int) -> UserLogoutResponse:
        """用户登出逻辑 (清理 Redis 用户缓存)"""
        await clear_user_cache(user_id)
        return UserLogoutResponse(
            code=200,
            message="登出成功",
            data=None,
            success=True,
        )

    # =========================================================================
    # 2. 用户 CRUD 与列表查询 (User Management CRUD)
    # =========================================================================

    @staticmethod
    async def get_user_list(
        session: AsyncSession, query: UserListQuery
    ) -> UserListResponse:
        """获取用户列表 (支持模糊搜索、状态筛选、角色筛选及分页)"""
        stmt = select(User).options(selectinload(User.roles)).where(User.status != 2)

        if query.username:
            stmt = stmt.where(User.username.ilike(f"%{query.username}%"))
        if query.email:
            stmt = stmt.where(User.email.ilike(f"%{query.email}%"))
        if query.status is not None:
            stmt = stmt.where(User.status == query.status)
        if query.role_id is not None:
            stmt = stmt.join(User.roles).where(Role.id == query.role_id)

        # 计算符合条件的总条数
        count_stmt = select(func.count()).select_from(stmt.subquery())
        total_result = await session.execute(count_stmt)
        total = total_result.scalar_one() or 0

        # 分页逻辑
        offset = (query.page - 1) * query.page_size
        stmt = stmt.offset(offset).limit(query.page_size).order_by(User.id.desc())

        result = await session.execute(stmt)
        users = result.scalars().unique().all()

        items: List[UserItem] = []
        for u in users:
            role_items = [
                RoleItem(id=r.id, code=r.code, name=r.name, description=r.description)
                for r in u.roles
            ]
            items.append(
                UserItem(
                    id=u.id,
                    username=u.username,
                    email=u.email,
                    status=u.status,
                    roles=role_items,
                    last_login_at=u.last_login_at,
                    created_at=u.created_at if hasattr(u, "created_at") else None,
                )
            )

        list_data = UserListData(
            total=total,
            page=query.page,
            page_size=query.page_size,
            items=items,
        )

        return UserListResponse(
            code=200,
            message="获取用户列表成功",
            data=list_data,
            success=True,
        )

    @staticmethod
    async def _create_user_entity(session: AsyncSession, data: UserCreate) -> User:
        """Create a new user entity and flush it to the DB."""
        conditions = [User.username == data.username]
        if data.email:
            conditions.append(User.email == data.email)

        exist_stmt = select(User).where(or_(*conditions))
        exist_result = await session.execute(exist_stmt)
        if exist_result.scalar_one_or_none():
            raise UserAlreadyExistsException("用户名或邮箱已被注册")

        hashed_password = security.get_password_hash(data.password)
        metadata_dict = (
            data.metadata.model_dump(exclude_unset=True) if data.metadata else {}
        )

        new_user = User(
            username=data.username,
            email=data.email,
            password_hash=hashed_password,
            status=0,
            metadata_=metadata_dict,
        )

        assigned_roles: List[Role] = []
        if data.role_ids:
            roles_stmt = select(Role).where(Role.id.in_(data.role_ids))
            roles_result = await session.execute(roles_stmt)
            assigned_roles = list(roles_result.scalars().all())
            new_user.roles = assigned_roles

        session.add(new_user)
        await session.flush()
        return new_user

    @staticmethod
    async def create_user(
        session: AsyncSession, data: UserCreate
    ) -> UserCreateResponse:
        """新建/注册用户"""
        new_user = await UserService._create_user_entity(session=session, data=data)

        role_items = [
            RoleItem(id=r.id, code=r.code, name=r.name, description=r.description)
            for r in new_user.roles
        ]
        created_item = UserItem(
            id=new_user.id,
            username=new_user.username,
            email=new_user.email,
            status=new_user.status,
            roles=role_items,
            last_login_at=new_user.last_login_at,
            created_at=new_user.created_at if hasattr(new_user, "created_at") else None,
        )

        return UserCreateResponse(
            code=200,
            message="创建用户成功",
            data=created_item,
            success=True,
        )

    @staticmethod
    async def register_and_login(
        session: AsyncSession, data: UserCreate, client_ip: Optional[str] = None
    ) -> UserLoginResponse:
        """创建新用户并立即登录，避免自动登录发生跨请求竞态。"""
        await UserService._create_user_entity(session=session, data=data)
        login_data = UserLogin(username=data.username, password=data.password)
        return await UserService.login(
            session=session,
            login_data=login_data,
            client_ip=client_ip,
        )

    @staticmethod
    async def get_user_detail(
        session: AsyncSession, user_id: int
    ) -> UserDetailResponse:
        """获取用户详情"""
        stmt = (
            select(User)
            .options(selectinload(User.roles))
            .where(User.id == user_id, User.status != 2)
        )
        result = await session.execute(stmt)
        user = result.scalar_one_or_none()

        if not user:
            raise UserNotFoundException()

        role_items = [
            RoleItem(id=r.id, code=r.code, name=r.name, description=r.description)
            for r in user.roles
        ]

        metadata_obj = (
            UserMetadata(**(user.metadata_ or {})) if user.metadata_ else None
        )

        detail_data = UserDetailData(
            id=user.id,
            username=user.username,
            email=user.email,
            status=user.status,
            roles=role_items,
            last_login_at=user.last_login_at,
            created_at=user.created_at if hasattr(user, "created_at") else None,
            last_login_ip=user.last_login_ip,
            metadata=metadata_obj,
            updated_at=user.updated_at if hasattr(user, "updated_at") else None,
        )

        return UserDetailResponse(
            code=200,
            message="获取用户详情成功",
            data=detail_data,
            success=True,
        )

    @staticmethod
    async def update_user(
        session: AsyncSession, user_id: int, data: UserUpdate
    ) -> UserUpdateResponse:
        """更新用户基本信息"""
        stmt = select(User).where(User.id == user_id, User.status != 2)
        result = await session.execute(stmt)
        user = result.scalar_one_or_none()

        if not user:
            raise UserNotFoundException()

        if data.username and data.username != user.username:
            # 检查用户名冲突
            dup_stmt = select(User).where(
                User.username == data.username, User.id != user_id
            )
            if (await session.execute(dup_stmt)).scalar_one_or_none():
                raise UserAlreadyExistsException("用户名已被占用")
            user.username = data.username

        if data.email and data.email != user.email:
            dup_stmt = select(User).where(User.email == data.email, User.id != user_id)
            if (await session.execute(dup_stmt)).scalar_one_or_none():
                raise UserAlreadyExistsException("邮箱已被占用")
            user.email = data.email

        if data.metadata:
            current_meta = dict(user.metadata_ or {})
            current_meta.update(data.metadata.model_dump(exclude_unset=True))
            user.metadata_ = current_meta

        await session.flush()
        await clear_user_cache(user_id)

        return UserUpdateResponse(
            code=200,
            message="更新用户信息成功",
            data=None,
            success=True,
        )

    # =========================================================================
    # 3. 角色分配与状态控制 (Roles & Status)
    # =========================================================================

    @staticmethod
    async def assign_roles(
        session: AsyncSession, user_id: int, role_ids: List[int]
    ) -> UserRoleAssignResponse:
        """为用户分配角色"""
        stmt = (
            select(User)
            .options(selectinload(User.roles))
            .where(User.id == user_id, User.status != 2)
        )
        result = await session.execute(stmt)
        user = result.scalar_one_or_none()

        if not user:
            raise UserNotFoundException()

        roles_stmt = select(Role).where(Role.id.in_(role_ids))
        roles_result = await session.execute(roles_stmt)
        roles = list(roles_result.scalars().all())

        user.roles = roles
        await session.flush()
        await clear_user_cache(user_id)

        return UserRoleAssignResponse(
            code=200,
            message="分配用户角色成功",
            data=None,
            success=True,
        )

    @staticmethod
    async def update_status(
        session: AsyncSession, user_id: int, status: int
    ) -> UserStatusUpdateResponse:
        """更新用户状态 (0: 启用, 1: 禁用)"""
        stmt = select(User).where(User.id == user_id, User.status != 2)
        result = await session.execute(stmt)
        user = result.scalar_one_or_none()

        if not user:
            raise UserNotFoundException()

        user.status = status
        await session.flush()
        await clear_user_cache(user_id)

        status_str = "启用" if status == 0 else "禁用"
        return UserStatusUpdateResponse(
            code=200,
            message=f"用户账号已成功{status_str}",
            data=None,
            success=True,
        )

    # =========================================================================
    # 4. 密码重置与修改 (Password Management)
    # =========================================================================

    @staticmethod
    async def reset_password(
        session: AsyncSession, user_id: int, new_password: str
    ) -> UserResetPasswordResponse:
        """管理员重置用户密码"""
        stmt = select(User).where(User.id == user_id, User.status != 2)
        result = await session.execute(stmt)
        user = result.scalar_one_or_none()

        if not user:
            raise UserNotFoundException()

        user.password_hash = security.get_password_hash(new_password)
        await session.flush()
        await clear_user_cache(user_id)

        return UserResetPasswordResponse(
            code=200,
            message="重置密码成功",
            data=None,
            success=True,
        )

    @staticmethod
    async def change_password(
        session: AsyncSession, user: User, old_password: str, new_password: str
    ) -> UserChangePasswordResponse:
        """用户个人修改密码"""
        is_valid, _ = security.verify_password(old_password, user.password_hash)
        if not is_valid:
            raise InvalidCredentialsException("原密码输入错误")

        user.password_hash = security.get_password_hash(new_password)
        await session.flush()
        await clear_user_cache(user.id)

        return UserChangePasswordResponse(
            code=200,
            message="修改密码成功",
            data=None,
            success=True,
        )

    # =========================================================================
    # 5. 批量操作与系统角色辅助接口 (Batch & Helper Operations)
    # =========================================================================

    @staticmethod
    async def batch_update_status(
        session: AsyncSession, user_ids: List[int], status: int
    ) -> UserBatchResponse:
        """批量启用/禁用用户"""
        stmt = select(User).where(User.id.in_(user_ids), User.status != 2)
        result = await session.execute(stmt)
        users = result.scalars().all()

        for u in users:
            u.status = status
            await clear_user_cache(u.id)

        await session.flush()
        return UserBatchResponse(
            code=200,
            message=f"批量更新 {len(users)} 个用户状态成功",
            data=None,
            success=True,
        )

    @staticmethod
    async def batch_delete_users(
        session: AsyncSession, user_ids: List[int]
    ) -> UserBatchResponse:
        """批量软删除用户"""
        stmt = select(User).where(User.id.in_(user_ids), User.status != 2)
        result = await session.execute(stmt)
        users = result.scalars().all()

        for u in users:
            u.status = 2  # 软删除标志
            await clear_user_cache(u.id)

        await session.flush()
        return UserBatchResponse(
            code=200,
            message=f"批量删除 {len(users)} 个用户成功",
            data=None,
            success=True,
        )

    @staticmethod
    async def get_role_options(session: AsyncSession) -> RoleOptionResponse:
        """获取所有可用系统角色列表 (用于配置角色时的下拉选择)"""
        stmt = select(Role).order_by(Role.id.asc())
        result = await session.execute(stmt)
        roles = result.scalars().all()

        items = [
            RoleItem(id=r.id, code=r.code, name=r.name, description=r.description)
            for r in roles
        ]
        return RoleOptionResponse(
            code=200,
            message="获取角色选择列表成功",
            data=items,
            success=True,
        )
