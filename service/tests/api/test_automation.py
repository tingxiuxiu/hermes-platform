import asyncio
import uuid
from datetime import UTC, datetime

import httpx
import pytest

from app.config import settings
from app.main import app

API_PREFIX = settings.API_V1_STR
BASE_URL = "http://127.0.0.1:8080"


def _now_iso() -> str:
    return datetime.now(UTC).isoformat()


def _user_payload() -> dict[str, str]:
    unique = uuid.uuid4().hex[:12]
    return {
        "username": f"auto_user_{unique}",
        "password": "StrongPass123!",
        "email": f"auto_user_{unique}@example.com",
    }


async def _register_user(client: httpx.AsyncClient) -> dict[str, str]:
    payload = _user_payload()
    response = await client.post(f"{API_PREFIX}/register", json=payload)
    assert response.status_code == 200, response.text
    token = response.json()["data"]["access_token"]
    return {"Authorization": f"Bearer {token}"}


async def _create_execution(
    client: httpx.AsyncClient,
    service_headers: dict[str, str],
    *,
    build_uid: str | None = None,
) -> dict:
    payload = {
        "build_uid": build_uid or str(uuid.uuid4()),
        "job_name": "smoke-test-pipeline",
        "job_url": "https://jenkins.example.com/job/smoke-test-pipeline",
        "project_name": "demo-project",
        "software_name": "hermes",
        "software_version": "1.2.3",
        "start_time": _now_iso(),
        "planned_cases_count": 10,
    }
    resp = await client.post(
        f"{API_PREFIX}/automation/executions",
        json=payload,
        headers=service_headers,
    )
    assert resp.status_code == 200, resp.text
    return resp.json()["data"]


def test_automation_execution_success_flow_httpx():
    async def run_case():
        transport = httpx.ASGITransport(app=app)
        async with httpx.AsyncClient(transport=transport, base_url=BASE_URL) as client:
            auth_headers = await _register_user(client)
            service_headers = {"X-Service-Token": settings.AUTOMATION_SERVICE_TOKEN}

            build_uid = str(uuid.uuid4())
            creation = {
                "build_uid": build_uid,
                "job_name": "smoke-test-pipeline",
                "job_url": "https://jenkins.example.com/job/smoke-test-pipeline",
                "project_name": "demo-project",
                "software_name": "hermes",
                "software_version": "1.2.3",
                "start_time": _now_iso(),
                "planned_cases_count": 12,
            }

            create_resp = await client.post(
                f"{API_PREFIX}/automation/executions",
                json=creation,
                headers=service_headers,
            )
            assert create_resp.status_code == 200, create_resp.text
            execution = create_resp.json()["data"]
            assert execution["build_uid"] == build_uid
            assert execution["status"] == "running"

            update_resp = await client.patch(
                f"{API_PREFIX}/automation/executions/{build_uid}",
                json={
                    "status": "completed",
                    "end_time": _now_iso(),
                    "duration": 15.5,
                },
                headers=service_headers,
            )
            assert update_resp.status_code == 200, update_resp.text
            assert update_resp.json()["data"]["status"] == "completed"

            list_resp = await client.get(
                f"{API_PREFIX}/automation/executions",
                params={"page": 1, "page_size": 10, "job_name": "smoke-test"},
                headers=auth_headers,
            )
            assert list_resp.status_code == 200, list_resp.text
            payload = list_resp.json()["data"]
            assert payload["total"] >= 1
            assert any(item["build_uid"] == build_uid for item in payload["items"])

    asyncio.run(run_case())


def test_automation_create_execution_negative_httpx():
    async def run_case():
        transport = httpx.ASGITransport(app=app)
        async with httpx.AsyncClient(transport=transport, base_url=BASE_URL) as client:
            payload = {
                "build_uid": str(uuid.uuid4()),
                "job_name": "invalid-job",
                "job_url": "https://jenkins.example.com/job/invalid-job",
                "project_name": "demo-project",
                "software_name": "hermes",
                "software_version": "1.2.3",
                "start_time": _now_iso(),
                "planned_cases_count": 10,
            }

            # 缺少服务令牌
            missing_token = await client.post(
                f"{API_PREFIX}/automation/executions",
                json=payload,
            )
            assert missing_token.status_code == 401

            # 错误服务令牌
            bad_token = await client.post(
                f"{API_PREFIX}/automation/executions",
                json=payload,
                headers={"X-Service-Token": "wrong-token"},
            )
            assert bad_token.status_code == 401

    asyncio.run(run_case())


def test_automation_item_retry_history_httpx():
    async def run_case():
        transport = httpx.ASGITransport(app=app)
        async with httpx.AsyncClient(transport=transport, base_url=BASE_URL) as client:
            auth_headers = await _register_user(client)
            service_headers = {"X-Service-Token": settings.AUTOMATION_SERVICE_TOKEN}
            execution = await _create_execution(client, service_headers)
            build_uid = execution["build_uid"]

            case_key = f"tests/test_login.py::test_valid_login[{uuid.uuid4().hex[:8]}]"
            case_uid_1 = str(uuid.uuid4())
            case_uid_2 = str(uuid.uuid4())

            create_first = await client.post(
                f"{API_PREFIX}/automation/executions/{build_uid}/items",
                json={
                    "build_uid": build_uid,
                    "case_key": case_key,
                    "case_name": "test_valid_login",
                    "case_uid": case_uid_1,
                    "start_time": _now_iso(),
                },
                headers=service_headers,
            )
            assert create_first.status_code == 200, create_first.text
            first_item = create_first.json()["data"]
            assert first_item["attempt_number"] == 1

            finish_first = await client.patch(
                f"{API_PREFIX}/automation/executions/{build_uid}/items/{first_item['case_uid']}",
                json={
                    "status": "failed",
                    "end_time": _now_iso(),
                    "duration": 2.1,
                    "error_message": "assert 1 == 2",
                },
                headers=service_headers,
            )
            assert finish_first.status_code == 200, finish_first.text

            create_retry = await client.post(
                f"{API_PREFIX}/automation/executions/{build_uid}/items",
                json={
                    "build_uid": build_uid,
                    "case_key": case_key,
                    "case_name": "test_valid_login",
                    "case_uid": case_uid_2,
                    "start_time": _now_iso(),
                },
                headers=service_headers,
            )
            assert create_retry.status_code == 200, create_retry.text
            retry_item = create_retry.json()["data"]
            assert retry_item["attempt_number"] == 2

            attempts_resp = await client.get(
                f"{API_PREFIX}/automation/executions/{build_uid}/items/attempts",
                params={"case_key": case_key},
                headers=auth_headers,
            )
            assert attempts_resp.status_code == 200, attempts_resp.text
            attempts = attempts_resp.json()["data"]["items"]
            assert len(attempts) == 2
            assert attempts[0]["attempt_number"] == 1
            assert attempts[1]["attempt_number"] == 2

            history_resp = await client.get(
                f"{API_PREFIX}/automation/cases/history",
                params={"case_key": case_key},
                headers=auth_headers,
            )
            assert history_resp.status_code == 200, history_resp.text
            history_items = history_resp.json()["data"]["items"]
            assert len(history_items) >= 1

    asyncio.run(run_case())


def test_automation_steps_positive_and_boundary_httpx():
    async def run_case():
        transport = httpx.ASGITransport(app=app)
        async with httpx.AsyncClient(transport=transport, base_url=BASE_URL) as client:
            auth_headers = await _register_user(client)
            service_headers = {"X-Service-Token": settings.AUTOMATION_SERVICE_TOKEN}
            execution = await _create_execution(client, service_headers)
            build_uid = execution["build_uid"]

            item_resp = await client.post(
                f"{API_PREFIX}/automation/executions/{build_uid}/items",
                json={
                    "build_uid": build_uid,
                    "case_key": "tc-step-boundary",
                    "case_name": "test_step_boundary",
                    "case_uid": str(uuid.uuid4()),
                    "start_time": _now_iso(),
                },
                headers=service_headers,
            )
            assert item_resp.status_code == 200, item_resp.text
            item_id = item_resp.json()["data"]["id"]

            step_resp = await client.post(
                f"{API_PREFIX}/automation/executions/{build_uid}/items/{item_id}/steps",
                json={
                    "case_uid": item_resp.json()["data"]["case_uid"],
                    "step_path": "0",
                    "step_name": "setup",
                    "status": "passed",
                },
                headers=service_headers,
            )
            assert step_resp.status_code == 200, step_resp.text

            list_steps = await client.get(
                f"{API_PREFIX}/automation/items/{item_id}/steps",
                headers=auth_headers,
            )
            assert list_steps.status_code == 200, list_steps.text
            assert len(list_steps.json()["data"]) >= 1

            invalid_page = await client.get(
                f"{API_PREFIX}/automation/executions",
                params={"page": 0, "page_size": 10},
                headers=auth_headers,
            )
            assert invalid_page.status_code == 422

            invalid_page_size = await client.get(
                f"{API_PREFIX}/automation/executions",
                params={"page": 1, "page_size": 101},
                headers=auth_headers,
            )
            assert invalid_page_size.status_code == 422

    asyncio.run(run_case())


@pytest.mark.parametrize(
    "params",
    [
        {"page": 0},
        {"page_size": 0},
        {"page_size": 101},
    ],
)
def test_automation_query_boundary_validation_httpx(params):
    async def run_case():
        transport = httpx.ASGITransport(app=app)
        async with httpx.AsyncClient(transport=transport, base_url=BASE_URL) as client:
            auth_headers = await _register_user(client)
            response = await client.get(
                f"{API_PREFIX}/automation/executions",
                params=params,
                headers=auth_headers,
            )
            assert response.status_code == 422, response.text

    asyncio.run(run_case())
