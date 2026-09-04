from typing import Annotated

from fastapi import APIRouter, Depends, Request
from fastapi.security import OAuth2PasswordRequestForm

from app.dependencies import SessionDep, get_current_user
from .models import User
from .schemas import UserCreate, UserLogin
from .constants import (
    UserLoginResponse,
    UserLogoutResponse,
)
from .service import login as user_login, register_and_login, logout as user_logout

auth_router = APIRouter(tags=["认证管理"])


@auth_router.post(
    "/login/access-token",
    response_model=UserLoginResponse,
    summary="OAuth2 表单登录 (支持 Swagger Authorize)",
)
async def login_access_token_endpoint(
    session: SessionDep,
    request: Request,
    form_data: Annotated[OAuth2PasswordRequestForm, Depends()],
) -> UserLoginResponse:
    """OAuth2 兼容的标准表单登录接口"""
    client_ip = request.client.host if request.client else None
    login_data = UserLogin(username=form_data.username, password=form_data.password)
    return await user_login(
        session=session,
        login_data=login_data,
        client_ip=client_ip,
    )


@auth_router.post(
    "/login",
    response_model=UserLoginResponse,
    summary="JSON Body 账号密码登录",
)
async def login_json_endpoint(
    session: SessionDep,
    request: Request,
    login_data: UserLogin,
) -> UserLoginResponse:
    """JSON 格式账号密码登录接口"""
    client_ip = request.client.host if request.client else None
    return await user_login(
        session=session,
        login_data=login_data,
        client_ip=client_ip,
    )


@auth_router.post(
    "/register",
    response_model=UserLoginResponse,
    summary="注册新用户并登录",
)
async def register_endpoint(
    session: SessionDep,
    request: Request,
    user_data: UserCreate,
) -> UserLoginResponse:
    """创建新用户账号并立即登录"""
    client_ip = request.client.host if request.client else None
    return await register_and_login(
        session=session,
        data=user_data,
        client_ip=client_ip,
    )


@auth_router.post(
    "/logout",
    response_model=UserLogoutResponse,
    summary="用户登出",
)
async def logout_endpoint(
    current_user: Annotated[User, Depends(get_current_user)],
) -> UserLogoutResponse:
    """退出登录并清理用户 Redis 缓存"""
    return await user_logout(user_id=current_user.id)
