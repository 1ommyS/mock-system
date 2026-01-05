import createClient from "openapi-fetch";
import type { paths } from "@/lib/api/openapi";
import { env } from "@/lib/env";

export const apiClient = createClient<paths>({
  baseUrl: env.NEXT_PUBLIC_API_BASE_URL,
  credentials: "include",
});
