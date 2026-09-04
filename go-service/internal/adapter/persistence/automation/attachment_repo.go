package automation

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"

	appautomation "github.com/hermes-platform/go-service/internal/application/automation"
	"github.com/hermes-platform/go-service/internal/domain/automation"
	"github.com/hermes-platform/go-service/internal/platform/database"
)

// AttachmentRepo 实现 application/automation.AttachmentRepository 端口。
type AttachmentRepo struct {
	db *database.Pool
}

// NewAttachmentRepo 构造附件仓储。
func NewAttachmentRepo(db *database.Pool) *AttachmentRepo {
	return &AttachmentRepo{db: db}
}

var _ appautomation.AttachmentRepository = (*AttachmentRepo)(nil)

const attachmentColumns = `id, item_id, step_id, attachment_type, file_name, url, mime_type, created_at, updated_at`

// Create 写入一条附件记录。
//
// 归属（Q1）：item_id 与 step_id 互斥，由领域实体保证；这里按快照落库。
func (r *AttachmentRepo) Create(ctx context.Context, a *automation.Attachment) (int64, error) {
	var id int64
	err := database.Executor(ctx, r.db).QueryRow(ctx, `
		INSERT INTO execution_case_step_attachments (
			step_id, item_id, attachment_type, file_name, url, mime_type
		)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id`,
		a.StepID(), a.ItemID(), string(a.AttachmentType()),
		nullIfEmpty(a.FileName()), a.URL(), nullIfEmpty(a.MimeType()),
	).Scan(&id)
	if err != nil {
		return 0, fmt.Errorf("automation: create attachment: %w", err)
	}
	return id, nil
}

// ListByItemID 返回某用例的全部附件（item_id 匹配，step_id 为 nil）。
func (r *AttachmentRepo) ListByItemID(ctx context.Context, itemID int64) ([]*automation.Attachment, error) {
	return r.list(ctx, `WHERE item_id = $1 ORDER BY id ASC`, itemID)
}

// ListByStepID 返回某步骤的全部附件（step_id 匹配）。
func (r *AttachmentRepo) ListByStepID(ctx context.Context, stepID int64) ([]*automation.Attachment, error) {
	return r.list(ctx, `WHERE step_id = $1 ORDER BY id ASC`, stepID)
}

// ListByStepIDs 返回多个步骤的全部附件。
func (r *AttachmentRepo) ListByStepIDs(ctx context.Context, stepIDs []int64) ([]*automation.Attachment, error) {
	if len(stepIDs) == 0 {
		return nil, nil
	}
	return r.list(ctx, `WHERE step_id = ANY($1) ORDER BY id ASC`, stepIDs)
}

func (r *AttachmentRepo) list(ctx context.Context, where string, args ...any) ([]*automation.Attachment, error) {
	rows, err := database.Executor(ctx, r.db).Query(ctx,
		`SELECT `+attachmentColumns+` FROM execution_case_step_attachments `+where, args...)
	if err != nil {
		return nil, fmt.Errorf("automation: list attachments: %w", err)
	}
	defer rows.Close()

	var out []*automation.Attachment
	for rows.Next() {
		a, err := scanAttachment(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, a)
	}
	return out, rows.Err()
}

func scanAttachment(row pgx.Row) (*automation.Attachment, error) {
	var (
		id             int64
		itemID         *int64
		stepID         *int64
		attachmentType string
		fileName       *string
		url            string
		mimeType       *string
		createdAt      time.Time
		updatedAt      time.Time
	)

	err := row.Scan(&id, &itemID, &stepID, &attachmentType,
		&fileName, &url, &mimeType, &createdAt, &updatedAt)
	if err != nil {
		return nil, fmt.Errorf("automation: scan attachment: %w", err)
	}

	fn := ""
	if fileName != nil {
		fn = *fileName
	}
	mt := ""
	if mimeType != nil {
		mt = *mimeType
	}

	return automation.RestoreAttachment(automation.AttachmentSnapshot{
		ID:             id,
		ItemID:         itemID,
		StepID:         stepID,
		AttachmentType: attachmentType,
		FileName:       fn,
		URL:            url,
		MimeType:       mt,
		CreatedAt:      createdAt,
		UpdatedAt:      updatedAt,
	}), nil
}
