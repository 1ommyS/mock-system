"use client";

import Link from "next/link";
import { useRouter, useSearchParams } from "next/navigation";
import { useState } from "react";
import { zodResolver } from "@hookform/resolvers/zod";
import { ArrowRight, Eye, EyeOff, ShieldCheck } from "lucide-react";
import { useForm } from "react-hook-form";

import { loginUser } from "@/lib/api/auth";
import type { ApiError } from "@/lib/api/errors";
import { storeTokens } from "@/lib/auth/session";
import { authLoggedIn } from "@/lib/state/auth";
import { loginSchema, type LoginFormValues } from "@/lib/validation/auth";
import { Alert, AlertDescription, AlertTitle } from "@/components/ui/alert";
import { Button } from "@/components/ui/button";
import {
  Card,
  CardContent,
  CardDescription,
  CardFooter,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";
import { Input } from "@/components/ui/input";
import { Separator } from "@/components/ui/separator";

function isApiError(error: unknown): error is ApiError {
  return Boolean(error && typeof error === "object" && "message" in error);
}

export default function LoginPage() {
  const router = useRouter();
  const searchParams = useSearchParams();
  const nextParam = searchParams.get("next");
  const nextPath =
    nextParam && nextParam.startsWith("/") ? nextParam : "/profile";
  const [serverError, setServerError] = useState<string | null>(null);
  const [showPassword, setShowPassword] = useState(false);
  const [remember, setRemember] = useState(true);

  const {
    register,
    handleSubmit,
    setError,
    formState: { errors, isSubmitting },
  } = useForm<LoginFormValues>({
    resolver: zodResolver(loginSchema),
    defaultValues: {
      email: "",
      password: "",
    },
  });

  const onSubmit = handleSubmit(async (values) => {
    setServerError(null);

    try {
      const token = await loginUser(values);

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
      authLoggedIn({
        accessToken: token.accessToken,
        refreshToken: token.refreshToken,
      });

      router.replace(nextPath);
    } catch (error) {
      if (isApiError(error)) {
        if (error.fieldErrors) {
          for (const [field, message] of Object.entries(error.fieldErrors)) {
            if (field === "email" || field === "password") {
              setError(field, { type: "server", message });
            }
          }
        }
        setServerError(error.message);
      } else {
        setServerError("Не удалось выполнить вход. Попробуйте позже.");
      }
    }
  });

  return (
    <Card className="w-full max-w-lg">
      <CardHeader>
        <div className="flex items-center gap-3 text-sm uppercase tracking-[0.3em] text-[var(--text-muted)]">
          <ShieldCheck className="h-5 w-5" />
          Авторизация
        </div>
        <CardTitle className="mt-4">С возвращением</CardTitle>
        <CardDescription>
          Введите email и пароль, чтобы продолжить работу в системе.
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
                placeholder="Введите пароль"
                autoComplete="current-password"
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
            <span className="cursor-not-allowed text-[var(--text-muted)] opacity-60">
              Забыли пароль?
            </span>
          </div>
        </CardContent>
        <CardFooter className="flex flex-col gap-4">
          <Button
            className="w-full text-base text-black dark:text-white"
            disabled={isSubmitting}
            type="submit"
          >
            {isSubmitting ? "Вход..." : "Войти в аккаунт"}
            <ArrowRight className="h-4 w-4" />
          </Button>
          <Separator />
          <p className="text-sm text-[var(--text-muted)]">
            Нет аккаунта?{" "}
            <Link className="text-brand-600 hover:underline" href="/register">
              Зарегистрироваться
            </Link>
          </p>
        </CardFooter>
      </form>
    </Card>
  );
}
