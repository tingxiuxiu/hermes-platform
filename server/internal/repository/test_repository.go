package repository

import (
	"com.hermes.platform/internal/models"
	"gorm.io/gorm"
)

// TestRepository 测试数据仓库接口
type TestRepository interface {
	// 测试任务相关操作
	CreateTestTask(task *models.TestTask) error
	GetTestTaskByID(id uint) (*models.TestTask, error)
	GetTestTaskByBuildID(buildID string) (*models.TestTask, error)
	UpdateTestTask(task *models.TestTask) error
	DeleteTestTask(id uint) error
	ListTestTasks(page, pageSize int, filters map[string]interface{}, sortBy string, sortDesc bool) ([]models.TestTask, int64, error)

	// 测试详情相关操作
	CreateTestDetail(detail *models.TestDetail) error
	GetTestDetailByID(id uint) (*models.TestDetail, error)
	UpdateTestDetail(detail *models.TestDetail) error
	DeleteTestDetail(id uint) error
	ListTestDetailsByTaskID(taskID uint, page, pageSize int) ([]models.TestDetail, int64, error)

	// 测试记录相关操作
	CreateTestRecord(record *models.TestRecord) error
	GetTestRecordByID(id uint) (*models.TestRecord, error)
	UpdateTestRecord(record *models.TestRecord) error
	DeleteTestRecord(id uint) error
	ListTestRecordsByTaskID(taskID uint, page, pageSize int, recordType string) ([]models.TestRecord, int64, error)

	// 测试步骤详情相关操作
	CreateTestStepDetail(step *models.TestStepDetail) error
	GetTestStepDetailByID(id uint) (*models.TestStepDetail, error)
	UpdateTestStepDetail(step *models.TestStepDetail) error
	DeleteTestStepDetail(id uint) error
	ListTestStepDetailsByTestDetailID(testDetailID uint, page, pageSize int) ([]models.TestStepDetail, int64, error)
}

// testRepository 测试数据仓库实现
type testRepository struct {
	db *gorm.DB
}

// NewTestRepository 创建测试数据仓库实例
func NewTestRepository(db *gorm.DB) TestRepository {
	return &testRepository{db: db}
}

// 测试任务相关操作
func (r *testRepository) CreateTestTask(task *models.TestTask) error {
	return r.db.Create(task).Error
}

func (r *testRepository) GetTestTaskByID(id uint) (*models.TestTask, error) {
	var task models.TestTask
	err := r.db.Preload("TestDetails").First(&task, id).Error
	if err != nil {
		return nil, err
	}
	return &task, nil
}

func (r *testRepository) GetTestTaskByBuildID(buildID string) (*models.TestTask, error) {
	var task models.TestTask
	err := r.db.Preload("TestDetails").Where("build_id = ?", buildID).First(&task).Error
	if err != nil {
		return nil, err
	}
	return &task, nil
}

func (r *testRepository) UpdateTestTask(task *models.TestTask) error {
	return r.db.Save(task).Error
}

func (r *testRepository) DeleteTestTask(id uint) error {
	return r.db.Delete(&models.TestTask{}, id).Error
}

func (r *testRepository) ListTestTasks(page, pageSize int, filters map[string]interface{}, sortBy string, sortDesc bool) ([]models.TestTask, int64, error) {
	var tasks []models.TestTask
	var total int64

	query := r.db.Model(&models.TestTask{})

	// 应用筛选条件
	for key, value := range filters {
		query = query.Where(key, value)
	}

	// 计算总数
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 应用排序
	if sortBy != "" {
		orderDir := "asc"
		if sortDesc {
			orderDir = "desc"
		}
		query = query.Order(sortBy + " " + orderDir)
	} else {
		// 默认按创建时间倒序
		query = query.Order("created_at desc")
	}

	// 应用分页
	offset := (page - 1) * pageSize
	if err := query.Preload("TestDetails").Offset(offset).Limit(pageSize).Find(&tasks).Error; err != nil {
		return nil, 0, err
	}

	return tasks, total, nil
}

// 测试详情相关操作
func (r *testRepository) CreateTestDetail(detail *models.TestDetail) error {
	return r.db.Create(detail).Error
}

func (r *testRepository) GetTestDetailByID(id uint) (*models.TestDetail, error) {
	var detail models.TestDetail
	err := r.db.First(&detail, id).Error
	if err != nil {
		return nil, err
	}
	return &detail, nil
}

func (r *testRepository) UpdateTestDetail(detail *models.TestDetail) error {
	return r.db.Save(detail).Error
}

func (r *testRepository) DeleteTestDetail(id uint) error {
	return r.db.Delete(&models.TestDetail{}, id).Error
}

func (r *testRepository) ListTestDetailsByTaskID(taskID uint, page, pageSize int) ([]models.TestDetail, int64, error) {
	var details []models.TestDetail
	var total int64

	query := r.db.Model(&models.TestDetail{}).Where("test_task_id = ?", taskID)

	// 计算总数
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 应用分页
	offset := (page - 1) * pageSize
	if err := query.Offset(offset).Limit(pageSize).Find(&details).Error; err != nil {
		return nil, 0, err
	}

	return details, total, nil
}

// 测试记录相关操作
func (r *testRepository) CreateTestRecord(record *models.TestRecord) error {
	return r.db.Create(record).Error
}

func (r *testRepository) GetTestRecordByID(id uint) (*models.TestRecord, error) {
	var record models.TestRecord
	err := r.db.First(&record, id).Error
	if err != nil {
		return nil, err
	}
	return &record, nil
}

func (r *testRepository) UpdateTestRecord(record *models.TestRecord) error {
	return r.db.Save(record).Error
}

func (r *testRepository) DeleteTestRecord(id uint) error {
	return r.db.Delete(&models.TestRecord{}, id).Error
}

func (r *testRepository) ListTestRecordsByTaskID(taskID uint, page, pageSize int, recordType string) ([]models.TestRecord, int64, error) {
	var records []models.TestRecord
	var total int64

	query := r.db.Model(&models.TestRecord{}).Where("test_task_id = ?", taskID)

	// 应用记录类型筛选
	if recordType != "" {
		query = query.Where("record_type = ?", recordType)
	}

	// 计算总数
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 应用分页
	offset := (page - 1) * pageSize
	if err := query.Order("record_time desc").Offset(offset).Limit(pageSize).Find(&records).Error; err != nil {
		return nil, 0, err
	}

	return records, total, nil
}

// 测试步骤详情相关操作
func (r *testRepository) CreateTestStepDetail(step *models.TestStepDetail) error {
	return r.db.Create(step).Error
}

func (r *testRepository) GetTestStepDetailByID(id uint) (*models.TestStepDetail, error) {
	var step models.TestStepDetail
	err := r.db.First(&step, id).Error
	if err != nil {
		return nil, err
	}
	return &step, nil
}

func (r *testRepository) UpdateTestStepDetail(step *models.TestStepDetail) error {
	return r.db.Save(step).Error
}

func (r *testRepository) DeleteTestStepDetail(id uint) error {
	return r.db.Delete(&models.TestStepDetail{}, id).Error
}

func (r *testRepository) ListTestStepDetailsByTestDetailID(testDetailID uint, page, pageSize int) ([]models.TestStepDetail, int64, error) {
	var steps []models.TestStepDetail
	var total int64

	query := r.db.Model(&models.TestStepDetail{}).Where("test_detail_id = ?", testDetailID)

	// 计算总数
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 应用分页
	offset := (page - 1) * pageSize
	if err := query.Order("start_time asc").Offset(offset).Limit(pageSize).Find(&steps).Error; err != nil {
		return nil, 0, err
	}

	return steps, total, nil
}
