import type { ApiResponse } from "./types";

const BASE_URL = import.meta.env.VITE_API_BASE_URL || "http://localhost:8080";

const TOKEN_KEY = "auth_token";

function getToken(): string | null {
  return localStorage.getItem(TOKEN_KEY);
}

function setToken(token: string): void {
  localStorage.setItem(TOKEN_KEY, token);
}

function removeToken(): void {
  localStorage.removeItem(TOKEN_KEY);
}

interface RequestOptions {
  method: "GET" | "POST" | "PUT" | "DELETE";
  headers?: Record<string, string>;
  body?: unknown;
}

async function request<T>(
  endpoint: string,
  options: RequestOptions = { method: "GET" },
): Promise<ApiResponse<T>> {
  const { method, headers = {}, body } = options;

  const token = getToken();
  const defaultHeaders: Record<string, string> = {
    "Content-Type": "application/json",
    ...headers,
  };

  if (token) {
    defaultHeaders["Authorization"] = `Bearer ${token}`;
  }

  const config: RequestInit = {
    method,
    headers: defaultHeaders,
  };

  if (body && method !== "GET") {
    config.body = JSON.stringify(body);
  }

  const response = await fetch(`${BASE_URL}${endpoint}`, config);

  const data = await response.json();

  if (!response.ok) {
    if (response.status === 401) {
      // 只有当当前不在登录页面时才跳转到登录页面
      if (window.location.pathname !== "/login" && window.location.pathname !== "/register" && window.location.pathname !== "/forgot-password") {
        removeToken();
        window.location.href = "/login";
      }
    }
    return {
      success: false,
      data: null as T,
      message: data.message || `HTTP error! status: ${response.status}`,
      code: data.code || response.status,
      error: { message: data.error || data.message || `HTTP error! status: ${response.status}`, code: data.code },
    };
  }

  return {
    success: data.code === 200 || data.code === 201,
    data: data.data,
    message: data.message,
    code: data.code,
    error: null,
  };
}

export const apiClient = {
  get<T>(
    endpoint: string,
    params?: Record<string, string | number | undefined>,
    headers?: Record<string, string>,
  ): Promise<ApiResponse<T>> {
    let url = endpoint;
    if (params) {
      const searchParams = new URLSearchParams();
      Object.entries(params).forEach(([key, value]) => {
        if (value !== undefined && value !== "") {
          searchParams.append(key, String(value));
        }
      });
      const queryString = searchParams.toString();
      if (queryString) {
        url = `${endpoint}?${queryString}`;
      }
    }
    return request<T>(url, { method: "GET", headers });
  },

  post<T>(
    endpoint: string,
    body?: unknown,
    headers?: Record<string, string>,
  ): Promise<ApiResponse<T>> {
    return request<T>(endpoint, { method: "POST", body, headers });
  },

  put<T>(
    endpoint: string,
    body?: unknown,
    headers?: Record<string, string>,
  ): Promise<ApiResponse<T>> {
    return request<T>(endpoint, { method: "PUT", body, headers });
  },

  delete<T>(
    endpoint: string,
    headers?: Record<string, string>,
  ): Promise<ApiResponse<T>> {
    return request<T>(endpoint, { method: "DELETE", headers });
  },
};

export { getToken, setToken, removeToken };
