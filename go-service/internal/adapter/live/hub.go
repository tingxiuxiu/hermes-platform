// Package live 用 Redis Pub/Sub 做 execution 级 live fan-out（不走 asynq）。
package live

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/google/uuid"

	appautomation "github.com/hermes-platform/go-service/internal/application/automation"
	"github.com/hermes-platform/go-service/internal/domain/automation"
	appredis "github.com/hermes-platform/go-service/internal/platform/redis"
)

const channelPrefix = "hermes:live:"

// Hub 同时实现 LivePublisher 与 LiveSubscriber。
type Hub struct {
	rdb *appredis.Client
}

// NewHub 构造 Redis live hub。
func NewHub(rdb *appredis.Client) *Hub {
	return &Hub{rdb: rdb}
}

func channel(buildUID uuid.UUID) string {
	return channelPrefix + buildUID.String()
}

type stepDelta struct {
	Type     string `json:"type"`
	BuildUID string `json:"build_uid"`
	CaseUID  string `json:"case_uid"`
	StepPath string `json:"step_path"`
	StepName string `json:"step_name"`
	Status   string `json:"status"`
}

// Publish 把领域事件写成 JSON 并 PUBLISH。未知类型忽略。
func (h *Hub) Publish(ctx context.Context, event automation.Event) error {
	if h == nil || h.rdb == nil {
		return nil
	}
	switch e := event.(type) {
	case automation.StepUpserted:
		payload, err := json.Marshal(stepDelta{
			Type:     "step.upserted",
			BuildUID: e.BuildUID.String(),
			CaseUID:  e.CaseUID.String(),
			StepPath: e.StepPath,
			StepName: e.StepName,
			Status:   e.Status,
		})
		if err != nil {
			return fmt.Errorf("live: marshal step delta: %w", err)
		}
		if err := h.rdb.Publish(ctx, channel(e.BuildUID), payload).Err(); err != nil {
			return fmt.Errorf("live: publish: %w", err)
		}
	}
	return nil
}

// Subscribe 订阅一次 execution 的 live 频道。调用方取消 ctx 后通道关闭。
func (h *Hub) Subscribe(ctx context.Context, buildUID uuid.UUID) (<-chan appautomation.LiveFrame, error) {
	if h == nil || h.rdb == nil {
		return nil, fmt.Errorf("live: hub not configured")
	}
	pubsub := h.rdb.Subscribe(ctx, channel(buildUID))
	if _, err := pubsub.Receive(ctx); err != nil {
		_ = pubsub.Close()
		return nil, fmt.Errorf("live: subscribe: %w", err)
	}

	out := make(chan appautomation.LiveFrame, 16)
	go func() {
		defer close(out)
		defer pubsub.Close()
		ch := pubsub.Channel()
		for {
			select {
			case <-ctx.Done():
				return
			case msg, ok := <-ch:
				if !ok {
					return
				}
				var peek struct {
					Type string `json:"type"`
				}
				_ = json.Unmarshal([]byte(msg.Payload), &peek)
				event := peek.Type
				if event == "" {
					event = "message"
				}
				frame := appautomation.LiveFrame{Event: event, JSON: []byte(msg.Payload)}
				select {
				case out <- frame:
				case <-ctx.Done():
					return
				}
			}
		}
	}()
	return out, nil
}
