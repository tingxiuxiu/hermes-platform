import { apiClient } from "../lib/api";
import type { ApiResponse } from "../lib/types";

export interface TestTask {
  id: number;
  build_id: string;
  task_name: string;
  status: "pending" | "running" | "completed" | "failed";
  start_time: number;
  end_time: number;
  duration: number;
  total_tests: number;
  passed_tests: number;
  failed_tests: number;
  worker_name: string;
  plan_key: string;
  created_at: string;
  updated_at: string;
}

export interface TestTaskListResponse {
  tasks: TestTask[];
  total: number;
  page: number;
  page_size: number;
}

export interface TestTaskQueryParams {
  page?: number;
  page_size?: number;
  status?: "pending" | "running" | "completed" | "failed";
  worker_name?: string;
  plan_key?: string;
  [key: string]: string | number | undefined;
}

export const testTaskApi = {
  getTestTasks: (
    params?: TestTaskQueryParams,
  ): Promise<ApiResponse<TestTaskListResponse>> => {
    return apiClient.get<TestTaskListResponse>("/api/test-tasks", params);
  },
};
