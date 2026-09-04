package worker_test

import (
	"context"
	"errors"
	"log/slog"
	"testing"

	"github.com/hibiken/asynq"

	"github.com/hermes-platform/go-service/internal/adapter/worker"
	"github.com/hermes-platform/go-service/internal/platform/queue"
)

// fakeProjector 记录调用并模拟结果。
type fakeProjector struct {
	executionCalls []int64
	allCalls       []int
	execErr        error
	allErr         error
}

func (f *fakeProjector) ProjectExecution(_ context.Context, id int64) error {
	f.executionCalls = append(f.executionCalls, id)
	return f.execErr
}

func (f *fakeProjector) RefreshAll(_ context.Context, days int) error {
	f.allCalls = append(f.allCalls, days)
	return f.allErr
}

func newHandler(t *testing.T, fp *fakeProjector) *worker.DashboardHandler {
	t.Helper()
	return worker.NewDashboardHandler(fp, slog.Default())
}

func TestHandleRefreshExecution(t *testing.T) {
	fp := &fakeProjector{}
	handler := newHandler(t, fp)

	task, _ := queue.NewDashboardRefreshExecutionTask(99)
	if err := handler.Handle(context.Background(), task); err != nil {
		t.Fatalf("handle execution refresh: %v", err)
	}

	if len(fp.executionCalls) != 1 || fp.executionCalls[0] != 99 {
		t.Errorf("execution calls = %v, want [99]", fp.executionCalls)
	}
	if len(fp.allCalls) != 0 {
		t.Errorf("must not call refresh-all, got %v", fp.allCalls)
	}
}

func TestHandleRefreshAllUsesDefaultDays(t *testing.T) {
	fp := &fakeProjector{}
	handler := newHandler(t, fp)

	task, _ := queue.NewDashboardRefreshAllTask(0) // 0 → 默认 7
	if err := handler.Handle(context.Background(), task); err != nil {
		t.Fatalf("handle all refresh: %v", err)
	}

	if len(fp.allCalls) != 1 || fp.allCalls[0] != 7 {
		t.Errorf("all calls = %v, want [7]", fp.allCalls)
	}
}

func TestHandleRefreshAllUsesGivenDays(t *testing.T) {
	fp := &fakeProjector{}
	handler := newHandler(t, fp)

	task, _ := queue.NewDashboardRefreshAllTask(30)
	if err := handler.Handle(context.Background(), task); err != nil {
		t.Fatalf("handle all refresh: %v", err)
	}

	if len(fp.allCalls) != 1 || fp.allCalls[0] != 30 {
		t.Errorf("all calls = %v, want [30]", fp.allCalls)
	}
}

func TestHandlerPropagatesErrorForRetry(t *testing.T) {
	fp := &fakeProjector{execErr: errors.New("projection failed")}
	handler := newHandler(t, fp)

	task, _ := queue.NewDashboardRefreshExecutionTask(5)
	err := handler.Handle(context.Background(), task)
	if err == nil {
		t.Fatal("projection error must propagate to trigger asynq retry")
	}
}

func TestHandlerRejectsBadExecutionPayload(t *testing.T) {
	fp := &fakeProjector{}
	handler := newHandler(t, fp)

	// 非法 payload（execution_id=0）
	task := asynq.NewTask(queue.TypeDashboardRefreshExecution, []byte(`{"execution_id":0}`))
	if err := handler.Handle(context.Background(), task); err != nil {
		t.Fatalf("bad payload must be treated as non-retryable (return nil), got %v", err)
	}
	if len(fp.executionCalls) != 0 {
		t.Error("must not call projector for bad payload")
	}
}
