import { apiClient } from "../lib/api";
import type { ApiResponse } from "../lib/types";

export interface APIToken {
  id: number;
  name: string;
  token: string;
  expires_at: string;
  is_revoked: boolean;
  created_at: string;
}

export interface CreateTokenRequest {
  name: string;
}

export interface TokenListResponse {
  tokens: APIToken[];
}

export const tokenApi = {
  createToken: (name: string): Promise<ApiResponse<APIToken>> => {
    return apiClient.post<APIToken>("/api/tokens", { name });
  },

  getTokens: (): Promise<ApiResponse<TokenListResponse>> => {
    return apiClient.get<TokenListResponse>("/api/tokens");
  },

  revokeToken: (id: number): Promise<ApiResponse<{ message: string }>> => {
    return apiClient.post<{ message: string }>(`/api/tokens/${id}/delete`);
  },
};
