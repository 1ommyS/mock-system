import { z } from "zod";

const serverEnvSchema = z.object({
  USER_SVC_INTERNAL_SECRET: z.string().min(1),
  NEXT_PUBLIC_API_BASE_URL: z.string().url(),
});

export const serverEnv = serverEnvSchema.parse({
  USER_SVC_INTERNAL_SECRET: process.env.USER_SVC_INTERNAL_SECRET,
  NEXT_PUBLIC_API_BASE_URL: process.env.NEXT_PUBLIC_API_BASE_URL,
});
