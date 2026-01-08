import createClient from "openapi-fetch";
import type { paths } from "@/shared/api/openapi";
import { env } from "@/shared/lib/env";

export const apiClient = createClient<paths>({
  baseUrl: env.NEXT_PUBLIC_API_BASE_URL,
  credentials: "include",
});
