package automation

import (
	"strings"
	"time"

	"github.com/google/uuid"
)

// PipelineStatus 是流水线的状态。
type PipelineStatus string

const (
	PipelineStatusActive   PipelineStatus = "active"
	PipelineStatusInactive PipelineStatus = "inactive"
)

// Valid 判断状态值是否合法。
func (s PipelineStatus) Valid() bool {
	switch s {
	case PipelineStatusActive, PipelineStatusInactive:
		return true
	default:
		return false
	}
}

// BuildStatus 是最后一次构建的状态。
type BuildStatus string

const (
	BuildStatusSuccess    BuildStatus = "success"
	BuildStatusFailure    BuildStatus = "failure"
	BuildStatusAborted    BuildStatus = "aborted"
	BuildStatusInProgress BuildStatus = "in_progress"
	BuildStatusUnstable   BuildStatus = "unstable"
	BuildStatusNotBuilt   BuildStatus = "not_built"
)

// Valid 判断构建状态是否合法。
func (s BuildStatus) Valid() bool {
	switch s {
	case BuildStatusSuccess, BuildStatusFailure, BuildStatusAborted,
		BuildStatusInProgress, BuildStatusUnstable, BuildStatusNotBuilt:
		return true
	default:
		return false
	}
}

// JenkinsPipeline 是一条 Jenkins 流水线。
type JenkinsPipeline struct {
	id                 int64
	jobName            string
	jobURL             string
	status             PipelineStatus
	lastBuildNumber    *int
	lastBuildUID       *uuid.UUID
	lastBuildStatus    BuildStatus
	lastBuildTimestamp *time.Time
	lastBuildDuration  *float64
	pipelineParams     map[string]any
	syncAt             *time.Time
	createdAt          time.Time
	updatedAt          time.Time
}

// PipelineSnapshot 用于从数据库重建流水线。
type PipelineSnapshot struct {
	ID                 int64
	JobName            string
	JobURL             string
	Status             string
	LastBuildNumber    *int
	LastBuildUID       *uuid.UUID
	LastBuildStatus    string
	LastBuildTimestamp *time.Time
	LastBuildDuration  *float64
	PipelineParams     map[string]any
	SyncAt             *time.Time
	CreatedAt          time.Time
	UpdatedAt          time.Time
}

// NewPipeline 创建一条流水线。
func NewPipeline(jobName, jobURL string) (*JenkinsPipeline, error) {
	p := &JenkinsPipeline{}
	if err := p.setJobName(jobName); err != nil {
		return nil, err
	}
	p.jobURL = strings.TrimSpace(jobURL)
	p.status = PipelineStatusActive
	return p, nil
}

// RestorePipeline 从数据库重建流水线。
func RestorePipeline(s PipelineSnapshot) *JenkinsPipeline {
	status := PipelineStatus(s.Status)
	if !status.Valid() {
		status = PipelineStatusActive
	}
	buildStatus := BuildStatus(s.LastBuildStatus)
	if !buildStatus.Valid() {
		buildStatus = ""
	}
	return &JenkinsPipeline{
		id:                 s.ID,
		jobName:            s.JobName,
		jobURL:             s.JobURL,
		status:             status,
		lastBuildNumber:    s.LastBuildNumber,
		lastBuildUID:       s.LastBuildUID,
		lastBuildStatus:    buildStatus,
		lastBuildTimestamp: s.LastBuildTimestamp,
		lastBuildDuration:  s.LastBuildDuration,
		pipelineParams:     s.PipelineParams,
		syncAt:             s.SyncAt,
		createdAt:          s.CreatedAt,
		updatedAt:          s.UpdatedAt,
	}
}

func (p *JenkinsPipeline) ID() int64                      { return p.id }
func (p *JenkinsPipeline) JobName() string                { return p.jobName }
func (p *JenkinsPipeline) JobURL() string                 { return p.jobURL }
func (p *JenkinsPipeline) Status() PipelineStatus         { return p.status }
func (p *JenkinsPipeline) LastBuildNumber() *int          { return p.lastBuildNumber }
func (p *JenkinsPipeline) LastBuildUID() *uuid.UUID       { return p.lastBuildUID }
func (p *JenkinsPipeline) LastBuildStatus() BuildStatus   { return p.lastBuildStatus }
func (p *JenkinsPipeline) LastBuildTimestamp() *time.Time { return p.lastBuildTimestamp }
func (p *JenkinsPipeline) LastBuildDuration() *float64    { return p.lastBuildDuration }
func (p *JenkinsPipeline) PipelineParams() map[string]any { return p.pipelineParams }
func (p *JenkinsPipeline) SyncAt() *time.Time             { return p.syncAt }
func (p *JenkinsPipeline) CreatedAt() time.Time           { return p.createdAt }
func (p *JenkinsPipeline) UpdatedAt() time.Time           { return p.updatedAt }

// UpdateLastBuild 用一次执行的结果刷新最后构建信息。
func (p *JenkinsPipeline) UpdateLastBuild(e *TestExecution) {
	status := BuildStatusFromExecution(e.Status())
	p.lastBuildStatus = status
	p.lastBuildUID = &e.buildUID
	p.lastBuildTimestamp = &e.startTime
	if e.duration != nil {
		d := *e.duration
		p.lastBuildDuration = &d
	}
	now := time.Now().UTC()
	p.syncAt = &now
}

// BuildStatusFromExecution 把 execution 状态映射为流水线的构建状态。
func BuildStatusFromExecution(s ExecutionStatus) BuildStatus {
	switch s {
	case ExecutionStatusCompleted:
		return BuildStatusSuccess
	case ExecutionStatusFailed:
		return BuildStatusFailure
	case ExecutionStatusAborted:
		return BuildStatusAborted
	default:
		return BuildStatusInProgress
	}
}

func (p *JenkinsPipeline) setJobName(v string) error {
	v = strings.TrimSpace(v)
	if v == "" {
		return ErrJobNameRequired
	}
	if len(v) > 255 {
		return ErrJobNameTooLong
	}
	p.jobName = v
	return nil
}
