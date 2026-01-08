"use client";

import { useEffect, useState } from "react";
import { useParams, useRouter } from "next/navigation";

import type { ApiError } from "@/shared/api/errors";
import { getMock, updateMock, type Mock } from "@/entities/mock/api/mocks";
import { getUserContext, isAuthenticated } from "@/entities/user/model/session";
import { MockForm, type MockFormValues } from "@/features/mocks/ui/mock-form";
import { Alert, AlertDescription, AlertTitle } from "@/shared/ui/alert";
import { Card, CardContent, CardHeader, CardTitle } from "@/shared/ui/card";

function isApiError(error: unknown): error is ApiError {
  return Boolean(error && typeof error === "object" && "message" in error);
}

export default function MockDetailsPage() {
  const router = useRouter();
  const params = useParams<{ id: string }>();
  const mockId = params.id;

  const [mock, setMock] = useState<Mock | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    if (!mockId) return;
    const context = getUserContext();
    if (!isAuthenticated() || !context?.userId) {
      router.replace(`/login?next=/mocks/${mockId}`);
      return;
    }

    const load = async () => {
      setLoading(true);
      setError(null);
      try {
        const data = await getMock(mockId);
        setMock(data);
      } catch (err) {
        setError(isApiError(err) ? err.message : "Не удалось загрузить мок.");
      } finally {
        setLoading(false);
      }
    };

    void load();
  }, [mockId, router]);

  const handleSubmit = async (values: MockFormValues) => {
    if (!mockId) return;
    await updateMock(mockId, {
      name: values.name,
      description: values.description || undefined,
      enabled: values.enabled ?? true,
      requestMatch: JSON.parse(values.requestMatch),
      responseTemplate: JSON.parse(values.responseTemplate),
    });
    router.replace("/mocks");
  };

  if (loading) {
    return (
      <p className="text-sm text-[var(--text-muted)]">Загрузка...</p>
    );
  }

  if (!mockId) {
    return (
      <p className="text-sm text-[var(--text-muted)]">Некорректный ID.</p>
    );
  }

  if (error) {
    return (
      <Alert>
        <AlertTitle>Ошибка</AlertTitle>
        <AlertDescription>{error}</AlertDescription>
      </Alert>
    );
  }

  if (!mock) {
    return (
      <p className="text-sm text-[var(--text-muted)]">Мок не найден.</p>
    );
  }

  return (
    <div className="mx-auto w-full max-w-3xl">
      <Card>
        <CardHeader>
          <CardTitle>Редактирование</CardTitle>
        </CardHeader>
        <CardContent>
          <MockForm
            submitLabel="Сохранить"
            defaultValues={{
              name: mock.name ?? "",
              description: mock.description ?? "",
              enabled: mock.enabled ?? true,
              requestMatch: JSON.stringify(mock.requestMatch ?? {}, null, 2),
              responseTemplate: JSON.stringify(mock.responseTemplate ?? {}, null, 2),
            }}
            onSubmit={handleSubmit}
          />
        </CardContent>
      </Card>
    </div>
  );
}
