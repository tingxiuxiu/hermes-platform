"""异步上报：execution/item 可丢；step enter/exit 不可静默丢弃。"""

from __future__ import annotations

import logging
import queue
import threading
import time
from concurrent.futures import Future, ThreadPoolExecutor
from typing import Any, Callable, TypeVar

import httpx

logger = logging.getLogger("hermes_plugin")

T = TypeVar("T")

API_PREFIX = "/tap/api/v1"


def normalize_base_url(url: str) -> str:
    u = url.rstrip("/")
    for suffix in ("/tap/api/v1", "/hermes-platform/api/v1"):
        if u.endswith(suffix):
            trimmed = u[: -len(suffix)]
            return trimmed or u
    return u


class HermesEventPublisher:
    def __init__(
        self,
        base_url: str | None,
        service_token: str | None,
        *,
        max_workers: int = 4,
        queue_maxsize: int = 2000,
        timeout: float = 10.0,
        max_retries: int = 2,
        retry_backoff: float = 0.5,
    ) -> None:
        origin = normalize_base_url(base_url) if base_url else None
        self.enabled = bool(origin and service_token)
        if not self.enabled:
            logger.warning(
                "hermes_plugin: --tap-url/--tap-token not fully configured; events will not be sent"
            )
        self._client = (
            httpx.Client(
                base_url=origin,
                headers={"X-Service-Token": service_token},
                timeout=timeout,
            )
            if self.enabled
            else None
        )
        self._max_retries = max_retries
        self._retry_backoff = retry_backoff

        self._executor = ThreadPoolExecutor(
            max_workers=max_workers, thread_name_prefix="hermes-plugin"
        )
        self._queue: queue.Queue = queue.Queue(maxsize=queue_maxsize)
        self._pending_lock = threading.Lock()
        self._pending: set[Future] = set()
        self._closed = False
        self._step_tails: dict[str, Future] = {}

    def submit(
        self, fn: Callable[..., T], *args: Any, droppable: bool = True, **kwargs: Any
    ) -> "Future[T | None]":
        if self._closed:
            future: Future = Future()
            future.set_result(None)
            return future

        tracked = True
        try:
            if droppable:
                self._queue.put_nowait(None)
            else:
                self._queue.put(None, timeout=2.0)
        except queue.Full:
            if not droppable:
                logger.error(
                    "hermes_plugin: step queue full; sending synchronously so live UI does not go blank"
                )
                try:
                    result = fn(*args, **kwargs)
                except Exception:
                    logger.warning(
                        "hermes_plugin: synchronous step send %r raised",
                        getattr(fn, "__name__", fn),
                        exc_info=True,
                    )
                    result = None
                done: Future = Future()
                done.set_result(result)
                return done
            tracked = False
            logger.warning(
                "hermes_plugin: event queue is full; dropping a non-step event"
            )

        def _run() -> T | None:
            try:
                return fn(*args, **kwargs)
            except Exception:
                logger.warning(
                    "hermes_plugin: background task %r raised",
                    getattr(fn, "__name__", fn),
                    exc_info=True,
                )
                return None
            finally:
                if tracked:
                    try:
                        self._queue.get_nowait()
                    except queue.Empty:
                        pass

        future = self._executor.submit(_run)
        with self._pending_lock:
            self._pending.add(future)
        future.add_done_callback(self._on_done)
        return future

    def _on_done(self, future: Future) -> None:
        with self._pending_lock:
            self._pending.discard(future)

    def _request(self, method: str, path: str, json_body: dict | None = None) -> Any:
        if not self.enabled or self._client is None:
            return None

        last_exc: Exception | None = None
        for attempt in range(self._max_retries + 1):
            try:
                response = self._client.request(method, path, json=json_body)
                if response.status_code >= 400:
                    body = response.text[:300] if response.content else ""
                    logger.warning(
                        "hermes_plugin: %s %s -> HTTP %s: %s",
                        method,
                        path,
                        response.status_code,
                        body,
                    )
                    return None
                if not response.content:
                    return None
                try:
                    return response.json().get("data")
                except Exception:
                    return None
            except httpx.HTTPStatusError as exc:
                body = exc.response.text[:300] if exc.response is not None else ""
                logger.warning(
                    "hermes_plugin: %s %s -> HTTP %s: %s",
                    method,
                    path,
                    exc.response.status_code,
                    body,
                )
                return None
            except httpx.HTTPError as exc:
                last_exc = exc
                if attempt < self._max_retries:
                    time.sleep(self._retry_backoff * (2**attempt))
                    continue
        logger.warning(
            "hermes_plugin: %s %s failed after %d retries: %s",
            method,
            path,
            self._max_retries,
            last_exc,
        )
        return None

    def create_execution(self, payload: dict) -> dict | None:
        return self._request("POST", f"{API_PREFIX}/automation/executions", payload)

    def update_execution(self, build_uid: str, payload: dict) -> dict | None:
        return self._request(
            "PATCH", f"{API_PREFIX}/automation/executions/{build_uid}", payload
        )

    def heartbeat(self, build_uid: str) -> dict | None:
        return self._request(
            "POST", f"{API_PREFIX}/automation/executions/{build_uid}/heartbeat", {}
        )

    def create_case(self, build_uid: str, payload: dict) -> dict | None:
        return self._request(
            "POST", f"{API_PREFIX}/automation/executions/{build_uid}/items", payload
        )

    def update_case(self, build_uid: str, case_uid: str, payload: dict) -> dict | None:
        return self._request(
            "PATCH",
            f"{API_PREFIX}/automation/executions/{build_uid}/items/{case_uid}",
            payload,
        )

    def upsert_step(self, case_uid: str, payload: dict, *, after: Future | None = None) -> Future:
        prev = self._step_tails.get(case_uid)

        def _run() -> Any:
            if after is not None:
                try:
                    after.result()
                except Exception:
                    logger.warning("hermes_plugin: waiting for case create failed", exc_info=True)
            if prev is not None:
                try:
                    prev.result()
                except Exception:
                    pass
            return self._request(
                "POST", f"{API_PREFIX}/automation/items/{case_uid}/steps", payload
            )

        fut = self.submit(_run, droppable=False)
        self._step_tails[case_uid] = fut
        return fut

    def close(self, timeout: float = 15.0) -> None:
        if self._closed:
            return
        self._closed = True

        deadline = time.monotonic() + timeout
        while True:
            with self._pending_lock:
                pending = list(self._pending)
            if not pending:
                break
            if time.monotonic() >= deadline:
                logger.warning(
                    "hermes_plugin: timed out waiting for %d pending event(s) to flush before exit",
                    len(pending),
                )
                break
            for future in pending:
                try:
                    future.result(timeout=max(0.0, deadline - time.monotonic()))
                except Exception:
                    pass

        self._executor.shutdown(wait=False)
        if self._client is not None:
            self._client.close()
