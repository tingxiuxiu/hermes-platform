from datetime import datetime

from typing import Optional, List, Dict, Any
from sqlalchemy import (
    BigInteger,
    String,
    Text,
    SmallInteger,
    Boolean,
    TIMESTAMP,
    ForeignKey,
    CheckConstraint,
    Index,
)
from sqlalchemy.dialects.postgresql import JSONB
from sqlalchemy.orm import Mapped, mapped_column, relationship
from sqlalchemy.sql import func

from app.models import Base


# ==========================================
# 关联表定义 (多对多)
# ==========================================
class UserRoles(Base):
    __tablename__ = "user_roles"

    id: Mapped[int] = mapped_column(BigInteger, primary_key=True, autoincrement=True)
    user_id: Mapped[int] = mapped_column(
        BigInteger, ForeignKey("users.id", ondelete="CASCADE")
    )
    role_id: Mapped[int] = mapped_column(
        BigInteger, ForeignKey("roles.id", ondelete="CASCADE")
    )

    __table_args__ = (Index("idx_user_roles_role_id", "role_id"),)


class RolePermissions(Base):
    __tablename__ = "role_permissions"

    id: Mapped[int] = mapped_column(BigInteger, primary_key=True, autoincrement=True)
    role_id: Mapped[int] = mapped_column(
        BigInteger, ForeignKey("roles.id", ondelete="CASCADE")
    )
    permission_id: Mapped[int] = mapped_column(
        BigInteger, ForeignKey("permissions.id", ondelete="CASCADE")
    )

    __table_args__ = (Index("idx_role_permissions_perm_id", "permission_id"),)


# ==========================================
# 实体表定义
# ==========================================
class User(Base):
    __tablename__ = "users"

    id: Mapped[int] = mapped_column(
        BigInteger,
        primary_key=True,
        index=True,
        autoincrement=True,
        comment="Primary Key",
    )
    username: Mapped[str] = mapped_column(String(64), nullable=False, unique=True)
    email: Mapped[Optional[str]] = mapped_column(
        String(255), nullable=False, comment="User Email", unique=True
    )
    # 登录密码（已哈希存储）
    password_hash: Mapped[str] = mapped_column(
        Text, nullable=False, comment="Password Hash"
    )
    # 0=Active, -1=Inactive, -2=Deleted
    status: Mapped[int] = mapped_column(
        SmallInteger, nullable=False, default=0, comment="Status"
    )
    # 扩展字段，用于存储其他用户信息（如手机号，头像，昵称，偏好设置等）
    metadata_: Mapped[Dict[str, Any]] = mapped_column(
        "metadata", JSONB, default=dict, server_default="{}", nullable=False
    )
    last_login_at: Mapped[Optional[datetime]] = mapped_column(
        TIMESTAMP(timezone=True), nullable=True, comment="Last Login Time"
    )
    last_login_ip: Mapped[Optional[str]] = mapped_column(
        String(45), nullable=True, comment="Last Login IP"
    )
    # 关系映射：多对多绑定 Role
    roles: Mapped[List["Role"]] = relationship(
        "Role", secondary=UserRoles.__table__, back_populates="users"
    )

    __table_args__ = (CheckConstraint("status IN (0, 1, 2)", name="chk_users_status"),)


class Role(Base):
    __tablename__ = "roles"

    id: Mapped[int] = mapped_column(BigInteger, primary_key=True, autoincrement=True)
    # 角色标识，如 "admin", "editor"
    code: Mapped[str] = mapped_column(String(64), nullable=False, unique=True)
    # 角色名称，如 "管理员", "编辑"
    name: Mapped[str] = mapped_column(String(64), nullable=False)
    description: Mapped[Optional[str]] = mapped_column(Text, nullable=True)
    # 是否为内置系统角色（内置角色不可删除
    is_system: Mapped[bool] = mapped_column(
        Boolean, default=False, server_default="false", nullable=False
    )

    created_at: Mapped[datetime] = mapped_column(
        TIMESTAMP(timezone=True), server_default=func.now(), nullable=False
    )
    updated_at: Mapped[datetime] = mapped_column(
        TIMESTAMP(timezone=True),
        server_default=func.now(),
        onupdate=func.now(),
        nullable=False,
    )

    # 关系映射
    users: Mapped[List["User"]] = relationship(
        "User", secondary=UserRoles.__table__, back_populates="roles"
    )
    permissions: Mapped[List["Permission"]] = relationship(
        "Permission", secondary=RolePermissions.__table__, back_populates="roles"
    )


class Permission(Base):
    __tablename__ = "permissions"

    id: Mapped[int] = mapped_column(BigInteger, primary_key=True, autoincrement=True)
    # 支持树形权限层级
    parent_id: Mapped[Optional[int]] = mapped_column(BigInteger, nullable=True)
    # 权限标识，如 "user:create", "order:read"
    code: Mapped[str] = mapped_column(String(128), nullable=False, unique=True)
    # 权限名称，如 "创建用户"
    name: Mapped[str] = mapped_column(String(64), nullable=False)
    # 1-菜单, 2-按钮/功能, 3-API接口
    resource_type: Mapped[int] = mapped_column(
        SmallInteger, default=1, server_default="1", nullable=False
    )
    # 前端路由 path 或后端 API 路径 (如 /api/v1/users)
    path: Mapped[Optional[str]] = mapped_column(Text, nullable=True)
    # 请求方式 GET, POST, PUT, DELETE (当类型为 API 时有用)
    method: Mapped[Optional[str]] = mapped_column(String(16), nullable=True)

    created_at: Mapped[datetime] = mapped_column(
        TIMESTAMP(timezone=True), server_default=func.now(), nullable=False
    )

    # 自引用树形关系 & 角色映射
    children: Mapped[List["Permission"]] = relationship(
        "Permission", backref="parent", remote_side=[id]
    )
    roles: Mapped[List["Role"]] = relationship(
        "Role", secondary=RolePermissions.__table__, back_populates="permissions"
    )

    __table_args__ = (
        CheckConstraint("resource_type IN (1, 2, 3)", name="chk_permissions_type"),
    )
