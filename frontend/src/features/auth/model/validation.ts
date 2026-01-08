import { z } from "zod";

const email = z
  .string()
  .min(1, "Email обязателен")
  .email("Некорректный email");

const password = z.string().min(1, "Пароль обязателен");

export const loginSchema = z.object({
  email,
  password,
});

export const registerSchema = z.object({
  email,
  password,
  name: z
    .string()
    .min(1, "Введите имя")
    .optional()
    .or(z.literal("")),
});

export type LoginFormValues = z.infer<typeof loginSchema>;
export type RegisterFormValues = z.infer<typeof registerSchema>;
