import createClient from "openapi-fetch";
import type { paths } from "@/shared/api/mock-openapi";
import { env } from "@/shared/lib/env";

export const mockApiClient = createClient<paths>({
  baseUrl: env.NEXT_PUBLIC_MOCK_API_BASE_URL,
});
