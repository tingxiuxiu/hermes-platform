import uuid
from datetime import UTC, datetime

from fastapi.testclient import TestClient

from app.config import settings

API_PREFIX = settings.API_V1_STR


def _now_iso() -> str:
    return datetime.now(UTC).isoformat()


def _create_execution(
    client: TestClient,
    service_headers: dict[str, str],
    *,
    pre_cases_count: int = 3,
) -> dict:
    unique = uuid.uuid4().hex[:8]
    resp = client.post(
        f"{API_PREFIX}/automation/executions",
        json={
            "job_id": 9000,
            "job_uid": str(uuid.uuid4()),
            "job_name": f"dashboard-job-{unique}",
            "job_url": f"https://jenkins.example.com/job/dashboard-job-{unique}",
            "start_at": _now_iso(),
            "pre_cases_count": pre_cases_count,
        },
        headers=service_headers,
    )
    assert resp.status_code == 200, resp.text
    return resp.json()["data"]


def _create_item(
    client: TestClient,
    service_headers: dict[str, str],
    execution_id: int,
    case_name: str,
) -> dict:
    case_key = f"tests/test_dashboard.py::{case_name}[{uuid.uuid4().hex[:8]}]"
    resp = client.post(
        f"{API_PREFIX}/automation/executions/{execution_id}/items",
        json={
            "case_key": case_key,
            "case_name": case_name,
            "start_at": _now_iso(),
        },
        headers=service_headers,
    )
    assert resp.status_code == 200, resp.text
    return resp.json()["data"]


class TestAutomationDashboardApi:
    def test_dashboard_overview_returns_summary_and_running_cards(
        self,
        client: TestClient,
        service_headers: dict[str, str],
        auth_headers: dict[str, str],
    ):
        execution = _create_execution(client, service_headers)

        success_item = _create_item(
            client, service_headers, execution["id"], "test_success_case"
        )
        update_resp = client.patch(
            f"{API_PREFIX}/automation/executions/{execution['id']}/items/{success_item['id']}",
            json={"status": "success", "end_at": _now_iso(), "duration": 1.3},
            headers=service_headers,
        )
        assert update_resp.status_code == 200, update_resp.text

        _create_item(client, service_headers, execution["id"], "test_running_case")

        overview_resp = client.get(
            f"{API_PREFIX}/automation/dashboard/overview", headers=auth_headers
        )
        assert overview_resp.status_code == 200, overview_resp.text

        data = overview_resp.json()["data"]
        assert data["summary"]["running_executions_count"] >= 1
        assert len(data["trends"]) >= 1

        running_card = next(
            item
            for item in data["running_executions"]
            if item["execution_id"] == execution["id"]
        )
        assert running_card["pre_cases_count"] == 3
        assert running_card["completed_cases_count"] == 1
        assert running_card["success_cases_count"] == 1
        assert running_card["failure_cases_count"] == 0
        assert running_card["running_cases_count"] == 1
        assert running_card["progress_percent"] > 0
        assert "test_running_case" in running_card["running_case_names"]

    def test_dashboard_overview_supports_job_filters(
        self,
        client: TestClient,
        service_headers: dict[str, str],
        auth_headers: dict[str, str],
    ):
        matched_execution = _create_execution(
            client, service_headers, pre_cases_count=2
        )
        matched_item = _create_item(
            client, service_headers, matched_execution["id"], "test_matched_case"
        )
        matched_update_resp = client.patch(
            f"{API_PREFIX}/automation/executions/{matched_execution['id']}/items/{matched_item['id']}",
            json={"status": "success", "end_at": _now_iso(), "duration": 0.9},
            headers=service_headers,
        )
        assert matched_update_resp.status_code == 200, matched_update_resp.text

        other_execution = _create_execution(client, service_headers, pre_cases_count=1)
        other_item = _create_item(
            client, service_headers, other_execution["id"], "test_other_case"
        )
        other_update_resp = client.patch(
            f"{API_PREFIX}/automation/executions/{other_execution['id']}/items/{other_item['id']}",
            json={"status": "failure", "end_at": _now_iso(), "duration": 1.2},
            headers=service_headers,
        )
        assert other_update_resp.status_code == 200, other_update_resp.text

        overview_resp = client.get(
            f"{API_PREFIX}/automation/dashboard/overview",
            params={"days": 7, "job_name": matched_execution["job_name"]},
            headers=auth_headers,
        )
        assert overview_resp.status_code == 200, overview_resp.text

        data = overview_resp.json()["data"]
        assert data["summary"]["success_cases_7d"] == 1
        assert data["summary"]["failure_cases_7d"] == 0
        assert data["summary"]["total_executions_count"] == 1
        assert all(
            item["job_name"] == matched_execution["job_name"]
            for item in data["running_executions"]
        )

    def test_running_execution_cases_endpoint_returns_current_case_snapshots(
        self,
        client: TestClient,
        service_headers: dict[str, str],
        auth_headers: dict[str, str],
    ):
        execution = _create_execution(client, service_headers, pre_cases_count=2)

        running_item = _create_item(
            client, service_headers, execution["id"], "test_case_running"
        )
        failed_item = _create_item(
            client, service_headers, execution["id"], "test_case_failed"
        )
        fail_resp = client.patch(
            f"{API_PREFIX}/automation/executions/{execution['id']}/items/{failed_item['id']}",
            json={
                "status": "failure",
                "end_at": _now_iso(),
                "duration": 2.1,
                "error_message": "assert x == y",
            },
            headers=service_headers,
        )
        assert fail_resp.status_code == 200, fail_resp.text

        cases_resp = client.get(
            f"{API_PREFIX}/automation/dashboard/running-executions/{execution['id']}/cases",
            headers=auth_headers,
        )
        assert cases_resp.status_code == 200, cases_resp.text

        items = cases_resp.json()["data"]["items"]
        assert len(items) == 2

        items_by_id = {item["item_id"]: item for item in items}
        assert items_by_id[running_item["id"]]["status"] == "running"
        assert items_by_id[failed_item["id"]]["status"] == "failure"
        assert items_by_id[failed_item["id"]]["error_message"] == "assert x == y"
