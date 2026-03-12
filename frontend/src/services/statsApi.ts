import { apiClient } from "../lib/api";
import type { ApiResponse } from "../lib/types";

export interface TimeRangeStats {
  total_cases: number;
  total_tasks: number;
  passed_cases: number;
  failed_cases: number;
  skipped_cases: number;
  pass_rate: number;
}

export interface DashboardStats {
  today: TimeRangeStats;
  week: TimeRangeStats;
  month: TimeRangeStats;
  year: TimeRangeStats;
  all_time: TimeRangeStats;
}

export interface TrendData {
  date: string;
  total_cases: number;
  passed_cases: number;
  failed_cases: number;
  pass_rate: number;
}

export interface RunningTask {
  id: number;
  task_name: string;
  worker_name: string;
  plan_key: string;
  start_time: number;
  progress: number;
  estimated_end_time: number;
  passed_tests: number;
  failed_tests: number;
  total_tests: number;
}

export interface DashboardStatsResponse {
  stats: DashboardStats;
}

export interface TrendDataResponse {
  trend: TrendData[];
}

export interface RunningTasksResponse {
  tasks: RunningTask[];
}

export type TimeRange = "today" | "week" | "month" | "year" | "all_time";

export const statsApi = {
  getDashboardStats: (): Promise<ApiResponse<DashboardStatsResponse>> => {
    return apiClient.get<DashboardStatsResponse>("/api/stats/dashboard");
  },

  getTrendData: (range: TimeRange): Promise<ApiResponse<TrendDataResponse>> => {
    return apiClient.get<TrendDataResponse>("/api/stats/trend", {
      range,
    });
  },

  getRunningTasks: (): Promise<ApiResponse<RunningTasksResponse>> => {
    return apiClient.get<RunningTasksResponse>("/api/stats/running-tasks");
  },
};
