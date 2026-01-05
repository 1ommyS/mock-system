import { apiClient } from "@/lib/api/client";
import type { components } from "@/lib/api/openapi";
import { normalizeApiError } from "@/lib/api/errors";

export type RegisterRequest = components["schemas"]["RegisterRequest"];
export type RegisterResponse = components["schemas"]["RegisterResponse"];
export type LoginRequest = components["schemas"]["LoginRequest"];
export type TokenResponse = components["schemas"]["TokenResponse"];
export type RefreshRequest = components["schemas"]["RefreshRequest"];
export type LogoutRequest = components["schemas"]["LogoutRequest"];
export type MeResponse = components["schemas"]["MeResponse"];

async function readErrorPayload(response?: Response): Promise<unknown> {
  if (!response) return undefined;
  const contentType = response.headers.get("content-type") ?? "";
  if (contentType.includes("application/json")) {
    try {
      return await response.json();
    } catch {
      return undefined;
    }
  }
  try {
    return await response.text();
  } catch {
    return undefined;
  }
}

export async function registerUser(
  payload: RegisterRequest
): Promise<RegisterResponse> {
  const { data, error, response } = await apiClient.POST("/auth/v1/register", {
    body: payload,
  });

  if (data) return data;

  const errorPayload = error ?? (await readErrorPayload(response));
  throw await normalizeApiError(response, errorPayload);
}

export async function loginUser(payload: LoginRequest): Promise<TokenResponse> {
  const { data, error, response } = await apiClient.POST("/auth/v1/login", {
    body: payload,
  });

  if (data) return data;

  const errorPayload = error ?? (await readErrorPayload(response));
  throw await normalizeApiError(response, errorPayload);
}

export async function refreshToken(
  payload: RefreshRequest
): Promise<TokenResponse> {
  const { data, error, response } = await apiClient.POST(
    "/auth/v1/token/refresh",
    { body: payload }
  );

  if (data) return data;

  const errorPayload = error ?? (await readErrorPayload(response));
  throw await normalizeApiError(response, errorPayload);
}

export async function logoutUser(
  payload: LogoutRequest,
  accessToken?: string | null
): Promise<void> {
  const { error, response } = await apiClient.POST("/auth/v1/logout", {
    body: payload,
    headers: accessToken ? { Authorization: `Bearer ${accessToken}` } : undefined,
  });

  if (response?.ok) return;

  const errorPayload = error ?? (await readErrorPayload(response));
  throw await normalizeApiError(response, errorPayload);
}

export async function getMe(accessToken: string): Promise<MeResponse> {
  const { data, error, response } = await apiClient.GET("/auth/v1/me", {
    headers: { Authorization: `Bearer ${accessToken}` },
  });

  if (data) return data;

  const errorPayload = error ?? (await readErrorPayload(response));
  throw await normalizeApiError(response, errorPayload);
}
