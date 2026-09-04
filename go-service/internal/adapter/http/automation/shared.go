package automation

import (
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	dtoautomation "github.com/hermes-platform/go-service/internal/adapter/http/dto/automation"
	domainautomation "github.com/hermes-platform/go-service/internal/domain/automation"
	"github.com/hermes-platform/go-service/internal/platform/response"
)

// pathUUID 解析路径参数中的 UUID。
func pathUUID(c *gin.Context, name string) (uuid.UUID, bool) {
	u, err := parseUUID(c.Param(name))
	if err != nil {
		response.Fail(c, 422, "路径参数不合法")
		return uuid.Nil, false
	}
	return u, true
}

// pathID 解析路径参数中的正整数 ID。
func pathID(c *gin.Context, name string) (int64, bool) {
	id, err := strconv.ParseInt(c.Param(name), 10, 64)
	if err != nil || id <= 0 {
		response.Fail(c, 422, "路径参数不合法")
		return 0, false
	}
	return id, true
}

func parseUUID(raw string) (uuid.UUID, error) {
	return uuid.Parse(raw)
}

// uuidFromString 把字符串解析为 UUID，失败返回 Nil。
func uuidFromString(raw string) uuid.UUID {
	u, _ := uuid.Parse(raw)
	return u
}

func derefTime(t *time.Time) time.Time {
	if t == nil {
		return time.Time{}
	}
	return *t
}

func defaultPage(page, pageSize int) (int, int) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 10
	}
	return page, pageSize
}

// toAttachmentSpecs 把请求 DTO 的附件列表转成领域 AttachmentSpec。
func toAttachmentSpecs(items []dtoautomation.AttachmentItemRequest) []domainautomation.AttachmentSpec {
	if len(items) == 0 {
		return nil
	}
	out := make([]domainautomation.AttachmentSpec, 0, len(items))
	for _, a := range items {
		out = append(out, domainautomation.AttachmentSpec{
			Type:     domainautomation.AttachmentType(a.AttachmentType),
			FileName: a.FileName,
			URL:      a.URL,
			MimeType: derefStr(a.MimeType),
		})
	}
	return out
}

func derefStr(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}
