package automation

import (
	"time"

	"github.com/hermes-platform/go-service/internal/domain/automation"
)

// 领域对象 → DTO 的映射。
// 所有响应字段与 Python constants 一一对应，null 语义用指针表达。

// ToExecutionRow 把领域 execution 转成响应 DTO。
func ToExecutionRow(e *automation.TestExecution) ExecutionRow {
	planned := e.PlannedCasesCount()
	row := ExecutionRow{
		ID:              e.ID(),
		BuildUID:        e.BuildUID().String(),
		JobName:         e.JobName(),
		JobURL:          nullIfEmpty(e.JobURL()),
		ProjectName:     e.ProjectName(),
		SoftwareName:    e.SoftwareName(),
		SoftwareVersion: e.SoftwareVersion(),
		Labels:          nilOrList(e.Labels()),
		Status:          string(e.Status()),
		StartTime:       e.StartTime(),
		EndTime:         e.EndTime(),
		Duration:        e.Duration(),
		PassRate:        e.PassRate(),
		LastHeartbeatAt: e.LastHeartbeatAt(),
	}
	if planned > 0 {
		plannedCopy := planned
		row.PlannedCasesCount = &plannedCopy
	}
	if e.PassCount() > 0 || e.FailureCount() > 0 || e.SkippedCount() > 0 {
		pc, fc, sc := e.PassCount(), e.FailureCount(), e.SkippedCount()
		row.PassCount, row.FailureCount, row.SkippedCount = &pc, &fc, &sc
	}
	return row
}

// ToItemSummary 把领域用例转成不含步骤/附件的摘要。
func ToItemSummary(item *automation.ExecutionItem) ItemSummaryRow {
	return ItemSummaryRow{
		ID:            item.ID(),
		CaseUID:       item.CaseUID().String(),
		CaseKey:       item.CaseKey(),
		CaseName:      item.CaseName(),
		AttemptNumber: item.AttemptNumber(),
		Status:        string(item.Status()),
		StartTime:     item.StartTime(),
		EndTime:       item.EndTime(),
		Duration:      item.Duration(),
	}
}

// ToLiveCurrentItem 把带步骤树的 running 用例转成 live 主卡片。
func ToLiveCurrentItem(item *automation.ExecutionItem) LiveCurrentItem {
	row := LiveCurrentItem{
		CaseUID:  item.CaseUID().String(),
		CaseKey:  item.CaseKey(),
		CaseName: item.CaseName(),
		Status:   string(item.Status()),
	}
	if steps := item.Steps(); len(steps) > 0 {
		row.Steps = toStepRows(steps)
	}
	return row
}

// ToExecutionItemRow 把领域用例转成响应 DTO（含附件与嵌套步骤树）。
func ToExecutionItemRow(item *automation.ExecutionItem) ExecutionItemRow {
	row := ExecutionItemRow{
		ID:             item.ID(),
		BuildUID:       item.BuildUID().String(),
		CaseUID:        item.CaseUID().String(),
		CaseKey:        item.CaseKey(),
		CaseName:       item.CaseName(),
		Labels:         nilOrList(item.Labels()),
		AttemptNumber:  item.AttemptNumber(),
		IsLatest:       item.IsLatest(),
		Status:         string(item.Status()),
		StartTime:      item.StartTime(),
		EndTime:        item.EndTime(),
		Duration:       item.Duration(),
		ErrorMessage:   nullIfEmpty(item.ErrorMessage()),
		ErrorTraceback: nullIfEmpty(item.ErrorTraceback()),
	}

	if atts := item.Attachments(); len(atts) > 0 {
		row.Attachments = toAttachmentRows(atts)
	}

	if steps := item.Steps(); len(steps) > 0 {
		_, orphanCount := automation.BuildStepTree(steps)
		_ = orphanCount // 查询用例已填充 steps，这里只负责序列化
		rows := toStepRows(steps)
		row.Steps = rows
	}
	return row
}

// toStepRows 把平铺步骤转成嵌套的 StepRow（利用 BuildStepTree 组装）。
func toStepRows(steps []*automation.ExecutionStep) []StepRow {
	roots, _ := automation.BuildStepTree(steps)
	out := make([]StepRow, 0, len(roots))
	for _, root := range roots {
		out = append(out, toStepRowNode(root))
	}
	return out
}

func toStepRowNode(node automation.StepNode) StepRow {
	row := StepRow{
		ID:              node.Step.ID(),
		CaseUID:         node.Step.CaseUID().String(),
		ParentStepID:    node.Step.ParentStepID(),
		ParentStepIndex: node.Step.ParentStepIndex(),
		StepIndex:       node.Step.StepIndex(),
		StepPath:        node.Step.StepPath().String(),
		Depth:           node.Step.Depth(),
		StepName:        node.Step.StepName(),
		Status:          string(node.Step.Status()),
		StartTime:       node.Step.StartTime(),
		EndTime:         node.Step.EndTime(),
		Duration:        node.Step.Duration(),
	}
	if atts := node.Step.Attachments(); len(atts) > 0 {
		row.Attachments = toAttachmentRows(atts)
	}
	if len(node.Children) > 0 {
		sub := make([]StepRow, 0, len(node.Children))
		for _, child := range node.Children {
			sub = append(sub, toStepRowNode(child))
		}
		row.SubSteps = sub
	}
	return row
}

// ToPipelineRow 把领域流水线转成响应 DTO。
func ToPipelineRow(p *automation.JenkinsPipeline) PipelineRow {
	row := PipelineRow{
		ID:                 p.ID(),
		JobName:            p.JobName(),
		JobURL:             p.JobURL(),
		Status:             string(p.Status()),
		LastBuildNumber:    p.LastBuildNumber(),
		LastBuildTimestamp: p.LastBuildTimestamp(),
		LastBuildDuration:  p.LastBuildDuration(),
		PipelineParams:     p.PipelineParams(),
		SyncAt:             p.SyncAt(),
	}
	if uid := p.LastBuildUID(); uid != nil {
		s := uid.String()
		row.LastBuildUID = &s
	}
	if bs := p.LastBuildStatus(); bs != "" {
		s := string(bs)
		row.LastBuildStatus = &s
	}
	return row
}

func toAttachmentRows(atts []*automation.Attachment) []AttachmentItemRow {
	out := make([]AttachmentItemRow, 0, len(atts))
	for _, a := range atts {
		row := AttachmentItemRow{
			AttachmentType: string(a.AttachmentType()),
			FileName:       a.FileName(),
			URL:            a.URL(),
		}
		if mt := a.MimeType(); mt != "" {
			row.MimeType = &mt
		}
		out = append(out, row)
	}
	return out
}

func nullIfEmpty(s string) *string {
	if s == "" {
		return nil
	}
	v := s
	return &v
}

func nilOrList(items []string) []string {
	if len(items) == 0 {
		return nil
	}
	return items
}

var _ = time.Time{}
