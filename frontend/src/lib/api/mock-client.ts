import createClient from "openapi-fetch";
import type { paths } from "@/lib/api/mock-openapi";
import { env } from "@/lib/env";

export const mockApiClient = createClient<paths>({
  baseUrl: env.NEXT_PUBLIC_MOCK_API_BASE_URL,
});
