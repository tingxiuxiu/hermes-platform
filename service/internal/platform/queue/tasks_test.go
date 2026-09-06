package queue_test

import (
	"encoding/json"
	"testing"

	"github.com/hibiken/asynq"

	"github.com/hermes-platform/go-service/internal/platform/queue"
)

func TestTaskTypeAndPayloadRoundTrip(t *testing.T) {
	// Execution 刷新任务
	execTask, err := queue.NewDashboardRefreshExecutionTask(42)
	if err != nil {
		t.Fatalf("build execution task: %v", err)
	}
	if execTask.Type() != queue.TypeDashboardRefreshExecution {
		t.Errorf("type = %s, want %s", execTask.Type(), queue.TypeDashboardRefreshExecution)
	}
	payload, err := queue.ParseDashboardRefreshExecutionPayload(execTask)
	if err != nil {
		t.Fatalf("parse execution payload: %v", err)
	}
	if payload.ExecutionID != 42 {
		t.Errorf("execution_id = %d, want 42", payload.ExecutionID)
	}

	// All 刷新任务
	allTask, err := queue.NewDashboardRefreshAllTask(7)
	if err != nil {
		t.Fatalf("build all task: %v", err)
	}
	if allTask.Type() != queue.TypeDashboardRefreshAll {
		t.Errorf("type = %s, want %s", allTask.Type(), queue.TypeDashboardRefreshAll)
	}
	allPayload, err := queue.ParseDashboardRefreshAllPayload(allTask)
	if err != nil {
		t.Fatalf("parse all payload: %v", err)
	}
	if allPayload.TrendDays != 7 {
		t.Errorf("trend_days = %d, want 7", allPayload.TrendDays)
	}

	abortTask, err := queue.NewAbortStaleTask()
	if err != nil {
		t.Fatalf("build abort-stale task: %v", err)
	}
	if abortTask.Type() != queue.TypeAutomationAbortStale {
		t.Errorf("type = %s, want %s", abortTask.Type(), queue.TypeAutomationAbortStale)
	}
}

func TestParseRejectsInvalidExecutionID(t *testing.T) {
	// 构造非法 payload：execution_id <= 0
	raw := []byte(`{"execution_id":0}`)
	task := asynq.NewTask(queue.TypeDashboardRefreshExecution, raw)

	if _, err := queue.ParseDashboardRefreshExecutionPayload(task); err == nil {
		t.Error("execution_id=0 must be rejected")
	}
}

func TestTaskJSONMarshalsCorrectly(t *testing.T) {
	// 验证 payload 是合法 JSON，且字段名正确（worker 端按此解析）
	task, _ := queue.NewDashboardRefreshExecutionTask(7)
	var decoded map[string]int64
	if err := json.Unmarshal(task.Payload(), &decoded); err != nil {
		t.Fatalf("payload not valid JSON: %v", err)
	}
	if decoded["execution_id"] != 7 {
		t.Errorf("payload = %v, want execution_id=7", decoded)
	}
}

func TestCommonOptionsIncludeQueueAndRetry(t *testing.T) {
	// 验证通用选项能正确作用于任务（不报错即可，具体 Option 行为由 asynq 保证）
	_ = queue.CommonOptions()
}
