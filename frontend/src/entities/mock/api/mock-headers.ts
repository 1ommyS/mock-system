import { getAccessToken, getUserContext } from "@/entities/user/model/session";

export function getMockHeaders() {
  const context = getUserContext();
  const accessToken = getAccessToken();
  if (!context?.userId || !accessToken) {
    return undefined;
  }

  const headers: Record<string, string> = {
    "X-User-Id": context.userId,
    Authorization: `Bearer ${accessToken}`,
  };

  if (context.roles.length > 0) {
    headers["X-Roles"] = context.roles.join(",");
  }

  return headers;
}
