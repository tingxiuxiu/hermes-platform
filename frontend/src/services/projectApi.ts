import { apiClient } from "../lib/api";
import type { ApiResponse } from "../lib/types";

export interface Project {
  id: number;
  name: string;
  description: string;
  created_by: number;
  updated_by: number;
  created_at: string;
  updated_at: string;
  versions?: ProjectVersion[];
}

export interface ProjectVersion {
  id: number;
  project_id: number;
  version: string;
  description: string;
  status: string;
  created_at: string;
  updated_at: string;
}

export interface TestPlan {
  id: number;
  version_id: number;
  name: string;
  description: string;
  status: string;
  created_by: number;
  updated_by: number;
  created_at: string;
  updated_at: string;
}

export interface TestCase {
  id: number;
  plan_id: number;
  name: string;
  case_key: string;
  status: string;
  remark: string;
  pre_condition: string;
  step_description: string;
  expected_result: string;
  priority: number;
  created_by: number;
  updated_by: number;
  created_at: string;
  updated_at: string;
}

export interface DictItem {
  value: string | number;
  label: string;
  label_en: string;
}

export interface PaginatedResponse<T> {
  items: T[];
  total: number;
  page: number;
  page_size: number;
}

const BASE_PROJECT_API = "/api/projects";

export const projectApi = {
  getProjects: (page = 1, pageSize = 10): Promise<ApiResponse<{ projects: Project[]; total: number; page: number; page_size: number }>> => {
    return apiClient.get<{ projects: Project[]; total: number; page: number; page_size: number }>(`${BASE_PROJECT_API}`, { page, page_size: pageSize });
  },

  getProject: (id: number): Promise<ApiResponse<{ project: Project }>> => {
    return apiClient.get<{ project: Project }>(`${BASE_PROJECT_API}/${id}`);
  },

  createProject: (data: { name: string; description: string }): Promise<ApiResponse<{ project: Project }>> => {
    return apiClient.post<{ project: Project }>(`${BASE_PROJECT_API}`, data);
  },

  updateProject: (id: number, data: { name: string; description: string }): Promise<ApiResponse<{ message: string }>> => {
    return apiClient.post<{ message: string }>(`${BASE_PROJECT_API}/${id}/update`, data);
  },

  deleteProject: (id: number): Promise<ApiResponse<{ message: string }>> => {
    return apiClient.post<{ message: string }>(`${BASE_PROJECT_API}/${id}/delete`, {});
  },

  getVersions: (projectId: number): Promise<ApiResponse<{ versions: ProjectVersion[] }>> => {
    return apiClient.get<{ versions: ProjectVersion[] }>(`${BASE_PROJECT_API}/${projectId}/versions`);
  },

  getVersion: (id: number): Promise<ApiResponse<{ version: ProjectVersion }>> => {
    return apiClient.get<{ version: ProjectVersion }>(`${BASE_PROJECT_API}/versions/${id}`);
  },

  createVersion: (projectId: number, data: { version: string; description: string }): Promise<ApiResponse<{ version: ProjectVersion }>> => {
    return apiClient.post<{ version: ProjectVersion }>(`${BASE_PROJECT_API}/${projectId}/versions`, data);
  },

  updateVersion: (id: number, data: { version: string; description: string; status: string }): Promise<ApiResponse<{ message: string }>> => {
    return apiClient.post<{ message: string }>(`${BASE_PROJECT_API}/versions/${id}/update`, data);
  },

  deleteVersion: (id: number): Promise<ApiResponse<{ message: string }>> => {
    return apiClient.post<{ message: string }>(`${BASE_PROJECT_API}/versions/${id}/delete`, {});
  },

  getTestPlans: (versionId: number, page = 1, pageSize = 10): Promise<ApiResponse<{ test_plans: TestPlan[]; total: number; page: number; page_size: number }>> => {
    return apiClient.get<{ test_plans: TestPlan[]; total: number; page: number; page_size: number }>(`${BASE_PROJECT_API}/versions/${versionId}/test-plans`, { page, page_size: pageSize });
  },

  getTestPlan: (id: number): Promise<ApiResponse<{ test_plan: TestPlan }>> => {
    return apiClient.get<{ test_plan: TestPlan }>(`${BASE_PROJECT_API}/test-plans/${id}`);
  },

  createTestPlan: (versionId: number, data: { name: string; description: string }): Promise<ApiResponse<{ test_plan: TestPlan }>> => {
    return apiClient.post<{ test_plan: TestPlan }>(`${BASE_PROJECT_API}/versions/${versionId}/test-plans`, data);
  },

  updateTestPlan: (id: number, data: { name: string; description: string; status: string }): Promise<ApiResponse<{ message: string }>> => {
    return apiClient.post<{ message: string }>(`${BASE_PROJECT_API}/test-plans/${id}/update`, data);
  },

  deleteTestPlan: (id: number): Promise<ApiResponse<{ message: string }>> => {
    return apiClient.post<{ message: string }>(`${BASE_PROJECT_API}/test-plans/${id}/delete`, {});
  },

  getTestCases: (planId: number, page = 1, pageSize = 10): Promise<ApiResponse<{ test_cases: TestCase[]; total: number; page: number; page_size: number }>> => {
    return apiClient.get<{ test_cases: TestCase[]; total: number; page: number; page_size: number }>(`${BASE_PROJECT_API}/test-plans/${planId}/test-cases`, { page, page_size: pageSize });
  },

  getTestCase: (id: number): Promise<ApiResponse<{ test_case: TestCase }>> => {
    return apiClient.get<{ test_case: TestCase }>(`${BASE_PROJECT_API}/test-cases/${id}`);
  },

  getTestCaseByKey: (caseKey: string): Promise<ApiResponse<{ test_case: TestCase }>> => {
    return apiClient.get<{ test_case: TestCase }>(`${BASE_PROJECT_API}/test-cases/key/${caseKey}`);
  },

  createTestCase: (planId: number, data: { name: string; status?: string; remark?: string; pre_condition?: string; step_description?: string; expected_result?: string; priority?: number }): Promise<ApiResponse<{ test_case: TestCase }>> => {
    return apiClient.post<{ test_case: TestCase }>(`${BASE_PROJECT_API}/test-plans/${planId}/test-cases`, data);
  },

  updateTestCase: (id: number, data: { name: string; status?: string; remark?: string; pre_condition?: string; step_description?: string; expected_result?: string; priority?: number }): Promise<ApiResponse<{ message: string }>> => {
    return apiClient.post<{ message: string }>(`${BASE_PROJECT_API}/test-cases/${id}/update`, data);
  },

  deleteTestCase: (id: number): Promise<ApiResponse<{ message: string }>> => {
    return apiClient.post<{ message: string }>(`${BASE_PROJECT_API}/test-cases/${id}/delete`, {});
  },

  getTestCaseStatuses: (): Promise<ApiResponse<{ statuses: DictItem[] }>> => {
    return apiClient.get<{ statuses: DictItem[] }>("/api/dict/test-case-statuses");
  },

  getPriorities: (): Promise<ApiResponse<{ priorities: DictItem[] }>> => {
    return apiClient.get<{ priorities: DictItem[] }>("/api/dict/priorities");
  },

  getTestPlanStatuses: (): Promise<ApiResponse<{ statuses: DictItem[] }>> => {
    return apiClient.get<{ statuses: DictItem[] }>("/api/dict/test-plan-statuses");
  },

  getVersionStatuses: (): Promise<ApiResponse<{ statuses: DictItem[] }>> => {
    return apiClient.get<{ statuses: DictItem[] }>("/api/dict/version-statuses");
  },
};
