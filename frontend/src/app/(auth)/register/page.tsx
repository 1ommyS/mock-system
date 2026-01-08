"use client";

import Link from "next/link";
import { useRouter, useSearchParams } from "next/navigation";
import { useState } from "react";
import { zodResolver } from "@hookform/resolvers/zod";
import { ArrowRight, Eye, EyeOff, UserPlus } from "lucide-react";
import { useForm } from "react-hook-form";

import { getMe, loginUser, registerUser } from "@/features/auth/api/auth";
import type { ApiError } from "@/shared/api/errors";
import { clearTokens, storeTokens, storeUserContext } from "@/entities/user/model/session";
import { authLoggedIn } from "@/entities/user/model/auth";
import { registerSchema, type RegisterFormValues } from "@/features/auth/model/validation";
import { Alert, AlertDescription, AlertTitle } from "@/shared/ui/alert";
import { Button } from "@/shared/ui/button";
import {
  Card,
  CardContent,
  CardDescription,
  CardFooter,
  CardHeader,
  CardTitle,
} from "@/shared/ui/card";
import { Input } from "@/shared/ui/input";
import { Separator } from "@/shared/ui/separator";

function isApiError(error: unknown): error is ApiError {
  return Boolean(error && typeof error === "object" && "message" in error);
}

export default function RegisterPage() {
  const router = useRouter();
  const searchParams = useSearchParams();
  const nextParam = searchParams.get("next");
  const nextPath =
    nextParam && nextParam.startsWith("/") ? nextParam : "/mocks";
  const [serverError, setServerError] = useState<string | null>(null);
  const [showPassword, setShowPassword] = useState(false);
  const [remember, setRemember] = useState(true);

  const {
    register,
    handleSubmit,
    setError,
    formState: { errors, isSubmitting },
  } = useForm<RegisterFormValues>({
    resolver: zodResolver(registerSchema),
    defaultValues: {
      email: "",
      password: "",
      name: "",
    },
  });

  const onSubmit = handleSubmit(async (values) => {
    setServerError(null);

    const payload = {
      email: values.email,
      password: values.password,
      name: values.name?.trim() || undefined,
    };

    try {
      await registerUser(payload);
      const token = await loginUser({
        email: values.email,
        password: values.password,
      });

      if (!token.accessToken || !token.refreshToken) {
        setServerError("Ответ сервера не содержит токены.");
        return;
      }

      storeTokens(
        {
          accessToken: token.accessToken,
          refreshToken: token.refreshToken,
          expiresIn: token.expiresIn ?? null,
        },
        remember
      );

      const me = await getMe(token.accessToken);
      if (!me.userId) {
        setServerError("Не удалось получить профиль пользователя.");
        clearTokens();
        return;
      }
      storeUserContext(
        {
          userId: me.userId,
          roles: me.roles ?? [],
        },
        remember
      );
      authLoggedIn({
        accessToken: token.accessToken,
        refreshToken: token.refreshToken,
        userId: me.userId,
        roles: me.roles ?? [],
      });

      router.replace(nextPath);
    } catch (error) {
      if (isApiError(error)) {
        if (error.fieldErrors) {
          for (const [field, message] of Object.entries(error.fieldErrors)) {
            if (field === "email" || field === "password" || field === "name") {
              setError(field, { type: "server", message });
            }
          }
        }
        setServerError(error.message);
      } else {
        setServerError("Не удалось завершить регистрацию. Попробуйте позже.");
      }
    }
  });

  return (
    <Card className="w-full max-w-lg">
      <CardHeader>
        <div className="flex items-center gap-3 text-sm uppercase tracking-[0.3em] text-[var(--text-muted)]">
          <UserPlus className="h-5 w-5" />
          Регистрация
        </div>
        <CardTitle className="mt-4">Создайте аккаунт</CardTitle>
        <CardDescription>
          Заполните профиль, чтобы получить доступ к возможностям платформы.
        </CardDescription>
      </CardHeader>
      <form onSubmit={onSubmit}>
        <CardContent className="space-y-4">
          {serverError ? (
            <Alert>
              <AlertTitle>Ошибка</AlertTitle>
              <AlertDescription>{serverError}</AlertDescription>
            </Alert>
          ) : null}
          <div className="space-y-2">
            <label className="text-sm font-medium">Имя</label>
            <Input
              type="text"
              placeholder="Иван Петров"
              autoComplete="name"
              {...register("name")}
            />
            {errors.name?.message ? (
              <p className="text-xs text-rose-600 dark:text-rose-300">
                {errors.name.message}
              </p>
            ) : null}
          </div>
          <div className="space-y-2">
            <label className="text-sm font-medium">Email</label>
            <Input
              type="email"
              placeholder="name@company.com"
              autoComplete="email"
              {...register("email")}
            />
            {errors.email?.message ? (
              <p className="text-xs text-rose-600 dark:text-rose-300">
                {errors.email.message}
              </p>
            ) : null}
          </div>
          <div className="space-y-2">
            <label className="text-sm font-medium">Пароль</label>
            <div className="relative">
              <Input
                type={showPassword ? "text" : "password"}
                placeholder="Придумайте пароль"
                autoComplete="new-password"
                {...register("password")}
              />
              <button
                type="button"
                className="absolute right-3 top-1/2 -translate-y-1/2 text-[var(--text-muted)] transition hover:text-[var(--text-primary)]"
                onClick={() => setShowPassword((value) => !value)}
                aria-label={showPassword ? "Hide password" : "Show password"}
              >
                {showPassword ? (
                  <EyeOff className="h-4 w-4" />
                ) : (
                  <Eye className="h-4 w-4" />
                )}
              </button>
            </div>
            {errors.password?.message ? (
              <p className="text-xs text-rose-600 dark:text-rose-300">
                {errors.password.message}
              </p>
            ) : null}
          </div>
          <div className="flex items-center justify-between text-sm">
            <label className="flex items-center gap-2 text-[var(--text-muted)]">
              <input
                type="checkbox"
                className="h-4 w-4 rounded border-[var(--card-border)] text-brand-600 focus:ring-brand-600"
                checked={remember}
                onChange={(event) => setRemember(event.target.checked)}
              />
              Запомнить меня
            </label>
          </div>
        </CardContent>
        <CardFooter className="flex flex-col gap-4">
          <Button
            className="w-full text-base text-black dark:text-white"
            disabled={isSubmitting}
            type="submit"
          >
            {isSubmitting ? "Создание..." : "Зарегистрироваться"}
            <ArrowRight className="h-4 w-4" />
          </Button>
          <Separator />
          <p className="text-sm text-[var(--text-muted)]">
            Уже есть аккаунт?{" "}
            <Link className="text-brand-600 hover:underline" href="/login">
              Войти
            </Link>
          </p>
        </CardFooter>
      </form>
    </Card>
  );
}
