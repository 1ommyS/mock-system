export type FieldErrors = Record<string, string>;

export interface ApiError {
  message: string;
  status?: number;
  fieldErrors?: FieldErrors;
}

function coerceMessage(value: unknown): string | null {
  if (!value) return null;
  if (typeof value === "string") return value;
  if (typeof value === "number") return String(value);
  if (typeof value === "object" && "message" in value) {
    const message = (value as { message?: unknown }).message;
    if (typeof message === "string") return message;
  }
  return null;
}

function extractFieldErrors(payload: unknown): FieldErrors | undefined {
  if (!payload || typeof payload !== "object") return undefined;
  const record = payload as Record<string, unknown>;
  const errors = record.errors ?? record.fieldErrors ?? record.violations;

  if (Array.isArray(errors)) {
    const mapped: FieldErrors = {};
    for (const entry of errors) {
      if (!entry || typeof entry !== "object") continue;
      const item = entry as Record<string, unknown>;
      const key = String(item.field ?? item.path ?? item.name ?? "").trim();
      const message = coerceMessage(item.message ?? item.error ?? item.reason);
      if (key && message) mapped[key] = message;
    }
    return Object.keys(mapped).length ? mapped : undefined;
  }

  if (errors && typeof errors === "object") {
    const mapped: FieldErrors = {};
    for (const [key, value] of Object.entries(errors)) {
      if (typeof value === "string") mapped[key] = value;
      else if (Array.isArray(value) && typeof value[0] === "string") {
        mapped[key] = value[0];
      }
    }
    return Object.keys(mapped).length ? mapped : undefined;
  }

  return undefined;
}

export async function normalizeApiError(
  response: Response | undefined,
  errorPayload: unknown
): Promise<ApiError> {
  const status = response?.status;
  const fieldErrors = extractFieldErrors(errorPayload);

  const message =
    coerceMessage(errorPayload) ||
    coerceMessage(
      typeof errorPayload === "object" && errorPayload
        ? (errorPayload as Record<string, unknown>).error
        : null
    ) ||
    (status === 401
      ? "Неверный email или пароль."
      : status === 400
        ? "Некорректные данные запроса."
      : status === 409
        ? "Данные уже используются."
      : status === 422
        ? "Проверьте корректность полей."
          : "Не удалось выполнить запрос. Попробуйте еще раз.");

  return { message, status, fieldErrors };
}
