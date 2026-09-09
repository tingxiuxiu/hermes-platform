# hermes_plugin

pytest 插件：把一次 pytest 运行上报到 Hermes Go API（`/tap/api/v1`），供 Web 会话页实时展示当前 case / step。

Allure 是步骤真相源。必须同时启用 Allure 文件日志与 Hermes 上报。

## 依赖

- `allure-pytest`
- `pytest` / `httpx` / `pydantic`

## 运行

```bash
pytest --enable-hermes-plugin \
  --alluredir allure-results \
  --tap-url http://127.0.0.1:8080 \
  --tap-token <X-Service-Token> \
  --build-uid <uuid> \
  tests/

pytest --enable-hermes-plugin --tap-token local-dev-automation-service-token-3f6a1b9c --build-uid cfd58d9f-5251-4c7e-87c8-57318d3e4897 tests\test_mock_business.py
```

`--tap-url` 填 API origin（不要带 `/tap/api/v1`），默认 `http://127.0.0.1:8080`。也可用环境变量：

| 参数 | 环境变量 |
| --- | --- |
| `--tap-url` | `HERMES_URL` 或 `TAP_URL` |
| `--tap-token` | `TAP_TOKEN` |
| `--build-uid` | `BUILD_UID` |

`--alluredir` 由 Allure 插件提供，不是 Hermes 参数；没有它拿不到 Allure 报告。

v1 **不支持 pytest-xdist**。检测到 worker 会 `UsageError`：`Hermes plugin v1 does not support pytest-xdist. Remove -n / xdist workers.`

步骤最多 5 层。第 6 层会在测试进程里 `pytest.fail`（文案含 `nested more than 5 levels`），不会上报 Hermes。

## 测试里怎么写步骤

```python
import allure
from hermes_plugin import step, attachment, AttachmentTypeEnum  # step is allure.step

def test_login():
    with allure.step("打开页面"):
        with step("输入账号"):
            attachment("hint", "demo", AttachmentTypeEnum.LOG)
            assert True
```

附件只写入 Allure，不上报 Hermes。

## 插件会上报什么

路径均相对 `/tap/api/v1`。

| 时机 | 请求 |
| --- | --- |
| collection 结束 | `POST /automation/executions` |
| 每个 attempt 开始/结束 | `POST/PATCH /automation/executions/{build_uid}/items` |
| Allure step enter/exit | `POST /automation/items/{case_uid}/steps` |
| 每 12s（`--hermes-heartbeat-interval`） | `POST /automation/executions/{build_uid}/heartbeat` |
| session 结束 | `PATCH /automation/executions/{build_uid}` |

时间字段为 RFC3339 UTC。
