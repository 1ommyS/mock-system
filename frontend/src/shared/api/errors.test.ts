import { describe, expect, it } from "vitest";

import { normalizeApiError } from "./errors";

describe("normalizeApiError", () => {
  it("uses status-based message when payload is empty", async () => {
    const response = { status: 401 } as Response;
    const error = await normalizeApiError(response, {});

    expect(error.message).toBe("Неверный email или пароль.");
    expect(error.status).toBe(401);
  });

  it("extracts field errors from array payload", async () => {
    const response = { status: 422 } as Response;
    const payload = {
      errors: [{ field: "email", message: "Некорректный email" }],
    };

    const error = await normalizeApiError(response, payload);

    expect(error.fieldErrors).toEqual({ email: "Некорректный email" });
    expect(error.message).toBe("Проверьте корректность полей.");
  });

  it("prefers explicit message from payload", async () => {
    const response = { status: 400 } as Response;
    const payload = { message: "Custom error" };

    const error = await normalizeApiError(response, payload);

    expect(error.message).toBe("Custom error");
  });
});
