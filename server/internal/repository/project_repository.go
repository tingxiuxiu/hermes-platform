package repository

import (
	"com.hermes.platform/internal/models"
	"fmt"

	"gorm.io/gorm"
)

type ProjectRepository interface {
	Create(project *models.Project) error
	FindByID(id uint) (*models.Project, error)
	FindAll(page, pageSize int) ([]models.Project, int64, error)
	Update(project *models.Project) error
	Delete(id uint) error

	CreateVersion(version *models.ProjectVersion) error
	FindVersionByID(id uint) (*models.ProjectVersion, error)
	FindVersionsByProjectID(projectID uint) ([]models.ProjectVersion, error)
	UpdateVersion(version *models.ProjectVersion) error
	DeleteVersion(id uint) error

	CreateTestPlan(plan *models.TestPlan) error
	FindTestPlanByID(id uint) (*models.TestPlan, error)
	FindTestPlansByVersionID(versionID uint, page, pageSize int) ([]models.TestPlan, int64, error)
	UpdateTestPlan(plan *models.TestPlan) error
	DeleteTestPlan(id uint) error

	CreateTestCase(testCase *models.TestCase) error
	FindTestCaseByID(id uint) (*models.TestCase, error)
	FindTestCaseByCaseKey(caseKey string) (*models.TestCase, error)
	FindTestCasesByPlanID(planID uint, page, pageSize int) ([]models.TestCase, int64, error)
	UpdateTestCase(testCase *models.TestCase) error
	DeleteTestCase(id uint) error
	GenerateCaseKey(planID uint) (string, error)
}

type projectRepository struct {
	db *gorm.DB
}

func NewProjectRepository(db *gorm.DB) ProjectRepository {
	return &projectRepository{db: db}
}

func (r *projectRepository) Create(project *models.Project) error {
	return r.db.Create(project).Error
}

func (r *projectRepository) FindByID(id uint) (*models.Project, error) {
	var project models.Project
	err := r.db.Preload("Versions").First(&project, id).Error
	if err != nil {
		return nil, err
	}
	return &project, nil
}

func (r *projectRepository) FindAll(page, pageSize int) ([]models.Project, int64, error) {
	var projects []models.Project
	var total int64

	if err := r.db.Model(&models.Project{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	err := r.db.Preload("Versions").Offset(offset).Limit(pageSize).Find(&projects).Error
	if err != nil {
		return nil, 0, err
	}

	return projects, total, nil
}

func (r *projectRepository) Update(project *models.Project) error {
	return r.db.Save(project).Error
}

func (r *projectRepository) Delete(id uint) error {
	return r.db.Delete(&models.Project{}, id).Error
}

func (r *projectRepository) CreateVersion(version *models.ProjectVersion) error {
	return r.db.Create(version).Error
}

func (r *projectRepository) FindVersionByID(id uint) (*models.ProjectVersion, error) {
	var version models.ProjectVersion
	err := r.db.First(&version, id).Error
	if err != nil {
		return nil, err
	}
	return &version, nil
}

func (r *projectRepository) FindVersionsByProjectID(projectID uint) ([]models.ProjectVersion, error) {
	var versions []models.ProjectVersion
	err := r.db.Where("project_id = ?", projectID).Order("created_at DESC").Find(&versions).Error
	if err != nil {
		return nil, err
	}
	return versions, nil
}

func (r *projectRepository) UpdateVersion(version *models.ProjectVersion) error {
	return r.db.Save(version).Error
}

func (r *projectRepository) DeleteVersion(id uint) error {
	return r.db.Delete(&models.ProjectVersion{}, id).Error
}

func (r *projectRepository) CreateTestPlan(plan *models.TestPlan) error {
	return r.db.Create(plan).Error
}

func (r *projectRepository) FindTestPlanByID(id uint) (*models.TestPlan, error) {
	var plan models.TestPlan
	err := r.db.First(&plan, id).Error
	if err != nil {
		return nil, err
	}
	return &plan, nil
}

func (r *projectRepository) FindTestPlansByVersionID(versionID uint, page, pageSize int) ([]models.TestPlan, int64, error) {
	var plans []models.TestPlan
	var total int64

	query := r.db.Model(&models.TestPlan{}).Where("version_id = ?", versionID)
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	err := query.Preload("TestCases").Offset(offset).Limit(pageSize).Find(&plans).Error
	if err != nil {
		return nil, 0, err
	}

	return plans, total, nil
}

func (r *projectRepository) UpdateTestPlan(plan *models.TestPlan) error {
	return r.db.Save(plan).Error
}

func (r *projectRepository) DeleteTestPlan(id uint) error {
	return r.db.Delete(&models.TestPlan{}, id).Error
}

func (r *projectRepository) CreateTestCase(testCase *models.TestCase) error {
	return r.db.Create(testCase).Error
}

func (r *projectRepository) FindTestCaseByID(id uint) (*models.TestCase, error) {
	var testCase models.TestCase
	err := r.db.First(&testCase, id).Error
	if err != nil {
		return nil, err
	}
	return &testCase, nil
}

func (r *projectRepository) FindTestCaseByCaseKey(caseKey string) (*models.TestCase, error) {
	var testCase models.TestCase
	err := r.db.Where("case_key = ?", caseKey).First(&testCase).Error
	if err != nil {
		return nil, err
	}
	return &testCase, nil
}

func (r *projectRepository) FindTestCasesByPlanID(planID uint, page, pageSize int) ([]models.TestCase, int64, error) {
	var testCases []models.TestCase
	var total int64

	query := r.db.Model(&models.TestCase{}).Where("plan_id = ?", planID)
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	err := query.Offset(offset).Limit(pageSize).Order("created_at DESC").Find(&testCases).Error
	if err != nil {
		return nil, 0, err
	}

	return testCases, total, nil
}

func (r *projectRepository) UpdateTestCase(testCase *models.TestCase) error {
	return r.db.Save(testCase).Error
}

func (r *projectRepository) DeleteTestCase(id uint) error {
	return r.db.Delete(&models.TestCase{}, id).Error
}

func (r *projectRepository) GenerateCaseKey(planID uint) (string, error) {
	var count int64
	if err := r.db.Model(&models.TestCase{}).Where("plan_id = ?", planID).Count(&count).Error; err != nil {
		return "", err
	}
	return planIDToCaseKey(planID, int(count)+1), nil
}

func planIDToCaseKey(planID uint, num int) string {
	return fmt.Sprintf("TC-%d-%04d", planID, num)
}
