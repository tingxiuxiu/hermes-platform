from __future__ import annotations

import uuid
from types import SimpleNamespace

import httpx
import pytest

from hermes_plugin import hermes_pytest


@pytest.fixture
def recorded_http(monkeypatch):
    calls: list[dict] = []

    def fake_request(self, method, url, **kwargs):
        req = httpx.Request(method, str(url))
        calls.append({"method": method, "url": str(url), "json": kwargs.get("json")})
        return httpx.Response(
            200,
            request=req,
            json={"code": 200, "message": "ok", "data": {}, "success": True},
        )

    monkeypatch.setattr(httpx.Client, "request", fake_request)
    return calls


def _hermes_args(build_uid: str, heartbeat: float = 12.0) -> list[str]:
    return [
        "--enable-hermes-plugin",
        "--tap-url=http://hermes.test",
        "--tap-token=test-token",
        f"--build-uid={build_uid}",
        f"--hermes-heartbeat-interval={heartbeat}",
        "-p",
        "no:cacheprovider",
    ]


def test_xdist_worker_configure_raises():
    config = SimpleNamespace(
        workerinput={"workerid": "gw0"},
        getoption=lambda name, default=None: True if name == "enable_hermes_plugin" else "x",
    )
    with pytest.raises(pytest.UsageError, match="pytest-xdist"):
        hermes_pytest.pytest_configure(config)


def test_five_level_steps_emit_running_then_passed(pytester, recorded_http):
    pytester.makepyfile(
        """
        import allure

        def test_nested():
            with allure.step("L1"):
                with allure.step("L2"):
                    with allure.step("L3"):
                        with allure.step("L4"):
                            with allure.step("L5"):
                                assert True
        """
    )
    build_uid = str(uuid.uuid4())
    result = pytester.runpytest_inprocess(*_hermes_args(build_uid))
    result.assert_outcomes(passed=1)

    step_calls = [c for c in recorded_http if c["url"].endswith("/steps")]
    paths = [(c["json"]["step_path"], c["json"]["status"]) for c in step_calls]
    expected_paths = ["0", "0.0", "0.0.0", "0.0.0.0", "0.0.0.0.0"]
    for path in expected_paths:
        assert (path, "running") in paths
        assert (path, "passed") in paths
    assert all(len(c["json"]["step_path"].split(".")) <= 5 for c in step_calls)
    assert any(c["json"].get("start_time", "").endswith("Z") for c in step_calls)


def test_sixth_level_fails_without_posting_sixth(pytester, recorded_http):
    pytester.makepyfile(
        """
        import allure

        def test_too_deep():
            with allure.step("L1"):
                with allure.step("L2"):
                    with allure.step("L3"):
                        with allure.step("L4"):
                            with allure.step("L5"):
                                with allure.step("L6"):
                                    assert True
        """
    )
    build_uid = str(uuid.uuid4())
    result = pytester.runpytest_inprocess(*_hermes_args(build_uid))
    result.assert_outcomes(failed=1)
    assert "nested more than 5 levels" in str(result.outlines)

    step_calls = [c for c in recorded_http if c["url"].endswith("/steps")]
    sixth = [c for c in step_calls if len(c["json"]["step_path"].split(".")) >= 6]
    assert sixth == []
    assert any(c["json"]["step_path"] == "0.0.0.0.0" for c in step_calls)


def test_session_creates_execution_and_heartbeats(pytester, recorded_http):
    pytester.makepyfile(
        """
        def test_ok():
            assert True
        """
    )
    build_uid = str(uuid.uuid4())
    result = pytester.runpytest_inprocess(*_hermes_args(build_uid, heartbeat=0.05))
    result.assert_outcomes(passed=1)

    methods_urls = [(c["method"], c["url"]) for c in recorded_http]
    assert any(
        m == "POST" and u.endswith("/automation/executions") for m, u in methods_urls
    ), methods_urls
    create = next(c for c in recorded_http if c["url"].endswith("/automation/executions"))
    assert create["json"]["build_uid"] == build_uid
    assert str(create["json"]["start_time"]).endswith("Z")
    assert any("/heartbeat" in u for _, u in methods_urls)
    assert any(
        m == "PATCH" and u.endswith(f"/automation/executions/{build_uid}")
        for m, u in methods_urls
    )
