package automation

import (
	"strings"
	"time"
)

// AttachmentType 是附件类型。
//
// 取值与 Python 的 `AttachmentType = Literal["screenshot","log","video","other"]` 一致。
type AttachmentType string

const (
	AttachmentTypeScreenshot AttachmentType = "screenshot"
	AttachmentTypeLog        AttachmentType = "log"
	AttachmentTypeVideo      AttachmentType = "video"
	AttachmentTypeOther      AttachmentType = "other"
)

// Valid 判断附件类型是否合法。
func (t AttachmentType) Valid() bool {
	switch t {
	case AttachmentTypeScreenshot, AttachmentTypeLog,
		AttachmentTypeVideo, AttachmentTypeOther:
		return true
	default:
		return false
	}
}

// ParseAttachmentType 解析附件类型。
func ParseAttachmentType(raw string) (AttachmentType, error) {
	t := AttachmentType(strings.TrimSpace(raw))
	if !t.Valid() {
		return "", ErrInvalidAttachmentType
	}
	return t, nil
}

// AttachmentSpec 是上报接口中附件字段的输入形态（不含 id / 归属）。
// 归属（item 级 / step 级）由所在请求决定，映射时再指定。
type AttachmentSpec struct {
	Type     AttachmentType
	FileName string
	URL      string
	MimeType string
}

// Attachment 是一条附件记录。
//
// 服务**只登记 URL，不接收文件内容**（ADR-0001）：
// 附件由插件自行上传到对象存储后把 URL 报上来，
// Go 侧不做文件 I/O，也不关心存储后端。
//
// 归属（评审确认项 Q1）：
//   - case 级附件记 ItemID，StepID 为 nil；
//   - step 级附件记 StepID，ItemID 为 nil；
//   - 两者互斥，构造时校验，避免写入语义不明的数据。
type Attachment struct {
	id             int64
	itemID         *int64
	stepID         *int64
	attachmentType AttachmentType
	fileName       string
	url            string
	mimeType       string
	createdAt      time.Time
	updatedAt      time.Time
}

// AttachmentSnapshot 用于从数据库重建附件。
type AttachmentSnapshot struct {
	ID             int64
	ItemID         *int64
	StepID         *int64
	AttachmentType string
	FileName       string
	URL            string
	MimeType       string
	CreatedAt      time.Time
	UpdatedAt      time.Time
}

// NewAttachmentCommand 是创建附件所需的输入。
type NewAttachmentCommand struct {
	ItemID   *int64
	StepID   *int64
	Type     AttachmentType
	FileName string
	URL      string
	MimeType string
}

// NewAttachment 创建一条附件记录。
func NewAttachment(cmd NewAttachmentCommand) (*Attachment, error) {
	if !cmd.Type.Valid() {
		return nil, ErrInvalidAttachmentType
	}

	url := strings.TrimSpace(cmd.URL)
	if url == "" {
		return nil, ErrAttachmentURLRequired
	}

	// 归属互斥：必须且只能指定一个
	if (cmd.ItemID == nil) == (cmd.StepID == nil) {
		return nil, ErrAttachmentOwnerRequired
	}

	return &Attachment{
		itemID:         cmd.ItemID,
		stepID:         cmd.StepID,
		attachmentType: cmd.Type,
		fileName:       strings.TrimSpace(cmd.FileName),
		url:            url,
		mimeType:       strings.TrimSpace(cmd.MimeType),
	}, nil
}

// RestoreAttachment 从数据库重建实体。
func RestoreAttachment(s AttachmentSnapshot) *Attachment {
	t := AttachmentType(s.AttachmentType)
	if !t.Valid() {
		t = AttachmentTypeOther
	}
	return &Attachment{
		id:             s.ID,
		itemID:         s.ItemID,
		stepID:         s.StepID,
		attachmentType: t,
		fileName:       s.FileName,
		url:            s.URL,
		mimeType:       s.MimeType,
		createdAt:      s.CreatedAt,
		updatedAt:      s.UpdatedAt,
	}
}

func (a *Attachment) ID() int64                      { return a.id }
func (a *Attachment) ItemID() *int64                 { return a.itemID }
func (a *Attachment) StepID() *int64                 { return a.stepID }
func (a *Attachment) AttachmentType() AttachmentType { return a.attachmentType }
func (a *Attachment) FileName() string               { return a.fileName }
func (a *Attachment) URL() string                    { return a.url }
func (a *Attachment) MimeType() string               { return a.mimeType }
func (a *Attachment) CreatedAt() time.Time           { return a.createdAt }
func (a *Attachment) UpdatedAt() time.Time           { return a.updatedAt }
