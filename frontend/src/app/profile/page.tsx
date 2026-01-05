"use client";

import { useRouter } from "next/navigation";
import { useEffect, useState } from "react";

import { logoutUser } from "@/lib/api/auth";
import {
  clearTokens,
  getAccessToken,
  getTokens,
  isAuthenticated,
} from "@/lib/auth/session";
import { authLoggedIn, authLoggedOut } from "@/lib/state/auth";
import { Button } from "@/components/ui/button";

export default function ProfilePage() {
  const router = useRouter();
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    const tokens = getTokens();
    if (tokens?.accessToken && tokens?.refreshToken) {
      authLoggedIn({
        accessToken: tokens.accessToken,
        refreshToken: tokens.refreshToken,
      });
    }

    if (!isAuthenticated()) {
      router.replace("/login?next=/profile");
      return;
    }
    setLoading(false);
  }, [router]);

  const handleLogout = async () => {
    setError(null);
    const tokens = getTokens();
    const accessToken = getAccessToken();

    if (!tokens?.refreshToken) {
      clearTokens();
      authLoggedOut();
      router.replace("/login");
      return;
    }

    try {
      await logoutUser({ refreshToken: tokens.refreshToken }, accessToken);
      clearTokens();
      authLoggedOut();
      router.replace("/login");
    } catch {
      setError("Не удалось выйти. Попробуйте еще раз.");
    }
  };

  if (loading) {
    return (
      <div className="flex min-h-screen items-center justify-center">
        <p className="text-sm text-[var(--text-muted)]">Проверяем доступ...</p>
      </div>
    );
  }

  return (
    <div className="relative min-h-screen overflow-hidden bg-[radial-gradient(circle_at_top,_var(--page-spot),_transparent_55%)]">
      <div className="pointer-events-none absolute inset-0 bg-[radial-gradient(circle_at_20%_20%,_rgba(20,184,166,0.18),_transparent_45%),_radial-gradient(circle_at_80%_10%,_rgba(250,204,21,0.2),_transparent_35%)]" />
      <main className="relative flex min-h-screen flex-col items-center justify-center gap-6 px-6 py-12 text-center">
        <div className="rounded-3xl border border-[var(--card-border)] bg-[var(--card-bg)]/90 px-10 py-12 shadow-soft backdrop-blur">
          <h1 className="text-3xl font-semibold">Вы вошли</h1>
          <p className="mt-3 text-sm text-[var(--text-muted)]">
            Сессия активна. Добро пожаловать в защищенный раздел.
          </p>
          {error ? (
            <p className="mt-4 text-sm text-rose-600 dark:text-rose-300">
              {error}
            </p>
          ) : null}
          <Button className="mt-6" variant="outline" onClick={handleLogout}>
            Выйти
          </Button>
        </div>
      </main>
    </div>
  );
}
