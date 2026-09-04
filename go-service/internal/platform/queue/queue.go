// Package queue 封装 asynq（任务队列）。
//
// 职责：
//   - Client：投递任务
//   - Server：消费任务（worker）
//   - Scheduler：定时投递（替代 celery-beat，单实例约束见 docs/01-architecture.md §4）
//
// 队列分级：default（普通）/ dashboard（读模型刷新，可独立调并发）。
// 本包只做 asynq 的薄封装，不包含任何业务逻辑。
package queue

import (
	"time"

	"github.com/hibiken/asynq"
)

// Queue 常量：队列名。
const (
	// QueueDefault 是普通任务队列。
	QueueDefault = "default"
	// QueueDashboard 是 dashboard 快照刷新队列。
	QueueDashboard = "dashboard"
)

// RetryConfig 是任务重试与保留策略。
type RetryConfig struct {
	// MaxRetry 最大重试次数。
	MaxRetry int
	// Timeout 单次执行超时。
	Timeout time.Duration
	// Retention 任务完成后保留时长（用于 asynq 的完成队列）。
	Retention time.Duration
}

// DefaultRetryConfig 默认策略：重试 3 次、10s 超时、1h 保留。
var DefaultRetryConfig = RetryConfig{
	MaxRetry:  3,
	Timeout:   10 * time.Second,
	Retention: time.Hour,
}

// Client 是任务投递客户端。
type Client struct {
	client *asynq.Client
}

// NewClient 构造投递客户端。
// 注意：返回的 Client 使用完后必须调用 Close 释放资源。
func NewClient(redisAddr string, redisDB int) *Client {
	client := asynq.NewClient(asynq.RedisClientOpt{
		Addr: redisAddr,
		DB:   redisDB,
	})
	return &Client{client: client}
}

// NewClientFromRedis 用已有的 Redis 配置构造客户端。
func NewClientFromRedis(addr string, db int) *Client {
	return NewClient(addr, db)
}

// Enqueue 投递一个任务到指定队列。
// unique 为 true 时启用 asynq 的 Unique 去重（在任务层通过 task.Option 控制）。
func (c *Client) Enqueue(task *asynq.Task, opts ...asynq.Option) (*asynq.TaskInfo, error) {
	info, err := c.client.Enqueue(task, opts...)
	if err != nil {
		return nil, err
	}
	return info, nil
}

// Close 释放客户端资源。
func (c *Client) Close() error {
	return c.client.Close()
}

// Server 是任务消费端（worker）。
type Server struct {
	server *asynq.Server
}

// NewServer 构造消费端。
// queues 是队列→并发数映射，例如 {QueueDefault: 10, QueueDashboard: 5}。
func NewServer(redisAddr string, redisDB int, queues map[string]int) *Server {
	server := asynq.NewServer(
		asynq.RedisClientOpt{Addr: redisAddr, DB: redisDB},
		asynq.Config{
			Concurrency: sumQueues(queues),
			Queues:      queues,
			// 优雅停机：等待在途任务完成（最长 30s）
			ShutdownTimeout: 30 * time.Second,
		},
	)
	return &Server{server: server}
}

// Start 开始消费。阻塞直到 ctx 取消或 Run 出错。
func (s *Server) Start(mux *asynq.ServeMux) error {
	return s.server.Run(mux)
}

// Scheduler 是定时投递端（替代 celery-beat，**单实例**）。
type Scheduler struct {
	scheduler *asynq.Scheduler
}

// NewScheduler 构造调度器。
func NewScheduler(redisAddr string, redisDB int) *Scheduler {
	s := asynq.NewScheduler(
		asynq.RedisClientOpt{Addr: redisAddr, DB: redisDB},
		&asynq.SchedulerOpts{},
	)
	return &Scheduler{scheduler: s}
}

// Register 注册一个 cron 表达式 + 任务。
// cron 表达式用标准 6 段（秒 分 时 日 月 周），与 celery-beat 的定时语义对应。
func (s *Scheduler) Register(cronSpec string, task *asynq.Task, opts ...asynq.Option) (string, error) {
	return s.scheduler.Register(cronSpec, task, opts...)
}

// Start 启动调度器（阻塞）。返回的 channel 用于等待信号。
func (s *Scheduler) Start() error {
	return s.scheduler.Run()
}

// ServeMux 是任务分发器。
func ServeMux() *asynq.ServeMux {
	return asynq.NewServeMux()
}

// WithMaxRetry 返回设置最大重试次数的选项。
func WithMaxRetry(n int) asynq.Option {
	return asynq.MaxRetry(n)
}

// WithQueue 返回指定队列的选项。
func WithQueue(q string) asynq.Option {
	return asynq.Queue(q)
}

// WithTimeout 返回设置超时的选项。
func WithTimeout(d time.Duration) asynq.Option {
	return asynq.Timeout(d)
}

// WithRetention 返回设置保留时长的选项。
func WithRetention(d time.Duration) asynq.Option {
	return asynq.Retention(d)
}

// WithUnique 返回启用唯一性约束（指定时长内同类型+同 payload 的任务只入队一个）的选项。
func WithUnique(within time.Duration) asynq.Option {
	return asynq.Unique(within)
}

func sumQueues(m map[string]int) int {
	total := 0
	for _, v := range m {
		total += v
	}
	return total
}
