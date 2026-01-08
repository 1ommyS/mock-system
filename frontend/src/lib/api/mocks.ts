import { mockApiClient } from "@/lib/api/mock-client";
import { normalizeApiError } from "@/lib/api/errors";
import { getMockHeaders } from "@/lib/api/mock-headers";
import type { components, paths } from "@/lib/api/mock-openapi";

export type Mock = components["schemas"]["Mock"];
export type Generation = components["schemas"]["Generation"];

export type CreateMockRequest =
  paths["/mocks/v1/mocks"]["post"]["requestBody"]["content"]["application/json"];
export type UpdateMockRequest =
  paths["/mocks/v1/mocks/{id}"]["patch"]["requestBody"]["content"]["application/json"];

export type ListMocksQuery =
  paths["/mocks/v1/mocks"]["get"]["parameters"]["query"];

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

export async function listMocks(query: ListMocksQuery = {}) {
  const headers = getMockHeaders();
  if (!headers) {
    throw { message: "Отсутствует X-User-Id. Войдите заново." };
  }

  const { data, error, response } = await mockApiClient.GET(
    "/mocks/v1/mocks",
    {
      params: { query },
      headers,
    }
  );

  if (data?.items) return data.items;

  const errorPayload = error ?? (await readErrorPayload(response));
  throw await normalizeApiError(response, errorPayload);
}

export async function getMock(id: string) {
  const headers = getMockHeaders();
  if (!headers) {
    throw { message: "Отсутствует X-User-Id. Войдите заново." };
  }

  const { data, error, response } = await mockApiClient.GET(
    "/mocks/v1/mocks/{id}",
    {
      params: { path: { id } },
      headers,
    }
  );

  if (data) return data;

  const errorPayload = error ?? (await readErrorPayload(response));
  throw await normalizeApiError(response, errorPayload);
}

export async function createMock(payload: CreateMockRequest) {
  const headers = getMockHeaders();
  if (!headers) {
    throw { message: "Отсутствует X-User-Id. Войдите заново." };
  }

  const { data, error, response } = await mockApiClient.POST(
    "/mocks/v1/mocks",
    {
      body: payload,
      headers,
    }
  );

  if (data) return data;

  const errorPayload = error ?? (await readErrorPayload(response));
  throw await normalizeApiError(response, errorPayload);
}

export async function updateMock(id: string, payload: UpdateMockRequest) {
  const headers = getMockHeaders();
  if (!headers) {
    throw { message: "Отсутствует X-User-Id. Войдите заново." };
  }

  const { data, error, response } = await mockApiClient.PATCH(
    "/mocks/v1/mocks/{id}",
    {
      params: { path: { id } },
      body: payload,
      headers,
    }
  );

  if (data) return data;

  const errorPayload = error ?? (await readErrorPayload(response));
  throw await normalizeApiError(response, errorPayload);
}

export async function deleteMock(id: string) {
  const headers = getMockHeaders();
  if (!headers) {
    throw { message: "Отсутствует X-User-Id. Войдите заново." };
  }

  const { error, response } = await mockApiClient.DELETE(
    "/mocks/v1/mocks/{id}",
    {
      params: { path: { id } },
      headers,
    }
  );

  if (response?.ok) return;

  const errorPayload = error ?? (await readErrorPayload(response));
  throw await normalizeApiError(response, errorPayload);
}
