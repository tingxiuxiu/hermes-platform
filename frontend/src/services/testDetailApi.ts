import { apiClient } from "../lib/api";
import type { ApiResponse } from "../lib/types";

export interface TestDetail {
  id: number;
  test_task_id: number;
  test_name: string;
  test_status: "passed" | "failed" | "skipped";
  error_message: string;
  test_start_time: number;
  test_end_time: number;
  duration: number;
  test_data: string;
  created_at: string;
  updated_at: string;
}

export interface TestStepDetail {
  id: number;
  test_detail_id: number;
  step_name: string;
  start_time: number;
  end_time: number;
  duration: number;
  passed: boolean;
  screenshot: string;
  verification_area: string;
  created_at: string;
  updated_at: string;
}

export interface TestDetailListResponse {
  details: TestDetail[];
  total: number;
  page: number;
  page_size: number;
}

export interface TestStepListResponse {
  steps: TestStepDetail[];
  total: number;
  page: number;
  page_size: number;
}

export interface TestDetailQueryParams {
  page?: number;
  page_size?: number;
  test_status?: "passed" | "failed" | "skipped";
  [key: string]: string | number | undefined;
}

export interface TestStepQueryParams {
  page?: number;
  page_size?: number;
  passed?: boolean;
}

export const testDetailApi = {
  getTestDetailsByTaskId: (
    taskId: number,
    params?: TestDetailQueryParams,
  ): Promise<ApiResponse<TestDetailListResponse>> => {
    return apiClient.get<TestDetailListResponse>(
      `/api/test-tasks/${taskId}/details`,
      params,
    );
  },

  getTestStepsByDetailId: (
    detailId: number,
    params?: TestStepQueryParams,
  ): Promise<ApiResponse<TestStepListResponse>> => {
    const queryParams: Record<string, string | number | undefined> | undefined =
      params
        ? {
            page: params.page,
            page_size: params.page_size,
            passed:
              params.passed !== undefined ? String(params.passed) : undefined,
          }
        : undefined;
    return apiClient.get<TestStepListResponse>(
      `/api/test-details/${detailId}/steps`,
      queryParams,
    );
  },
};
