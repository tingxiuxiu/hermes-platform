package services

import (
	"com.hermes.platform/internal/models"
	"com.hermes.platform/internal/repository"
	"errors"
)

type ProjectService interface {
	CreateProject(name, description string, userID uint) (*models.Project, error)
	GetProject(id uint) (*models.Project, error)
	ListProjects(page, pageSize int) ([]models.Project, int64, error)
	UpdateProject(id uint, name, description string) error
	DeleteProject(id uint) error

	CreateVersion(projectID uint, version, description string) (*models.ProjectVersion, error)
	GetVersion(id uint) (*models.ProjectVersion, error)
	ListVersions(projectID uint) ([]models.ProjectVersion, error)
	UpdateVersion(id uint, version, description, status string) error
	DeleteVersion(id uint) error

	CreateTestPlan(versionID uint, name, description string, userID uint) (*models.TestPlan, error)
	GetTestPlan(id uint) (*models.TestPlan, error)
	ListTestPlans(versionID uint, page, pageSize int) ([]models.TestPlan, int64, error)
	UpdateTestPlan(id uint, name, description, status string) error
	DeleteTestPlan(id uint) error

	CreateTestCase(planID uint, name, status, remark, preCondition, stepDescription, expectedResult string, priority int, userID uint) (*models.TestCase, error)
	GetTestCase(id uint) (*models.TestCase, error)
	GetTestCaseByCaseKey(caseKey string) (*models.TestCase, error)
	ListTestCases(planID uint, page, pageSize int) ([]models.TestCase, int64, error)
	UpdateTestCase(id uint, name, status, remark, preCondition, stepDescription, expectedResult string, priority int) error
	DeleteTestCase(id uint) error
}

type projectService struct {
	projectRepo repository.ProjectRepository
}

func NewProjectService(projectRepo repository.ProjectRepository) ProjectService {
	return &projectService{
		projectRepo: projectRepo,
	}
}

func (s *projectService) CreateProject(name, description string, userID uint) (*models.Project, error) {
	project := &models.Project{
		Name:        name,
		Description: description,
		CreatedBy:   userID,
		UpdatedBy:   userID,
	}
	err := s.projectRepo.Create(project)
	if err != nil {
		return nil, err
	}
	return project, nil
}

func (s *projectService) GetProject(id uint) (*models.Project, error) {
	return s.projectRepo.FindByID(id)
}

func (s *projectService) ListProjects(page, pageSize int) ([]models.Project, int64, error) {
	return s.projectRepo.FindAll(page, pageSize)
}

func (s *projectService) UpdateProject(id uint, name, description string) error {
	project, err := s.projectRepo.FindByID(id)
	if err != nil {
		return err
	}
	project.Name = name
	project.Description = description
	return s.projectRepo.Update(project)
}

func (s *projectService) DeleteProject(id uint) error {
	return s.projectRepo.Delete(id)
}

func (s *projectService) CreateVersion(projectID uint, version, description string) (*models.ProjectVersion, error) {
	project, err := s.projectRepo.FindByID(projectID)
	if err != nil {
		return nil, errors.New("project not found")
	}

	versionRecord := &models.ProjectVersion{
		ProjectID:   project.ID,
		Version:     version,
		Description: description,
		Status:      "active",
	}
	err = s.projectRepo.CreateVersion(versionRecord)
	if err != nil {
		return nil, err
	}

	project.Versions = append(project.Versions, *versionRecord)
	return versionRecord, nil
}

func (s *projectService) GetVersion(id uint) (*models.ProjectVersion, error) {
	return s.projectRepo.FindVersionByID(id)
}

func (s *projectService) ListVersions(projectID uint) ([]models.ProjectVersion, error) {
	return s.projectRepo.FindVersionsByProjectID(projectID)
}

func (s *projectService) UpdateVersion(id uint, version, description, status string) error {
	versionRecord, err := s.projectRepo.FindVersionByID(id)
	if err != nil {
		return err
	}
	versionRecord.Version = version
	versionRecord.Description = description
	versionRecord.Status = status
	return s.projectRepo.UpdateVersion(versionRecord)
}

func (s *projectService) DeleteVersion(id uint) error {
	return s.projectRepo.DeleteVersion(id)
}

func (s *projectService) CreateTestPlan(versionID uint, name, description string, userID uint) (*models.TestPlan, error) {
	version, err := s.projectRepo.FindVersionByID(versionID)
	if err != nil {
		return nil, errors.New("version not found")
	}

	plan := &models.TestPlan{
		VersionID:   version.ID,
		Name:        name,
		Description: description,
		Status:      "draft",
		CreatedBy:   userID,
		UpdatedBy:   userID,
	}
	err = s.projectRepo.CreateTestPlan(plan)
	if err != nil {
		return nil, err
	}

	return plan, nil
}

func (s *projectService) GetTestPlan(id uint) (*models.TestPlan, error) {
	return s.projectRepo.FindTestPlanByID(id)
}

func (s *projectService) ListTestPlans(versionID uint, page, pageSize int) ([]models.TestPlan, int64, error) {
	return s.projectRepo.FindTestPlansByVersionID(versionID, page, pageSize)
}

func (s *projectService) UpdateTestPlan(id uint, name, description, status string) error {
	plan, err := s.projectRepo.FindTestPlanByID(id)
	if err != nil {
		return err
	}
	plan.Name = name
	plan.Description = description
	plan.Status = status
	return s.projectRepo.UpdateTestPlan(plan)
}

func (s *projectService) DeleteTestPlan(id uint) error {
	return s.projectRepo.DeleteTestPlan(id)
}

func (s *projectService) CreateTestCase(planID uint, name, status, remark, preCondition, stepDescription, expectedResult string, priority int, userID uint) (*models.TestCase, error) {
	plan, err := s.projectRepo.FindTestPlanByID(planID)
	if err != nil {
		return nil, errors.New("test plan not found")
	}

	caseKey, err := s.projectRepo.GenerateCaseKey(planID)
	if err != nil {
		return nil, err
	}

	testCase := &models.TestCase{
		PlanID:           plan.ID,
		Name:             name,
		CaseKey:          caseKey,
		Status:           models.TestCaseStatus(status),
		Remark:           remark,
		PreCondition:     preCondition,
		StepDescription:  stepDescription,
		ExpectedResult:   expectedResult,
		Priority:         priority,
		CreatedBy:        userID,
		UpdatedBy:        userID,
	}

	if testCase.Status == "" {
		testCase.Status = models.TestCaseStatusDeveloping
	}

	err = s.projectRepo.CreateTestCase(testCase)
	if err != nil {
		return nil, err
	}

	return testCase, nil
}

func (s *projectService) GetTestCase(id uint) (*models.TestCase, error) {
	return s.projectRepo.FindTestCaseByID(id)
}

func (s *projectService) GetTestCaseByCaseKey(caseKey string) (*models.TestCase, error) {
	return s.projectRepo.FindTestCaseByCaseKey(caseKey)
}

func (s *projectService) ListTestCases(planID uint, page, pageSize int) ([]models.TestCase, int64, error) {
	return s.projectRepo.FindTestCasesByPlanID(planID, page, pageSize)
}

func (s *projectService) UpdateTestCase(id uint, name, status, remark, preCondition, stepDescription, expectedResult string, priority int) error {
	testCase, err := s.projectRepo.FindTestCaseByID(id)
	if err != nil {
		return err
	}
	testCase.Name = name
	testCase.Status = models.TestCaseStatus(status)
	testCase.Remark = remark
	testCase.PreCondition = preCondition
	testCase.StepDescription = stepDescription
	testCase.ExpectedResult = expectedResult
	testCase.Priority = priority
	return s.projectRepo.UpdateTestCase(testCase)
}

func (s *projectService) DeleteTestCase(id uint) error {
	return s.projectRepo.DeleteTestCase(id)
}
