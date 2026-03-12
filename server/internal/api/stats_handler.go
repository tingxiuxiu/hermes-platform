package api

import (
	"com.hermes.platform/internal/services"
	"github.com/gin-gonic/gin"
)

type StatsHandler struct {
	statsService services.StatsService
}

func NewStatsHandler(statsService services.StatsService) *StatsHandler {
	return &StatsHandler{
		statsService: statsService,
	}
}

func (h *StatsHandler) GetDashboardStats(c *gin.Context) {
	stats, err := h.statsService.GetDashboardStats()
	if err != nil {
		InternalServerError(c, "Failed to get dashboard stats")
		return
	}

	SuccessWithData(c, gin.H{
		"stats": stats,
	})
}

func (h *StatsHandler) GetTrendData(c *gin.Context) {
	rangeType := c.DefaultQuery("range", "week")

	trend, err := h.statsService.GetTrendData(rangeType)
	if err != nil {
		InternalServerError(c, "Failed to get trend data")
		return
	}

	SuccessWithData(c, gin.H{
		"trend": trend,
	})
}

func (h *StatsHandler) GetRunningTasks(c *gin.Context) {
	tasks, err := h.statsService.GetRunningTasks()
	if err != nil {
		InternalServerError(c, "Failed to get running tasks")
		return
	}

	SuccessWithData(c, gin.H{
		"tasks": tasks,
	})
}
