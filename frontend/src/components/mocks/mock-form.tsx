"use client";

import { useState } from "react";
import { zodResolver } from "@hookform/resolvers/zod";
import { Controller, useForm } from "react-hook-form";
import { z } from "zod";

import { Alert, AlertDescription, AlertTitle } from "@/components/ui/alert";
import { Button } from "@/components/ui/button";
import { CodeEditor } from "@/components/ui/code-editor";
import { Input } from "@/components/ui/input";

const jsonSchema = z
  .string()
  .min(1, "Поле обязательно")
  .refine((value) => {
    try {
      JSON.parse(value);
      return true;
    } catch {
      return false;
    }
  }, "Нужно валидное JSON значение");

const requestMatchSchema = z
  .string()
  .min(1, "Поле обязательно")
  .superRefine((value, ctx) => {
    let parsed: unknown;
    try {
      parsed = JSON.parse(value);
    } catch {
      ctx.addIssue({
        code: z.ZodIssueCode.custom,
        message: "Нужно валидное JSON значение",
      });
      return;
    }
    if (!parsed || typeof parsed !== "object" || Array.isArray(parsed)) {
      ctx.addIssue({
        code: z.ZodIssueCode.custom,
        message: "JSON должен быть объектом",
      });
      return;
    }
    const record = parsed as Record<string, unknown>;
    if (typeof record.method !== "string" || record.method.trim() === "") {
      ctx.addIssue({
        code: z.ZodIssueCode.custom,
        message: "Укажите method в requestMatch",
      });
    }
    if (typeof record.path !== "string" || record.path.trim() === "") {
      ctx.addIssue({
        code: z.ZodIssueCode.custom,
        message: "Укажите path в requestMatch",
      });
    }
  });

const formSchema = z.object({
  name: z.string().min(1, "Введите имя"),
  description: z.string().optional(),
  enabled: z.boolean().optional(),
  requestMatch: requestMatchSchema,
  responseTemplate: jsonSchema,
});

export type MockFormValues = z.infer<typeof formSchema>;

type MockFormProps = {
  defaultValues?: Partial<MockFormValues>;
  submitLabel: string;
  onSubmit: (values: MockFormValues) => Promise<void>;
};

export function MockForm({
  defaultValues,
  submitLabel,
  onSubmit,
}: MockFormProps) {
  const [serverError, setServerError] = useState<string | null>(null);

  const {
    register,
    handleSubmit,
    control,
    formState: { errors, isSubmitting },
  } = useForm<MockFormValues>({
    resolver: zodResolver(formSchema),
    defaultValues: {
      name: "",
      description: "",
      enabled: true,
      requestMatch: JSON.stringify(
        { method: "GET", path: "/example" },
        null,
        2
      ),
      responseTemplate: "{}",
      ...defaultValues,
    },
  });

  const submitHandler = handleSubmit(async (values) => {
    setServerError(null);
    try {
      await onSubmit(values);
    } catch (error) {
      if (error && typeof error === "object" && "message" in error) {
        setServerError(String(error.message));
      } else {
        setServerError("Не удалось сохранить мок. Попробуйте еще раз.");
      }
    }
  });

  return (
    <form className="space-y-5" onSubmit={submitHandler}>
      {serverError ? (
        <Alert>
          <AlertTitle>Ошибка</AlertTitle>
          <AlertDescription>{serverError}</AlertDescription>
        </Alert>
      ) : null}
      <div className="grid gap-2">
        <label className="text-sm font-medium">Имя</label>
        <Input placeholder="Новый мок" {...register("name")} />
        {errors.name?.message ? (
          <p className="text-xs text-rose-400">{errors.name.message}</p>
        ) : null}
      </div>
      <div className="grid gap-2">
        <label className="text-sm font-medium">Описание</label>
        <Input placeholder="Опционально" {...register("description")} />
      </div>
      <div className="grid gap-2">
        <Controller
          name="requestMatch"
          control={control}
          render={({ field }) => (
            <CodeEditor
              label="Request match (JSON)"
              value={field.value}
              onChange={field.onChange}
              hint="Обязательно укажите method и path. Ключи автоматически оборачиваются в кавычки."
              error={errors.requestMatch?.message}
              autoQuoteKeys
            />
          )}
        />
      </div>
      <div className="grid gap-2">
        <Controller
          name="responseTemplate"
          control={control}
          render={({ field }) => (
            <CodeEditor
              label="Response template (JSON)"
              value={field.value}
              onChange={field.onChange}
              hint="Ключи автоматически оборачиваются в кавычки."
              error={errors.responseTemplate?.message}
              autoQuoteKeys
            />
          )}
        />
      </div>
      <label className="flex items-center gap-2 text-sm text-[var(--text-muted)]">
        <input
          type="checkbox"
          className="h-4 w-4 rounded border-[var(--card-border)] text-sky-400 focus:ring-sky-400"
          {...register("enabled")}
        />
        Мок активен
      </label>
      <Button className="w-full" disabled={isSubmitting} type="submit">
        {isSubmitting ? "Сохранение..." : submitLabel}
      </Button>
    </form>
  );
}
