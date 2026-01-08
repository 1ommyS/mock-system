"use client";

import { useEffect } from "react";
import { useRouter } from "next/navigation";

import { createMock } from "@/lib/api/mocks";
import { getUserContext, isAuthenticated } from "@/lib/auth/session";
import { MockForm, type MockFormValues } from "@/components/mocks/mock-form";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";

export default function NewMockPage() {
  const router = useRouter();

  useEffect(() => {
    const context = getUserContext();
    if (!isAuthenticated() || !context?.userId) {
      router.replace("/login?next=/mocks/new");
    }
  }, [router]);

  const handleSubmit = async (values: MockFormValues) => {
    await createMock({
      name: values.name,
      description: values.description || undefined,
      enabled: values.enabled ?? true,
      requestMatch: JSON.parse(values.requestMatch),
      responseTemplate: JSON.parse(values.responseTemplate),
    });
    router.replace("/mocks");
  };

  return (
    <div className="mx-auto w-full max-w-3xl">
      <Card>
        <CardHeader>
          <CardTitle>Новый мок</CardTitle>
        </CardHeader>
        <CardContent>
          <MockForm submitLabel="Создать" onSubmit={handleSubmit} />
        </CardContent>
      </Card>
    </div>
  );
}
