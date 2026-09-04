"""HermesListener：把 pytest + Allure 步骤事件上报到 Go `/tap/api/v1`。"""

from __future__ import annotations

import logging
import threading
import time
from concurrent.futures import Future

import pytest
from allure_commons import hookimpl, plugin_manager

from hermes_plugin._event_publisher import HermesEventPublisher
from hermes_plugin.context import CaseRecord, StepFrame, _current_case
from hermes_plugin.models import (
    CaseStatusEnum,
    CreateCase,
    CreateExecution,
    ExecutionStatusEnum,
    StepStatusEnum,
    UpdateCase,
    UpdateExecution,
    UpsertStep,
    utc_now,
)

logger = logging.getLogger("hermes_plugin")

MAX_STEP_DEPTH = 5

_DEPTH_FAIL = (
    "Hermes forbids Allure steps nested more than 5 levels "
    "(path would be {path}). Flatten the test or split cases."
)


class _Heartbeat:
    def __init__(
        self,
        publisher: HermesEventPublisher,
        build_uid: str,
        interval: float,
        ready: Future | None,
    ) -> None:
        self._publisher = publisher
        self._build_uid = build_uid
        self._interval = max(interval, 0.01)
        self._ready = ready
        self._stop = threading.Event()
        self._thread = threading.Thread(target=self._run, name="hermes-heartbeat", daemon=True)

    def start(self) -> None:
        self._thread.start()

    def stop(self) -> None:
        self._stop.set()
        self._thread.join(timeout=2.0)

    def _run(self) -> None:
        try:
            if self._ready is not None:
                try:
                    self._ready.result(timeout=30)
                except Exception:
                    logger.warning(
                        "hermes_plugin: execution create did not finish before heartbeat",
                        exc_info=True,
                    )
            self._publisher.heartbeat(self._build_uid)
            while not self._stop.wait(self._interval):
                self._publisher.heartbeat(self._build_uid)
        except Exception:
            logger.warning("hermes_plugin: heartbeat thread stopped", exc_info=True)


class HermesAllureHooks:
    def __init__(self, listener: "HermesListener") -> None:
        self._listener = listener

    @hookimpl
    def start_step(self, uuid, title, params):
        self._listener.on_start_step(title)

    @hookimpl
    def stop_step(self, uuid, exc_type, exc_val, exc_tb, title=None):
        self._listener.on_stop_step(exc_type)


class HermesListener:
    def __init__(self, config: pytest.Config) -> None:
        self._build_uid: str = config.getoption("build_uid")
        self._job_name: str = config.getoption("jenkins_job_name")
        self._job_url: str = config.getoption("jenkins_job_url")
        self._project_name: str = config.getoption("tap_project_name")
        self._software_name: str = config.getoption("tap_software_name")
        self._software_version: str = config.getoption("tap_software_version")
        labels = config.getoption("tap_labels")
        self._labels: list[str] = [item for item in labels.split(",") if item] if labels else []

        self._timeout: float = config.getoption("hermes_timeout")
        self._max_retries = 2
        self._shutdown_timeout = max(self._timeout * 3, 15.0)
        self._heartbeat_interval: float = config.getoption("hermes_heartbeat_interval")

        self.publisher = HermesEventPublisher(
            base_url=config.getoption("tap_url"),
            service_token=config.getoption("tap_token"),
            max_workers=config.getoption("hermes_max_workers"),
            queue_maxsize=config.getoption("hermes_queue_maxsize"),
            timeout=self._timeout,
            max_retries=self._max_retries,
        )

        self._execution_future: Future | None = None
        self._session_perf_start: float | None = None
        self._heartbeat: _Heartbeat | None = None
        self._allure = HermesAllureHooks(self)
        plugin_manager.register(self._allure)

    def _safe_result(self, future: Future | None, timeout: float | None = None):
        if future is None:
            return None
        try:
            return future.result(timeout=timeout)
        except Exception:
            return None

    def _wait_timeout(self) -> float:
        return self._timeout * (self._max_retries + 2)

    def pytest_collection_finish(self, session: pytest.Session) -> None:
        self._session_perf_start = time.perf_counter()
        payload = CreateExecution(
            build_uid=str(self._build_uid),
            job_name=self._job_name,
            job_url=self._job_url,
            project_name=self._project_name,
            software_name=self._software_name,
            software_version=self._software_version,
            labels=self._labels,
            start_time=utc_now(),
            planned_cases_count=len(session.items),
        )
        self._execution_future = self.publisher.submit(
            self.publisher.create_execution, payload.to_dict()
        )
        self._heartbeat = _Heartbeat(
            self.publisher, self._build_uid, self._heartbeat_interval, self._execution_future
        )
        self._heartbeat.start()

    def pytest_sessionfinish(self, session: pytest.Session, exitstatus: int) -> None:
        duration = (
            time.perf_counter() - self._session_perf_start
            if self._session_perf_start
            else 0
        )
        payload = UpdateExecution(
            status=(
                ExecutionStatusEnum.COMPLETED
                if exitstatus == 0
                else ExecutionStatusEnum.FAILED
            ),
            end_time=utc_now(),
            duration=duration,
        )
        self.publisher.submit(
            self.publisher.update_execution, self._build_uid, payload.to_dict()
        )
        if self._heartbeat is not None:
            self._heartbeat.stop()
        self.publisher.close(timeout=self._shutdown_timeout)

    def pytest_unconfigure(self, config: pytest.Config) -> None:
        if self._heartbeat is not None:
            self._heartbeat.stop()
        try:
            plugin_manager.unregister(plugin=self._allure)
        except Exception:
            pass
        self.publisher.close(timeout=self._shutdown_timeout)

    @pytest.hookimpl(wrapper=True)
    def pytest_runtest_protocol(self, item: pytest.Item, nextitem):
        self._safe_result(self._execution_future, timeout=self._wait_timeout())
        case = CaseRecord(
            build_uid=self._build_uid,
            case_key=item.nodeid,
            case_name=item.name,
        )
        token = _current_case.set(case)
        create_payload = CreateCase(
            build_uid=case.build_uid,
            case_key=case.case_key,
            case_name=case.case_name,
            case_uid=case.case_uid,
            labels=case.case_labels,
            start_time=case.start_time,
        )
        case.create_future = self.publisher.submit(
            self.publisher.create_case, case.build_uid, create_payload.to_dict()
        )
        self._safe_result(case.create_future, timeout=self._wait_timeout())
        started = time.perf_counter()
        try:
            yield
        finally:
            payload = UpdateCase(
                status=case.status if case.status != CaseStatusEnum.RUNNING else CaseStatusEnum.PASSED,
                end_time=utc_now(),
                duration=max(time.perf_counter() - started, 0.0),
                error_message=case.error_message,
                error_traceback=case.error_traceback,
            )
            create_fut = case.create_future

            def _finish_case():
                if create_fut is not None:
                    try:
                        create_fut.result()
                    except Exception:
                        pass
                return self.publisher.update_case(
                    self._build_uid, case.case_uid, payload.to_dict()
                )

            self.publisher.submit(_finish_case)
            _current_case.reset(token)

    @pytest.hookimpl(hookwrapper=True)
    def pytest_runtest_makereport(self, item, call):
        outcome = yield
        report = outcome.get_result()
        case = _current_case.get()
        if not case:
            return

        if report.failed:
            if report.when == "call":
                case.status = CaseStatusEnum.FAILED
            else:
                case.status = CaseStatusEnum.BROKEN
            if report.longreprtext:
                case.error_traceback = report.longreprtext
                case.error_message = report.longreprtext.splitlines()[0][:500]
        elif report.skipped and case.status == CaseStatusEnum.RUNNING:
            case.status = CaseStatusEnum.SKIPPED
        elif report.passed and report.when == "call" and case.status == CaseStatusEnum.RUNNING:
            case.status = CaseStatusEnum.PASSED

    def on_start_step(self, title: str) -> None:
        case = _current_case.get()
        if case is None:
            return
        if len(case.step_stack) >= MAX_STEP_DEPTH:
            parent = case.step_stack[-1]
            would_be = f"{parent.path}.{parent.child_count}"
            pytest.fail(_DEPTH_FAIL.format(path=would_be))

        if not case.step_stack:
            index = case.root_count
            path = str(index)
            case.root_count += 1
        else:
            parent = case.step_stack[-1]
            index = parent.child_count
            path = f"{parent.path}.{index}"
            parent.child_count += 1

        start = utc_now()
        case.step_stack.append(StepFrame(path=path, name=title, start_time=start))
        self.publisher.upsert_step(
            case.case_uid,
            UpsertStep(
                step_path=path,
                step_name=title,
                status=StepStatusEnum.RUNNING,
                start_time=start,
            ).to_dict(),
            after=case.create_future,
        )

    def on_stop_step(self, exc_type) -> None:
        case = _current_case.get()
        if case is None or not case.step_stack:
            return
        frame = case.step_stack.pop()
        end = utc_now()
        status = StepStatusEnum.FAILED if exc_type is not None else StepStatusEnum.PASSED
        self.publisher.upsert_step(
            case.case_uid,
            UpsertStep(
                step_path=frame.path,
                step_name=frame.name,
                status=status,
                start_time=frame.start_time,
                end_time=end,
            ).to_dict(),
            after=case.create_future,
        )
