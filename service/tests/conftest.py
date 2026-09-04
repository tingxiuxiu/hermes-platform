import asyncio
import sys
import uuid

import pytest
from fastapi.testclient import TestClient

from app.config import settings
from app.main import app

# psycopg 的异步驱动在 Windows 上要求使用 SelectorEventLoop，
# 而 anyio/TestClient 默认可能使用 ProactorEventLoop，需要在创建事件循环前显式切换策略。
if sys.platform == "win32":
    asyncio.set_event_loop_policy(asyncio.WindowsSelectorEventLoopPolicy())

API_PREFIX = settings.API_V1_STR


@pytest.fixture(scope="session")
def client() -> TestClient:
    """基于真实配置 (.env 中的 Postgres/Redis) 启动的同步测试客户端"""
    with TestClient(app) as c:
        yield c


@pytest.fixture()
def auth_headers(client: TestClient) -> dict[str, str]:
    """注册一个随机新用户并返回其鉴权 Header，保证测试之间数据互不冲突"""
    unique = uuid.uuid4().hex[:12]
    payload = {
        "username": f"test_{unique}",
        "password": "Test@123456",
        "email": f"test_{unique}@example.com",
    }
    resp = client.post(f"{API_PREFIX}/register", json=payload)
    assert resp.status_code == 200, resp.text
    token = resp.json()["data"]["access_token"]
    return {"Authorization": f"Bearer {token}"}


@pytest.fixture()
def service_headers() -> dict[str, str]:
    """自动化用例上报接口 (pytest/CI Runner) 使用的独立服务令牌 Header，与用户登录态区分"""
    return {"X-Service-Token": settings.AUTOMATION_SERVICE_TOKEN}
