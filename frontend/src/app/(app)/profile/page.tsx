"use client";

import { useRouter } from "next/navigation";
import { useCallback, useEffect, useMemo, useState } from "react";

import { getMe, logoutUser } from "@/features/auth/api/auth";
import type { ApiError } from "@/shared/api/errors";
import {
  clearTokens,
  getAccessToken,
  getTokens,
  getUserContext,
  hasRememberedTokens,
  isAuthenticated,
  storeUserContext,
} from "@/entities/user/model/session";
import { authLoggedIn, authLoggedOut } from "@/entities/user/model/auth";
import { Alert, AlertDescription, AlertTitle } from "@/shared/ui/alert";
import { Button } from "@/shared/ui/button";
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@/shared/ui/card";
import { Separator } from "@/shared/ui/separator";

type ProfileInfo = {
  userId: string;
  roles: string[];
  email?: string;
  status?: string;
};

const STATUS_LABELS: Record<
  string,
  {
    label: string;
    className: string;
  }
> = {
  ACTIVE: { label: "Активен", className: "text-emerald-500 dark:text-emerald-300" },
  BLOCKED: {
    label: "Заблокирован",
    className: "text-rose-600 dark:text-rose-300",
  },
  DELETED: { label: "Удален", className: "text-rose-600 dark:text-rose-300" },
};

function isApiError(error: unknown): error is ApiError {
  return Boolean(error && typeof error === "object" && "message" in error);
}

function InfoRow({
  label,
  children,
}: {
  label: string;
  children: React.ReactNode;
}) {
  return (
    <div className="grid gap-1">
      <span className="text-xs uppercase tracking-wide text-[var(--text-muted)]">
        {label}
      </span>
      <div className="text-sm text-[var(--text-primary)]">{children}</div>
    </div>
  );
}

export default function ProfilePage() {
  const router = useRouter();
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [profile, setProfile] = useState<ProfileInfo | null>(null);
  const [copied, setCopied] = useState<"idle" | "done" | "failed">("idle");

  const sessionLabel = useMemo(
    () =>
      hasRememberedTokens()
        ? "Запомнена (localStorage)"
        : "Сессия браузера (sessionStorage)",
    []
  );

  const tokens = useMemo(() => getTokens(), []);
  const expiresLabel =
    tokens?.expiresIn != null ? `${tokens.expiresIn} сек.` : "—";

  const loadProfile = useCallback(async () => {
    setLoading(true);
    setError(null);

    const currentTokens = getTokens();
    if (!currentTokens?.accessToken || !currentTokens.refreshToken) {
      router.replace("/login?next=/profile");
      return;
    }

    const context = getUserContext();

    if (context?.userId) {
      setProfile({
        userId: context.userId,
        roles: context.roles,
      });
    }

    try {
      const me = await getMe(currentTokens.accessToken);
      if (!me.userId) {
        throw new Error("missing user id");
      }
      const roles = me.roles ?? [];
      storeUserContext(
        {
          userId: me.userId,
          roles,
        },
        hasRememberedTokens()
      );
      authLoggedIn({
        accessToken: currentTokens.accessToken,
        refreshToken: currentTokens.refreshToken,
        userId: me.userId,
        roles,
      });
      setProfile({
        userId: me.userId,
        roles,
        email: me.email ?? undefined,
        status: me.status ?? undefined,
      });
    } catch (err) {
      if (isApiError(err)) {
        if (err.status === 401 || err.status === 403) {
          clearTokens();
          authLoggedOut();
          router.replace("/login?next=/profile");
          return;
        }
        setError(err.message);
        setLoading(false);
        return;
      }
      setError("Не удалось загрузить профиль. Попробуйте еще раз.");
      setLoading(false);
      return;
    }

    if (!isAuthenticated()) {
      router.replace("/login?next=/profile");
      return;
    }
    setLoading(false);
  }, [router]);

  useEffect(() => {
    void loadProfile();
  }, [loadProfile]);

  const handleLogout = async () => {
    setError(null);
    const currentTokens = getTokens();
    const accessToken = getAccessToken();

    if (!currentTokens?.refreshToken) {
      clearTokens();
      authLoggedOut();
      router.replace("/login");
      return;
    }

    try {
      await logoutUser({ refreshToken: currentTokens.refreshToken }, accessToken);
      clearTokens();
      authLoggedOut();
      router.replace("/login");
    } catch {
      setError("Не удалось выйти. Попробуйте еще раз.");
    }
  };

  const handleCopy = async () => {
    if (!profile?.userId) return;
    try {
      await navigator.clipboard.writeText(profile.userId);
      setCopied("done");
    } catch {
      setCopied("failed");
    }
    window.setTimeout(() => setCopied("idle"), 1500);
  };

  if (loading) {
    return (
      <div className="flex min-h-[60vh] items-center justify-center">
        <p className="text-sm text-[var(--text-muted)]">Проверяем доступ...</p>
      </div>
    );
  }

  const statusKey = profile?.status?.toUpperCase();
  const statusMeta = statusKey ? STATUS_LABELS[statusKey] : undefined;
  const statusLabel = statusMeta?.label ?? profile?.status ?? "—";
  const statusClassName = statusMeta?.className ?? "text-[var(--text-muted)]";

  return (
    <div className="mx-auto w-full max-w-5xl space-y-6">
      <div className="flex flex-wrap items-center justify-between gap-4">
        <div>
          <h1 className="text-3xl font-semibold">Профиль</h1>
          <p className="mt-1 text-sm text-[var(--text-muted)]">
            Вы вошли в систему. Управление учетной записью и активной сессией.
          </p>
        </div>
        <div className="flex flex-wrap gap-2">
          <Button variant="ghost" onClick={loadProfile}>
            Обновить
          </Button>
          <Button variant="outline" onClick={handleLogout}>
            Выйти
          </Button>
        </div>
      </div>

      {error ? (
        <Alert>
          <AlertTitle>Ошибка</AlertTitle>
          <AlertDescription>{error}</AlertDescription>
        </Alert>
      ) : null}

      <div className="grid gap-6 lg:grid-cols-[minmax(0,1fr)_minmax(0,0.8fr)]">
        <Card>
          <CardHeader>
            <CardTitle>Аккаунт</CardTitle>
            <CardDescription>
              Основные параметры пользователя и доступы.
            </CardDescription>
          </CardHeader>
          <CardContent className="space-y-5">
            <InfoRow label="Email">
              <span className="text-sm text-[var(--text-muted)]">
                {profile?.email ?? "—"}
              </span>
            </InfoRow>
            <InfoRow label="User ID">
              <div className="flex flex-wrap items-center gap-3">
                <span className="rounded-full border border-white/15 bg-white/10 px-3 py-1 font-mono text-xs text-[var(--text-primary)]">
                  {profile?.userId ?? "—"}
                </span>
                <Button size="sm" variant="ghost" onClick={handleCopy}>
                  Скопировать
                </Button>
                {copied === "done" ? (
                  <span className="text-xs text-emerald-300">Скопировано</span>
                ) : null}
                {copied === "failed" ? (
                  <span className="text-xs text-rose-300">
                    Не удалось скопировать
                  </span>
                ) : null}
              </div>
            </InfoRow>
            <InfoRow label="Статус">
              <span className={`text-sm ${statusClassName}`}>{statusLabel}</span>
            </InfoRow>
            <InfoRow label="Роли">
              <div className="flex flex-wrap gap-2">
                {profile?.roles?.length ? (
                  profile.roles.map((role) => (
                    <span
                      key={role}
                      className="rounded-full border border-white/15 bg-white/10 px-3 py-1 text-xs text-[var(--text-primary)]"
                    >
                      {role}
                    </span>
                  ))
                ) : (
                  <span className="text-sm text-[var(--text-muted)]">Нет</span>
                )}
              </div>
            </InfoRow>
            <Separator />
            <InfoRow label="Тип сессии">{sessionLabel}</InfoRow>
            <InfoRow label="TTL access token">{expiresLabel}</InfoRow>
          </CardContent>
        </Card>

        <Card>
          <CardHeader>
            <CardTitle>Быстрые действия</CardTitle>
            <CardDescription>Перейти к рабочим разделам.</CardDescription>
          </CardHeader>
          <CardContent className="space-y-3">
            <Button className="w-full" onClick={() => router.push("/mocks")}>
              Перейти к мокам
            </Button>
            <Button
              className="w-full"
              variant="outline"
              onClick={() => router.push("/dsl-runner")}
            >
              Открыть DSL Runner
            </Button>
          </CardContent>
        </Card>
      </div>
    </div>
  );
}
