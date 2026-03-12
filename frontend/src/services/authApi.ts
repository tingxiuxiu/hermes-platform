import { apiClient } from "../lib/api";
import type { ApiResponse } from "../lib/types";

export interface RegisterRequest {
  name: string;
  email: string;
  password: string;
}

export interface LoginRequest {
  email: string;
  password: string;
}

export interface ChangePasswordRequest {
  old_password: string;
  new_password: string;
}

export interface ForgotPasswordRequest {
  email: string;
}

export interface AuthUser {
  id: number;
  name: string;
  email: string;
}

export interface UserProfile extends AuthUser {
  roles: string[];
}

export interface RegisterResponse {
  message: string;
  user: AuthUser;
}

export interface LoginResponse {
  message: string;
  token: string;
}

export interface ProfileResponse {
  user: UserProfile;
}

export const authApi = {
  register: (data: RegisterRequest): Promise<ApiResponse<RegisterResponse>> => {
    return apiClient.post<RegisterResponse>("/api/auth/register", data);
  },

  login: (data: LoginRequest): Promise<ApiResponse<LoginResponse>> => {
    return apiClient.post<LoginResponse>("/api/auth/login", data);
  },

  getProfile: (): Promise<ApiResponse<ProfileResponse>> => {
    return apiClient.get<ProfileResponse>("/api/auth/profile");
  },

  updateProfile: (data: { name: string }): Promise<ApiResponse<{ message: string }>> => {
    return apiClient.put<{ message: string }>("/api/auth/profile", data);
  },

  changePassword: (data: ChangePasswordRequest): Promise<ApiResponse<{ message: string }>> => {
    return apiClient.post<{ message: string }>("/api/auth/change-password", data);
  },

  forgotPassword: (data: ForgotPasswordRequest): Promise<ApiResponse<{ message: string }>> => {
    return apiClient.post<{ message: string }>("/api/auth/forgot-password", data);
  },
};
