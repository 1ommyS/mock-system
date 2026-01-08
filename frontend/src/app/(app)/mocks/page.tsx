"use client";

import { useEffect, useState } from "react";
import { useRouter } from "next/navigation";

import type { ApiError } from "@/lib/api/errors";
import { deleteMock, listMocks, type Mock } from "@/lib/api/mocks";
import { getUserContext, isAuthenticated } from "@/lib/auth/session";
import { Alert, AlertDescription, AlertTitle } from "@/components/ui/alert";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Input } from "@/components/ui/input";

function isApiError(error: unknown): error is ApiError {
  return Boolean(error && typeof error === "object" && "message" in error);
}

export default function MocksPage() {
  const router = useRouter();
  const [mocks, setMocks] = useState<Mock[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [query, setQuery] = useState("");

  const loadMocks = async (q?: string) => {
    setLoading(true);
    setError(null);
    try {
      const items = await listMocks({
        q: q || undefined,
        sort: "updatedAtDesc",
        limit: 50,
      });
      setMocks(items ?? []);
    } catch (err) {
      setError(isApiError(err) ? err.message : "Не удалось загрузить моки.");
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    const context = getUserContext();
    if (!isAuthenticated() || !context?.userId) {
      router.replace("/login?next=/mocks");
      return;
    }
    void loadMocks();
  }, [router]);

  const handleDelete = async (id: string) => {
    if (!confirm("Удалить мок?")) return;
    setError(null);
    try {
      await deleteMock(id);
      await loadMocks(query);
    } catch (err) {
      setError(isApiError(err) ? err.message : "Не удалось удалить мок.");
    }
  };

  return (
    <div className="mx-auto w-full max-w-5xl space-y-6">
      <div className="flex flex-wrap items-center justify-between gap-4">
        <div>
          <h1 className="text-3xl font-semibold">Моки</h1>
          <p className="mt-1 text-sm text-[var(--text-muted)]">
            Управляйте шаблонами ответов и правилами матчей.
          </p>
        </div>
        <Button onClick={() => router.push("/mocks/new")}>Создать мок</Button>
      </div>

      <Card>
        <CardHeader>
          <CardTitle>Поиск и фильтр</CardTitle>
        </CardHeader>
        <CardContent className="space-y-3">
          <div className="flex flex-col gap-3 sm:flex-row">
            <Input
              placeholder="Поиск по имени"
              value={query}
              onChange={(event) => setQuery(event.target.value)}
            />
            <Button
              className="sm:w-[180px]"
              onClick={() => loadMocks(query)}
            >
              Найти
            </Button>
          </div>
        </CardContent>
      </Card>

      {error ? (
        <Alert>
          <AlertTitle>Ошибка</AlertTitle>
          <AlertDescription>{error}</AlertDescription>
        </Alert>
      ) : null}

      <div className="grid gap-4">
        {loading ? (
          <p className="text-sm text-[var(--text-muted)]">Загрузка...</p>
        ) : mocks.length === 0 ? (
          <p className="text-sm text-[var(--text-muted)]">
            Моки не найдены. Создайте первый мок.
          </p>
        ) : (
          mocks.map((mock) => (
            <Card key={mock.id ?? mock.name ?? "mock"}>
              <CardContent className="flex flex-col gap-4 sm:flex-row sm:items-center sm:justify-between">
                <div>
                  <div className="flex items-center gap-3">
                    <h2 className="text-lg font-semibold">{mock.name}</h2>
                    <span className="rounded-full border border-white/20 px-3 py-1 text-xs text-[var(--text-muted)]">
                      {mock.enabled ? "Активен" : "Выключен"}
                    </span>
                  </div>
                  <p className="mt-2 text-sm text-[var(--text-muted)]">
                    {mock.description || "Описание не задано"}
                  </p>
                  <p className="mt-1 text-xs text-[var(--text-muted)]">
                    Обновлен: {mock.updatedAt ? new Date(mock.updatedAt).toLocaleString() : "—"}
                  </p>
                </div>
                <div className="flex flex-wrap gap-2">
                  <Button
                    variant="ghost"
                    disabled={!mock.id}
                    onClick={() => mock.id && router.push(`/mocks/${mock.id}`)}
                  >
                    Открыть
                  </Button>
                  <Button
                    variant="outline"
                    disabled={!mock.id}
                    onClick={() => mock.id && handleDelete(mock.id)}
                  >
                    Удалить
                  </Button>
                </div>
              </CardContent>
            </Card>
          ))
        )}
      </div>
    </div>
  );
}
