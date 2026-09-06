package dashboard

import (
	"strconv"
	"time"
)

// 投影器是纯函数：输入源数据 → 输出快照。
// 所有「源表 → 快照表」的重算逻辑都收拢在这里，便于单测与复用。

// ProgressPercent 计算进度百分比。
// 口径：已完成用例数 / 预计用例总数；预计为 0 时返回 0（避免除零）。
func ProgressPercent(completed, pre int) float64 {
	if pre <= 0 {
		return 0
	}
	return float64(completed) / float64(pre) * 100
}

// BuildExecutionSnapshot 由源数据投影出一条 execution 快照。
func BuildExecutionSnapshot(in ExecutionInput, now time.Time) ExecutionSnapshot {
	var completed, success, failure, skipped, running int
	runningNames := make([]string, 0, 5)

	for _, c := range in.Cases {
		switch c.Status {
		case "passed":
			success++
			completed++
		case "failed", "broken":
			failure++
			completed++
		case "skipped":
			skipped++
			completed++
		case "running":
			running++
			if len(runningNames) < 5 {
				runningNames = append(runningNames, c.CaseName)
			}
		}
	}

	return ExecutionSnapshot{
		ExecutionID:         in.ExecutionID,
		BuildUID:            in.BuildUID,
		JobID:               in.JobID,
		JobName:             in.JobName,
		JobURL:              in.JobURL,
		Status:              in.Status,
		StartedAt:           in.StartedAt,
		Duration:            in.Duration,
		PreCasesCount:       in.PreCasesCount,
		CompletedCasesCount: completed,
		SuccessCasesCount:   success,
		FailureCasesCount:   failure,
		SkippedCasesCount:   skipped,
		RunningCasesCount:   running,
		ProgressPercent:     ProgressPercent(completed, in.PreCasesCount),
		RunningCaseNames:    runningNames,
		UpdatedAt:           now,
	}
}

// BuildCaseSnapshots 由源数据投影出 case 快照列表。
// 只投影**最新** attempt（避免历史 attempt 污染读模型）。
//
// 缺陷 D-15：投影的 ItemID 就是 execution_items.id（case_id 列），
// 语义是「该用例当前生效的 attempt」。
func BuildCaseSnapshots(in ExecutionInput, now time.Time) []CaseSnapshot {
	// 按 (case_key) 分组，每组只取 attempt_number 最大的
	latest := make(map[string]CaseInput)
	for _, c := range in.Cases {
		prev, ok := latest[c.CaseKey]
		if !ok || c.AttemptNumber > prev.AttemptNumber {
			latest[c.CaseKey] = c
		}
	}

	out := make([]CaseSnapshot, 0, len(latest))
	for _, c := range latest {
		out = append(out, CaseSnapshot{
			ExecutionID:   in.ExecutionID,
			ItemID:        c.ItemID,
			CaseKey:       c.CaseKey,
			CaseName:      c.CaseName,
			AttemptNumber: c.AttemptNumber,
			Status:        c.Status,
			StartedAt:     c.StartedAt,
			EndedAt:       c.EndedAt,
			Duration:      c.Duration,
			ErrorMessage:  c.ErrorMessage,
			UpdatedAt:     now,
		})
	}
	return out
}

// CaseStatusCategory 把用例状态归类为 success / failure / skipped。
//
// 缺陷 D-16：passed → success；failed + broken → failure；skipped → skipped。
// running 视为未归类（返回空串）。
func CaseStatusCategory(status string) string {
	switch status {
	case "passed":
		return "success"
	case "failed", "broken":
		return "failure"
	case "skipped":
		return "skipped"
	default:
		return ""
	}
}

// SummaryInput 是全局摘要投影所需的聚合输入。
type SummaryInput struct {
	// TerminalExecutions 是近 7 天的终止态 execution（用于 execution 数与平均耗时）。
	TerminalExecutions []TerminalExecution
	// TerminalCases 是近 7 天的终止态用例（用于成功/失败/跳过计数与通过率）。
	TerminalCases []TerminalCase
	// ActiveJobsCount 是启用中的流水线数。
	ActiveJobsCount int
	// RunningExecutionsCount 是当前运行中的 execution 数。
	RunningExecutionsCount int
	Now                    time.Time
}

// BuildSummarySnapshot 由聚合输入投影出全局摘要快照。
//
// 通过率口径（与 execution 的 RefreshCounts 一致）：
// 成功 / (成功+失败+跳过)；无已结束用例时为 0。
// 平均耗时：近 7 天终止态 execution 的 duration 均值；无则 nil。
func BuildSummarySnapshot(in SummaryInput) SummarySnapshot {
	var success, failure, skipped int
	for _, c := range in.TerminalCases {
		switch CaseStatusCategory(c.Status) {
		case "success":
			success++
		case "failure":
			failure++
		case "skipped":
			skipped++
		}
	}

	finished := success + failure + skipped
	rate := 0.0
	if finished > 0 {
		rate = float64(success) / float64(finished)
	}

	var avgDuration *float64
	if len(in.TerminalExecutions) > 0 {
		var sum float64
		var count int
		for _, e := range in.TerminalExecutions {
			if e.Duration != nil {
				sum += *e.Duration
				count++
			}
		}
		if count > 0 {
			v := sum / float64(count)
			avgDuration = &v
		}
	}

	return SummarySnapshot{
		SnapshotKey:            SnapshotKey,
		ActiveJobsCount:        in.ActiveJobsCount,
		TotalExecutionsCount:   len(in.TerminalExecutions),
		RunningExecutionsCount: in.RunningExecutionsCount,
		SuccessCases7d:         success,
		FailureCases7d:         failure,
		SkippedCases7d:         skipped,
		PassRate7d:             rate,
		AvgExecutionDuration7d: avgDuration,
		UpdatedAt:              in.Now,
	}
}

// DayKey 返回时间戳的日期键（用于趋势分组与空日期补齐）。
func DayKey(t time.Time) time.Time {
	y, m, d := t.Date()
	return time.Date(y, m, d, 0, 0, 0, 0, t.Location())
}

// BuildTrendSnapshots 由近 N 天的终止态用例聚合出每日趋势，并补齐缺失日期。
//
// 每个 execution 每天计一次 execution_total（按 started_at 归属当天），
// 每个用例按 (execution, case_key) 在当天计一次状态。
func BuildTrendSnapshots(
	cases []TerminalCase,
	executions []TerminalExecution,
	end time.Time, days int,
) []TrendSnapshot {
	// 逐日累计
	byDay := make(map[time.Time]*TrendSnapshot)
	// execution 去重：同一 execution 每天只计一次 execution_total
	execSeen := make(map[string]bool)
	// case 去重：同一 execution + case_key 每天只计一次
	caseSeen := make(map[string]bool)

	for _, e := range executions {
		day := DayKey(e.StartedAt)
		ts := trendFor(byDay, day)
		key := strconv.FormatInt(e.ExecutionID, 10) + ":" + day.Format("2006-01-02")
		if !execSeen[key] {
			ts.ExecutionTotal++
			execSeen[key] = true
		}
	}

	for _, c := range cases {
		if c.StartedAt == nil {
			continue
		}
		day := DayKey(*c.StartedAt)
		ts := trendFor(byDay, day)
		key := strconv.FormatInt(c.ExecutionID, 10) + ":" + c.CaseKey + ":" + day.Format("2006-01-02")
		if caseSeen[key] {
			continue
		}
		caseSeen[key] = true
		switch CaseStatusCategory(c.Status) {
		case "success":
			ts.SuccessCases++
		case "failure":
			ts.FailureCases++
		case "skipped":
			ts.SkippedCases++
		}
	}

	// 补齐缺失日期（从 end-days+1 到 end）
	out := make([]TrendSnapshot, 0, days)
	start := DayKey(end).AddDate(0, 0, -(days - 1))
	for i := 0; i < days; i++ {
		d := start.AddDate(0, 0, i)
		if ts, ok := byDay[d]; ok {
			out = append(out, *ts)
		} else {
			out = append(out, TrendSnapshot{StatDate: d})
		}
	}
	return out
}

func trendFor(m map[time.Time]*TrendSnapshot, day time.Time) *TrendSnapshot {
	if ts, ok := m[day]; ok {
		return ts
	}
	ts := &TrendSnapshot{StatDate: day}
	m[day] = ts
	return ts
}
