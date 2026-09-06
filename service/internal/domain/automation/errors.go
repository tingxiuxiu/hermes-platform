package automation

import "github.com/hermes-platform/go-service/internal/platform/errors"

// 本文件集中声明 automation 上下文的领域错误（任务 T-3.5）。

// ---- Execution ----

// ErrExecutionNotFound 执行不存在。
var ErrExecutionNotFound = errors.New(errors.KindNotFound, "执行记录不存在")

// ErrInvalidBuildUID build_uid 缺失或非法。
var ErrInvalidBuildUID = errors.New(errors.KindValidation, "build_uid 必须是合法的 UUID")

// ErrEndTimeBeforeStart 结束时间早于开始时间。
var ErrEndTimeBeforeStart = errors.New(errors.KindValidation, "结束时间不能早于开始时间")

// ErrJobNameRequired 任务名称必填。
var ErrJobNameRequired = errors.New(errors.KindValidation, "job_name 不能为空")

// ErrJobNameTooLong 任务名称超长。
var ErrJobNameTooLong = errors.New(errors.KindValidation, "job_name 长度不能超过 255 个字符")

// ErrProjectNameRequired 项目名称必填。
var ErrProjectNameRequired = errors.New(errors.KindValidation, "project_name 不能为空")

// ErrSoftwareNameRequired 被测软件名称必填。
var ErrSoftwareNameRequired = errors.New(errors.KindValidation, "software_name 不能为空")

// ErrSoftwareVersionRequired 被测软件版本必填。
var ErrSoftwareVersionRequired = errors.New(errors.KindValidation, "software_version 不能为空")

// ---- Item ----

// ErrItemNotFound 用例不存在。
var ErrItemNotFound = errors.New(errors.KindNotFound, "测试用例不存在")

// ErrInvalidCaseUID case_uid 缺失或非法。
var ErrInvalidCaseUID = errors.New(errors.KindValidation, "case_uid 必须是合法的 UUID")

// ErrCaseKeyRequired case_key 必填。
var ErrCaseKeyRequired = errors.New(errors.KindValidation, "case_key 不能为空")

// ErrCaseNameRequired case_name 必填。
var ErrCaseNameRequired = errors.New(errors.KindValidation, "case_name 不能为空")

// ErrInvalidAttemptNumber attempt 序号必须 ≥ 1。
var ErrInvalidAttemptNumber = errors.New(errors.KindValidation, "重试序号必须大于等于 1")

// ---- Step ----

// ErrStepNotFound 步骤不存在。
var ErrStepNotFound = errors.New(errors.KindNotFound, "测试步骤不存在")

// ErrInvalidStepPath 步骤路径格式不合法。
var ErrInvalidStepPath = errors.New(errors.KindValidation,
	"step_path 格式不合法，应形如 0 / 0.1 / 0.1.2")

// ErrStepNameRequired 步骤名称必填。
var ErrStepNameRequired = errors.New(errors.KindValidation, "step_name 不能为空")

// ErrStepPathTooDeep 步骤层级超出限制。
var ErrStepPathTooDeep = errors.New(errors.KindValidation, "步骤层级不能超过 5 层")

// ---- Attachment ----

// ErrInvalidAttachmentType 附件类型不合法。
var ErrInvalidAttachmentType = errors.New(errors.KindValidation,
	"附件类型必须是 screenshot / log / video / other 之一")

// ErrAttachmentURLRequired 附件 URL 必填。
var ErrAttachmentURLRequired = errors.New(errors.KindValidation, "附件 url 不能为空")

// ErrAttachmentOwnerRequired 附件必须且只能归属于一个用例或一个步骤（评审确认项 Q1）。
var ErrAttachmentOwnerRequired = errors.New(errors.KindValidation,
	"附件必须且只能指定归属之一：item_id（用例级）或 step_id（步骤级）")

// ---- 通用 ----

// ErrPipelineNotFound 流水线不存在。
var ErrPipelineNotFound = errors.New(errors.KindNotFound, "流水线不存在")

// ErrInvalidStatus 状态值不合法，message 需带上具体取值。
func ErrInvalidStatus(field, value string) *errors.Error {
	return errors.Errorf(errors.KindValidation, "非法的 %s 状态: %q", field, value)
}

// ErrPipelineNotFoundByJobName 带 job_name 的流水线不存在。
func ErrPipelineNotFoundByJobName(jobName string) *errors.Error {
	return errors.Errorf(errors.KindNotFound, "流水线 %s 不存在", jobName)
}

// ErrExecutionNotFoundByUID 带 build_uid 的执行不存在。
func ErrExecutionNotFoundByUID(buildUID string) *errors.Error {
	return errors.Errorf(errors.KindNotFound, "执行记录 %s 不存在", buildUID)
}

// ErrItemNotFoundByUID 带 case_uid 的用例不存在。
func ErrItemNotFoundByUID(caseUID string) *errors.Error {
	return errors.Errorf(errors.KindNotFound, "测试用例 %s 不存在", caseUID)
}
