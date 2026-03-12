import { apiClient } from "../lib/api";
import type { ApiResponse } from "../lib/types";

export interface User {
  id: number;
  name: string;
  email: string;
  roles: string[];
  status: "active" | "inactive";
  created_at: string;
  updated_at: string;
}

export interface UserListResponse {
  users: User[];
  total: number;
  page: number;
  page_size: number;
}

export interface UserQueryParams {
  page?: number;
  page_size?: number;
  status?: "active" | "inactive";
  name?: string;
  email?: string;
  [key: string]: string | number | undefined;
}

export interface UpdateUserRequest {
  name?: string;
  email?: string;
  status?: "active" | "inactive";
}

export interface AssignRolesRequest {
  role_ids: number[];
}

export interface ResetPasswordRequest {
  new_password: string;
}

export const userApi = {
  getUsers: (
    params?: UserQueryParams,
  ): Promise<ApiResponse<UserListResponse>> => {
    return apiClient.get<UserListResponse>("/api/users", params);
  },

  updateUser: (
    id: number,
    data: UpdateUserRequest,
  ): Promise<ApiResponse<{ message: string; user: User }>> => {
    return apiClient.put<{ message: string; user: User }>(
      `/api/users/${id}`,
      data,
    );
  },

  deleteUser: (id: number): Promise<ApiResponse<{ message: string }>> => {
    return apiClient.delete<{ message: string }>(`/api/users/${id}`);
  },

  assignRoles: (
    id: number,
    roleIds: number[],
  ): Promise<ApiResponse<{ message: string }>> => {
    return apiClient.put<{ message: string }>(`/api/users/${id}/roles`, {
      role_ids: roleIds,
    });
  },

  resetPassword: (
    id: number,
    newPassword: string,
  ): Promise<ApiResponse<{ message: string }>> => {
    return apiClient.post<{ message: string }>(
      `/api/users/${id}/reset-password`,
      { new_password: newPassword },
    );
  },
};
