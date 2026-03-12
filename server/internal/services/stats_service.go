package services

import (
	"time"

	"com.hermes.platform/internal/models"
	"gorm.io/gorm"
)

type TimeRangeStats struct {
	TotalCases   int     `json:"total_cases"`
	TotalTasks   int     `json:"total_tasks"`
	PassedCases  int     `json:"passed_cases"`
	FailedCases  int     `json:"failed_cases"`
	SkippedCases int     `json:"skipped_cases"`
	PassRate     float64 `json:"pass_rate"`
}

type DashboardStats struct {
	Today   TimeRangeStats `json:"today"`
	Week    TimeRangeStats `json:"week"`
	Month   TimeRangeStats `json:"month"`
	Year    TimeRangeStats `json:"year"`
	AllTime TimeRangeStats `json:"all_time"`
}

type TrendData struct {
	Date        string  `json:"date"`
	TotalCases  int     `json:"total_cases"`
	PassedCases int     `json:"passed_cases"`
	FailedCases int     `json:"failed_cases"`
	PassRate    float64 `json:"pass_rate"`
}

type RunningTask struct {
	ID               uint    `json:"id"`
	TaskName         string  `json:"task_name"`
	WorkerName       string  `json:"worker_name"`
	PlanKey          string  `json:"plan_key"`
	StartTime        int64   `json:"start_time"`
	Progress         float64 `json:"progress"`
	EstimatedEndTime int64   `json:"estimated_end_time"`
	PassedTests      int     `json:"passed_tests"`
	FailedTests      int     `json:"failed_tests"`
	TotalTests       int     `json:"total_tests"`
}

type StatsService interface {
	GetDashboardStats() (*DashboardStats, error)
	GetTrendData(rangeType string) ([]TrendData, error)
	GetRunningTasks() ([]RunningTask, error)
}

type statsService struct {
	db *gorm.DB
}

func NewStatsService(db *gorm.DB) StatsService {
	return &statsService{
		db: db,
	}
}

func (s *statsService) GetDashboardStats() (*DashboardStats, error) {
	now := time.Now()

	todayStart := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	weekStart := todayStart.AddDate(0, 0, -int(todayStart.Weekday()))
	monthStart := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())
	yearStart := time.Date(now.Year(), 1, 1, 0, 0, 0, 0, now.Location())

	stats := &DashboardStats{
		Today:   s.getStatsForRange(todayStart.Unix(), now.Unix()),
		Week:    s.getStatsForRange(weekStart.Unix(), now.Unix()),
		Month:   s.getStatsForRange(monthStart.Unix(), now.Unix()),
		Year:    s.getStatsForRange(yearStart.Unix(), now.Unix()),
		AllTime: s.getAllTimeStats(),
	}

	return stats, nil
}

func (s *statsService) getStatsForRange(startTime, endTime int64) TimeRangeStats {
	var stats TimeRangeStats

	var tasks []models.TestTask
	s.db.Model(&models.TestTask{}).
		Where("start_time >= ? AND start_time <= ?", startTime, endTime).
		Find(&tasks)

	stats.TotalTasks = len(tasks)
	for _, task := range tasks {
		stats.TotalCases += task.TotalTests
		stats.PassedCases += task.PassedTests
		stats.FailedCases += task.FailedTests
	}

	var skippedCount int64
	s.db.Model(&models.TestDetail{}).
		Joins("JOIN test_tasks ON test_details.test_task_id = test_tasks.id").
		Where("test_tasks.start_time >= ? AND test_tasks.start_time <= ?", startTime, endTime).
		Where("test_details.test_status = ?", "skipped").
		Count(&skippedCount)
	stats.SkippedCases = int(skippedCount)

	if stats.TotalCases > 0 {
		stats.PassRate = float64(stats.PassedCases) / float64(stats.TotalCases) * 100
	}

	return stats
}

func (s *statsService) getAllTimeStats() TimeRangeStats {
	var stats TimeRangeStats

	var taskStats struct {
		TotalTasks  int
		TotalCases  int
		PassedCases int
		FailedCases int
	}

	s.db.Model(&models.TestTask{}).
		Select("COUNT(*) as total_tasks, COALESCE(SUM(total_tests), 0) as total_cases, COALESCE(SUM(passed_tests), 0) as passed_cases, COALESCE(SUM(failed_tests), 0) as failed_cases").
		Scan(&taskStats)

	stats.TotalTasks = taskStats.TotalTasks
	stats.TotalCases = taskStats.TotalCases
	stats.PassedCases = taskStats.PassedCases
	stats.FailedCases = taskStats.FailedCases

	var skippedCount int64
	s.db.Model(&models.TestDetail{}).
		Where("test_status = ?", "skipped").
		Count(&skippedCount)
	stats.SkippedCases = int(skippedCount)

	if stats.TotalCases > 0 {
		stats.PassRate = float64(stats.PassedCases) / float64(stats.TotalCases) * 100
	}

	return stats
}

func (s *statsService) GetTrendData(rangeType string) ([]TrendData, error) {
	now := time.Now()
	var startTime time.Time
	var dateFormat string

	switch rangeType {
	case "today":
		startTime = time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
		dateFormat = "2006-01-02 15:04"
	case "week":
		startTime = now.AddDate(0, 0, -7)
		dateFormat = "2006-01-02"
	case "month":
		startTime = now.AddDate(0, -1, 0)
		dateFormat = "2006-01-02"
	case "year":
		startTime = now.AddDate(-1, 0, 0)
		dateFormat = "2006-01"
	default:
		startTime = now.AddDate(0, 0, -7)
		dateFormat = "2006-01-02"
	}

	var tasks []models.TestTask
	s.db.Model(&models.TestTask{}).
		Where("start_time >= ?", startTime.Unix()).
		Order("start_time asc").
		Find(&tasks)

	trendMap := make(map[string]*TrendData)
	for _, task := range tasks {
		dateKey := time.Unix(task.StartTime, 0).Format(dateFormat)
		if _, exists := trendMap[dateKey]; !exists {
			trendMap[dateKey] = &TrendData{
				Date: dateKey,
			}
		}
		trendMap[dateKey].TotalCases += task.TotalTests
		trendMap[dateKey].PassedCases += task.PassedTests
		trendMap[dateKey].FailedCases += task.FailedTests
	}

	var trendData []TrendData
	for _, data := range trendMap {
		if data.TotalCases > 0 {
			data.PassRate = float64(data.PassedCases) / float64(data.TotalCases) * 100
		}
		trendData = append(trendData, *data)
	}

	return trendData, nil
}

func (s *statsService) GetRunningTasks() ([]RunningTask, error) {
	var tasks []models.TestTask

	err := s.db.Model(&models.TestTask{}).
		Where("status = ?", "running").
		Order("start_time desc").
		Find(&tasks).Error
	if err != nil {
		return nil, err
	}

	var runningTasks []RunningTask
	for _, task := range tasks {
		progress := 0.0
		if task.TotalTests > 0 {
			progress = float64(task.PassedTests+task.FailedTests) / float64(task.TotalTests) * 100
		}

		estimatedEndTime := int64(0)
		if task.StartTime > 0 && progress > 0 && progress < 100 {
			elapsed := time.Now().Unix() - task.StartTime
			estimatedTotal := float64(elapsed) / progress * 100
			estimatedEndTime = task.StartTime + int64(estimatedTotal)
		}

		runningTasks = append(runningTasks, RunningTask{
			ID:               task.ID,
			TaskName:         task.TaskName,
			WorkerName:       task.WorkerName,
			PlanKey:          task.PlanKey,
			StartTime:        task.StartTime,
			Progress:         progress,
			EstimatedEndTime: estimatedEndTime,
			PassedTests:      task.PassedTests,
			FailedTests:      task.FailedTests,
			TotalTests:       task.TotalTests,
		})
	}

	return runningTasks, nil
}
