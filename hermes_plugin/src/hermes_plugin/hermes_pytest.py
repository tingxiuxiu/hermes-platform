from __future__ import annotations

import os

import pytest

from hermes_plugin._hermes_listener import HermesListener


def pytest_addoption(parser: pytest.Parser) -> None:
    group = parser.getgroup(
        "hermes-plugin", "Hermes pytest plugin for live execution reporting"
    )
    group.addoption(
        "--enable-hermes-plugin",
        action="store_true",
        default=False,
        help="Enable Hermes event publishing",
    )
    group.addoption(
        "--tap-url",
        action="store",
        default=os.getenv("HERMES_URL") or os.getenv("TAP_URL") or "http://127.0.0.1:8080",
        help="Hermes API origin, e.g. http://127.0.0.1:8080 (paths use /tap/api/v1)",
    )
    group.addoption(
        "--tap-token",
        default=os.getenv("TAP_TOKEN"),
        help="Automation service token, sent as X-Service-Token (env: TAP_TOKEN)",
    )
    group.addoption(
        "--build-uid",
        default=os.getenv("BUILD_UID"),
        help="Unique build UID for this pytest run (one TestExecution)",
    )
    group.addoption(
        "--tap-labels",
        default=None,
        help="Comma-separated labels for the build",
    )
    group.addoption(
        "--jenkins-job-url",
        default="http://localhost:8080/job/default",
        help="Jenkins job URL",
    )
    group.addoption(
        "--jenkins-job-name",
        default="Default Job",
        help="Jenkins job name",
    )
    group.addoption(
        "--tap-project-name",
        default="Default Project",
        help="Project name",
    )
    group.addoption(
        "--tap-software-name",
        default="Default Software",
        help="Software under test name",
    )
    group.addoption(
        "--tap-software-version",
        default="0.0.0",
        help="Software under test version",
    )
    group.addoption(
        "--hermes-max-workers",
        type=int,
        default=4,
        help="Thread pool size for publishing events",
    )
    group.addoption(
        "--hermes-queue-maxsize",
        type=int,
        default=2000,
        help="Max in-flight events before non-step events are dropped",
    )
    group.addoption(
        "--hermes-timeout",
        type=float,
        default=10,
        help="HTTP timeout in seconds for each Hermes API call",
    )
    group.addoption(
        "--hermes-heartbeat-interval",
        type=float,
        default=12.0,
        help="Seconds between execution heartbeats (default 12)",
    )


def pytest_configure(config: pytest.Config) -> None:
    if not config.getoption("enable_hermes_plugin"):
        return
    if getattr(config, "workerinput", None) is not None:
        raise pytest.UsageError(
            "Hermes plugin v1 does not support pytest-xdist. Remove -n / xdist workers."
        )

    tap_url = config.getoption("tap_url")
    tap_token = config.getoption("tap_token")
    build_uid = config.getoption("build_uid")
    missing = [
        flag
        for flag, value in (
            ("--tap-url", tap_url),
            ("--tap-token", tap_token),
            ("--build-uid", build_uid),
        )
        if not value
    ]
    if missing:
        raise pytest.UsageError(
            f"Hermes pytest plugin requires {', '.join(missing)} "
            "(or HERMES_URL / TAP_TOKEN / BUILD_UID)"
        )

    config.pluginmanager.register(HermesListener(config), "hermes-listener")
