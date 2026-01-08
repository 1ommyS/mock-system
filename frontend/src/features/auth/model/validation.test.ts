import { describe, expect, it } from "vitest";

import { loginSchema, registerSchema } from "./validation";

describe("auth validation", () => {
  it("rejects invalid login payload", () => {
    const result = loginSchema.safeParse({
      email: "not-an-email",
      password: "",
    });

    expect(result.success).toBe(false);
  });

  it("allows register payload with empty name", () => {
    const result = registerSchema.safeParse({
      email: "user@example.com",
      password: "secret",
      name: "",
    });

    expect(result.success).toBe(true);
  });
});
