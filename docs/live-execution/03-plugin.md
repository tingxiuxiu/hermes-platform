# 03 · pytest 插件（Allure 为源）

## 1. 依赖

`pyproject.toml` 硬依赖：

- `allure-pytest`（及传递的 `allure-python-commons`）
- 现有 `pytest` / `httpx` / `pydantic`

用户必须同时启用 Allure 文件日志（出报告）与 Hermes 上报，例如：

```bash
pytest --enable-hermes-plugin \
  --alluredir allure-results \
  --tap-token ... --build-uid ... \
  tests/
```

`--alluredir` 不是 Hermes 的职责，但没有它客户拿不到 Allure 报告。插件文档必须写清两条缺一不可。

## 2. 公共 API

```python
# 推荐（客户已有 Allure 习惯）
import allure
with allure.step("登录"):
    ...

# 兼容别名（现有 demo / 测试）
from hermes_plugin import step, attachment  # step is allure.step
```

删除或掏空自研 `context.step` 树作为真相源。别名实现：`step = allure.step`。`attachment` 包装 `allure.attach`，**不** 再构造 `data:` URL 打给 Hermes。

## 3. Listener 职责

一个 `HermesListener`（pytest plugin）+ 订阅 Allure listener：

| 钩子 / Allure 事件 | Hermes 动作 |
| --- | --- |
| `pytest_configure` | 无 `--enable-hermes-plugin` 则注册空操作。检测到 xdist **worker** → **`pytest.UsageError` / `pytest.exit`**，文案：本插件 v1 不支持 pytest-xdist，请去掉 `-n`。 |
| `pytest_collection_finish` | `POST /executions`（running + planned_cases_count） |
| Allure test start ≈ `pytest_runtest_protocol` 进入 | `POST /executions/{build_uid}/items`（case running）。`case_key=nodeid`，`case_uid` 每 attempt 新 UUID |
| Allure `start_step` | 维护栈；栈深将变为 6 → **立刻失败该测试**（raise / `pytest.fail`），并带路径说明。否则 `POST .../items/{case_uid}/steps` status=`running`，`end_time` 省略 |
| Allure `stop_step` | 弹栈；`POST` 同 path status=passed/failed/... 带 end_time |
| Allure test stop / `pytest_runtest_makereport` | `PATCH` item 终态（passed/failed/skipped/broken），**不上报 attachments** |
| 后台线程 | 每 10–15s `POST .../executions/{build_uid}/heartbeat` |
| `pytest_sessionfinish` | `PATCH` execution completed/failed；`close()` flush。去掉「只写本地 hermes-report.json 且覆盖 sessionfinish」的重复钩子 |
| `pytest_unconfigure` | flush 心跳停止 |

步骤 `step_path`：插件按 **当前 case 内 Allure 步骤栈** 自己编号（与现 `context.py` 相同：兄弟用递增 index，子节点 `parent.path + "." + index`）。Allure 内部 UUID **不** 当幂等键。

## 4. 深度失败（调试期强制）

```text
Hermes forbids Allure steps nested more than 5 levels
(path would be 0.1.2.3.4.5). Flatten the test or split cases.
```

失败发生在 **测试进程**，不等后端。后端仍拒写第 6 层，防止绕过插件。

Allure 本地结果里仍可能有第 6 层（若失败前已写入）。Hermes 树与 Allure 树在违规用例上允许不一致。

## 5. 并发与顺序

- `create_execution` 的 Future 必须在 **第一条 item/step 上报前** 在工作线程 `result()`（现设计已有，须保住）。
- **step enter/exit** 对同一 `case_uid` 必须保序：同一 case 用单线程队列或 per-case lock，避免 exit 先于 enter 到达。
- 队列满：**step 上报不可 drop**。`put` 带超时；超时则 log error 且该 step 同步补发一次。宁可拖慢测试几十毫秒，不能让 live 卡片空白。
- 心跳、execution 结束允许 drop+retry。

## 6. xdist

`config.workerinput is not None` → **硬失败**，禁止静默 skip。

## 7. 时间

`start_time` / `end_time` 用 UTC ISO-8601 或 Unix 秒，与 Go `time.Time` JSON 对齐。现插件混用 `time.time()` float；Go DTO 是 `time.Time`。插件改为 **RFC3339 UTC 字符串**，与集成测试 `createExecution` 一致。

## 8. 测试接缝（插件）

| 接缝 | 测什么 |
| --- | --- |
| pytest + allure-pytest（真实钩子，httpx mock transport） | 5 层 step 上报 running→passed；第 6 层测试失败且无第 6 层请求 |
| 同上 | xdist 标记存在时 configure 失败 |
| 同上 | session 开始有 execution create；heartbeat 至少一次（可注入短间隔） |
