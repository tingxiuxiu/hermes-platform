import asyncio
import uuid

import httpx

from app.config import settings
from app.main import app

API_PREFIX = settings.API_V1_STR
BASE_URL = "http://127.0.0.1:8080"


def _unique_suffix() -> str:
    return uuid.uuid4().hex[:12]


def _user_payload() -> dict[str, str]:
    suffix = _unique_suffix()
    return {
        "username": f"auth_user_{suffix}",
        "password": "StrongPass123!",
        "email": f"auth_user_{suffix}@example.com",
    }


async def _request(method: str, path: str, **kwargs):
    transport = httpx.ASGITransport(app=app)
    async with httpx.AsyncClient(transport=transport, base_url=BASE_URL) as client:
        return await client.request(method, path, **kwargs)


def test_register_success_and_auto_login():
    payload = _user_payload()

    response = asyncio.run(_request("POST", f"{API_PREFIX}/register", json=payload))

    assert response.status_code == 200, response.text
    body = response.json()
    assert body["success"] is True
    assert body["message"] == "登录成功"
    assert body["data"]["access_token"]
    assert body["data"]["user"]["username"] == payload["username"]
    assert body["data"]["user"]["email"] == payload["email"]


def test_login_access_token_success_and_invalid_credentials():
    payload = _user_payload()
    register_response = asyncio.run(
        _request("POST", f"{API_PREFIX}/register", json=payload)
    )
    assert register_response.status_code == 200, register_response.text

    form_data = {
        "username": payload["username"],
        "password": payload["password"],
    }
    login_response = asyncio.run(
        _request(
            "POST",
            f"{API_PREFIX}/login/access-token",
            data=form_data,
        )
    )

    assert login_response.status_code == 200, login_response.text
    body = login_response.json()
    assert body["success"] is True
    assert body["data"]["token_type"] == "bearer"
    assert body["data"]["user"]["username"] == payload["username"]

    bad_response = asyncio.run(
        _request(
            "POST",
            f"{API_PREFIX}/login/access-token",
            data={"username": payload["username"], "password": "WrongPass123!"},
        )
    )
    assert bad_response.status_code == 400, bad_response.text
    assert "detail" in bad_response.json()


def test_login_json_success_and_duplicate_register_rejected():
    payload = _user_payload()
    register_response = asyncio.run(
        _request("POST", f"{API_PREFIX}/register", json=payload)
    )
    assert register_response.status_code == 200, register_response.text

    login_response = asyncio.run(
        _request(
            "POST",
            f"{API_PREFIX}/login",
            json={
                "username": payload["username"],
                "password": payload["password"],
            },
        )
    )

    assert login_response.status_code == 200, login_response.text
    assert login_response.json()["data"]["access_token"]

    duplicate_response = asyncio.run(
        _request("POST", f"{API_PREFIX}/register", json=payload)
    )
    assert duplicate_response.status_code == 400, duplicate_response.text
    assert duplicate_response.json()["detail"] == "用户名或邮箱已被注册"


def test_logout_success_requires_valid_bearer_token():
    payload = _user_payload()
    register_response = asyncio.run(
        _request("POST", f"{API_PREFIX}/register", json=payload)
    )
    assert register_response.status_code == 200, register_response.text
    token = register_response.json()["data"]["access_token"]

    logout_response = asyncio.run(
        _request(
            "POST",
            f"{API_PREFIX}/logout",
            headers={"Authorization": f"Bearer {token}"},
        )
    )

    assert logout_response.status_code == 200, logout_response.text
    body = logout_response.json()
    assert body["success"] is True
    assert body["message"] == "登出成功"

    missing_token_response = asyncio.run(_request("POST", f"{API_PREFIX}/logout"))
    assert missing_token_response.status_code == 401, missing_token_response.text
    assert "detail" in missing_token_response.json()


def test_async_httpx_login_round_trip():
    payload = _user_payload()

    async def run_case():
        async with httpx.AsyncClient(
            transport=httpx.ASGITransport(app=app), base_url=BASE_URL
        ) as client:
            register_response = await client.post(
                f"{API_PREFIX}/register",
                json=payload,
            )
            assert register_response.status_code == 200, register_response.text

            login_response = await client.post(
                f"{API_PREFIX}/login",
                json={
                    "username": payload["username"],
                    "password": payload["password"],
                },
            )
            assert login_response.status_code == 200, login_response.text
            assert (
                login_response.json()["data"]["user"]["username"] == payload["username"]
            )

    asyncio.run(run_case())
