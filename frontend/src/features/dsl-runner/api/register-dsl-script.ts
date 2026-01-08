import { normalizeApiError } from "@/shared/api/errors";

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

export async function registerDslScriptResource(
  dslScriptId: string,
  ownerUserId: string
) {
  const response = await fetch("/api/authz/dsl-script", {
    method: "POST",
    headers: {
      "Content-Type": "application/json",
    },
    body: JSON.stringify({ dslScriptId, ownerUserId }),
  });

  if (response.ok) {
    try {
      return await response.json();
    } catch {
      return undefined;
    }
  }

  const errorPayload = await readErrorPayload(response);
  throw await normalizeApiError(response, errorPayload);
}
