// Package clock 抽象时间源，让 domain / application 层不直接依赖 time.Now，
// 从而在单元测试中注入固定时钟（ADR-0004 关键纪律第 4 条）。
package clock

import (
	"context"
	"sync"
	"time"
)

// Clock 是时间源端口。
type Clock interface {
	Now() time.Time
}

// System 使用真实系统时间。
type System struct{}

func (System) Now() time.Time { return time.Now().UTC() }

// NewSystem 返回真实时钟，时间统一转为 UTC，
// 与数据库 timestamptz 列的存储语义保持一致。
func NewSystem() System { return System{} }

// Fixed 是可控时钟，供测试使用。
// 零值不可用，必须通过 NewFixed 构造。
type Fixed struct {
	mu    sync.Mutex
	now   time.Time
	delta time.Duration
}

// NewFixed 以 t 为起点创建固定时钟。t 会被规范化为 UTC。
func NewFixed(t time.Time) *Fixed {
	return &Fixed{now: t.UTC()}
}

// Now 返回当前时刻后自动前进一个步长（默认 0，即完全静止）。
func (f *Fixed) Now() time.Time {
	f.mu.Lock()
	defer f.mu.Unlock()
	cur := f.now
	f.now = f.now.Add(f.delta)
	return cur
}

// Advance 手动推进时钟，返回推进后的时刻。
func (f *Fixed) Advance(d time.Duration) time.Time {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.now = f.now.Add(d)
	return f.now
}

// SetStep 设置每次 Now() 调用自动前进的步长，用于制造严格递增的时间序列。
func (f *Fixed) SetStep(d time.Duration) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.delta = d
}

// Set 直接把时钟设置到 t（规范化为 UTC）。
func (f *Fixed) Set(t time.Time) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.now = t.UTC()
}

type ctxKey struct{}

// WithClock 把时钟注入 context。
func WithClock(ctx context.Context, c Clock) context.Context {
	return context.WithValue(ctx, ctxKey{}, c)
}

// FromContext 取出 context 中的时钟，缺省返回真实时钟。
// 缺省兜底保证未显式注入的代码路径依然可用。
func FromContext(ctx context.Context) Clock {
	if c, ok := ctx.Value(ctxKey{}).(Clock); ok && c != nil {
		return c
	}
	return System{}
}
