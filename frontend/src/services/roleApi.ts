import { apiClient } from "../lib/api";
import type { ApiResponse } from "../lib/types";

export interface Role {
  id: number;
  name: string;
  description: string;
  permissions: string[];
  created_at: string;
  updated_at: string;
}

export interface RoleListResponse {
  roles: Role[];
  total: number;
}

export const roleApi = {
  getRoles: (): Promise<ApiResponse<RoleListResponse>> => {
    return apiClient.get<RoleListResponse>("/api/roles");
  },
};
