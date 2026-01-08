import { NextResponse } from "next/server";
import { z } from "zod";

import { serverEnv } from "@/shared/lib/server-env";

const requestSchema = z.object({
  dslScriptId: z.string().uuid(),
  ownerUserId: z.string().uuid(),
});

export async function POST(request: Request) {
  let payload: unknown;
  try {
    payload = await request.json();
  } catch {
    return NextResponse.json(
      { message: "Некорректный JSON в запросе." },
      { status: 400 }
    );
  }

  const parsed = requestSchema.safeParse(payload);
  if (!parsed.success) {
    return NextResponse.json(
      { message: "Неверные параметры регистрации." },
      { status: 400 }
    );
  }

  const { dslScriptId, ownerUserId } = parsed.data;
  const baseUrl = serverEnv.NEXT_PUBLIC_API_BASE_URL.replace(/\/$/, "");

  const response = await fetch(`${baseUrl}/authz/v1/resources`, {
    method: "POST",
    headers: {
      "Content-Type": "application/json",
      "X-Internal-Secret": serverEnv.USER_SVC_INTERNAL_SECRET,
    },
    body: JSON.stringify({
      resourceType: "dsl_script",
      resourceId: dslScriptId,
      ownerUserId,
    }),
  });

  let responseBody: unknown = null;
  const contentType = response.headers.get("content-type") ?? "";
  if (contentType.includes("application/json")) {
    try {
      responseBody = await response.json();
    } catch {
      responseBody = null;
    }
  } else {
    try {
      const text = await response.text();
      responseBody = text ? { message: text } : null;
    } catch {
      responseBody = null;
    }
  }

  return NextResponse.json(responseBody ?? {}, { status: response.status });
}
