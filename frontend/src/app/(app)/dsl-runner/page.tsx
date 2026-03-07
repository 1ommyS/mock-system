"use client";

import { useEffect, useMemo, useRef, useState } from "react";
import { useRouter } from "next/navigation";
import { zodResolver } from "@hookform/resolvers/zod";
import { Controller, useForm } from "react-hook-form";
import { z } from "zod";

import type { ApiError } from "@/shared/api/errors";
import { registerDslScriptResource } from "@/features/dsl-runner/api/register-dsl-script";
import {
  createGeneration,
  getGeneration,
  listMocks,
  type CreateGenerationAcceptedResponse,
  type CreateGenerationPreviewResponse,
  type CreateGenerationRequest,
  type Generation,
  type GenerationPlanItem,
  type Mock,
} from "@/entities/mock/api/mocks";
import { getUserContext, isAuthenticated } from "@/entities/user/model/session";
import { Alert, AlertDescription, AlertTitle } from "@/shared/ui/alert";
import { Button } from "@/shared/ui/button";
import { Card, CardContent, CardHeader, CardTitle } from "@/shared/ui/card";
import { CodeEditor } from "@/shared/ui/code-editor";
import { Input } from "@/shared/ui/input";

const DSL_EXAMPLE = `version 1

let basePath = input.base.request_match.path

repeat 2 {
  emit "gen-" + loop.index {
    requestMatch {
      method: input.base.request_match.method
      path: basePath + "/v" + loop.index
    }
    responseTemplate {
      status: 200
      body: {
        stableId: randUUID()
        n: randInt(1, 10)
      }
    }
    meta {
      from: "dsl"
      i: loop.index
    }
  }
}`;

const dslScriptSchema = z
  .string()
  .min(1, "Введите DSL скрипт")
  .superRefine((value, ctx) => {
    const source = value.trim();
    if (!source) return;

    if (!source.startsWith("{")) {
      if (!/\bversion\s+1\b/.test(source)) {
        ctx.addIssue({
          code: z.ZodIssueCode.custom,
          message: "Для текстового DSL укажите заголовок `version 1`",
        });
      }
      if (!/\b(let|if|repeat|for|emit|assert)\b/.test(source)) {
        ctx.addIssue({
          code: z.ZodIssueCode.custom,
          message: "Добавьте хотя бы одну DSL-операцию: let/if/repeat/for/emit/assert",
        });
      }
      return;
    }

    let parsed: unknown;
    try {
      parsed = JSON.parse(source);
    } catch {
      ctx.addIssue({
        code: z.ZodIssueCode.custom,
        message: "JSON DSL: нужно валидное JSON значение",
      });
      return;
    }
    if (!parsed || typeof parsed !== "object" || Array.isArray(parsed)) {
      ctx.addIssue({
        code: z.ZodIssueCode.custom,
        message: "JSON DSL должен быть JSON-объектом",
      });
      return;
    }
    const record = parsed as Record<string, unknown>;
    if (record.version !== 1) {
      ctx.addIssue({
        code: z.ZodIssueCode.custom,
        message: "JSON DSL: поддерживается только версия 1",
      });
    }
    if (!Array.isArray(record.steps) || record.steps.length === 0) {
      ctx.addIssue({
        code: z.ZodIssueCode.custom,
        message: "JSON DSL: добавьте хотя бы один шаг в steps",
      });
    }
  });

const paramsSchema = z
  .string()
  .optional()
  .superRefine((value, ctx) => {
    if (!value || value.trim() === "") return;
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
        message: "params должен быть JSON-объектом",
      });
    }
  });

const formSchema = z.object({
  baseMockId: z.string().uuid("Выберите базовый мок"),
  dslScriptId: z.string().uuid("Нужен валидный UUID"),
  dslScript: dslScriptSchema,
  params: paramsSchema,
});

type DslFormValues = z.infer<typeof formSchema>;

const DSL_OPERATIONS = ["let", "if", "repeat", "for", "emit", "assert"];
const DSL_FUNCTIONS = [
  "get",
  "concat",
  "join",
  "lower",
  "upper",
  "len",
  "toString",
  "toInt",
  "sha256",
  "randInt",
  "randBool",
  "randChoice",
  "randString",
  "randUUID",
];

function isApiError(error: unknown): error is ApiError {
  return Boolean(error && typeof error === "object" && "message" in error);
}

function generateUuid() {
  if (typeof crypto !== "undefined" && "randomUUID" in crypto) {
    return crypto.randomUUID();
  }
  const bytes = new Uint8Array(16);
  if (typeof crypto !== "undefined" && crypto.getRandomValues) {
    crypto.getRandomValues(bytes);
  } else {
    for (let i = 0; i < bytes.length; i += 1) {
      bytes[i] = Math.floor(Math.random() * 256);
    }
  }
  bytes[6] = (bytes[6] & 0x0f) | 0x40;
  bytes[8] = (bytes[8] & 0x3f) | 0x80;
  const hex = Array.from(bytes, (byte) =>
    byte.toString(16).padStart(2, "0")
  );
  return `${hex.slice(0, 4).join("")}-${hex.slice(4, 6).join("")}-${hex
    .slice(6, 8)
    .join("")}-${hex.slice(8, 10).join("")}-${hex.slice(10, 16).join("")}`;
}

function parseParams(value?: string) {
  if (!value || value.trim() === "") return undefined;
  return JSON.parse(value) as Record<string, unknown>;
}

function describeRequestMatch(match: unknown) {
  if (!match || typeof match !== "object") return null;
  const record = match as Record<string, unknown>;
  const method = typeof record.method === "string" ? record.method : null;
  const path = typeof record.path === "string" ? record.path : null;
  if (!method && !path) return null;
  return `${method ?? "ANY"} ${path ?? ""}`.trim();
}

export default function DslRunnerPage() {
  const router = useRouter();
  const [mocks, setMocks] = useState<Mock[]>([]);
  const [loadingMocks, setLoadingMocks] = useState(true);
  const [mockQuery, setMockQuery] = useState("");
  const [serverError, setServerError] = useState<string | null>(null);
  const [preview, setPreview] = useState<GenerationPlanItem[] | null>(null);
  const [generation, setGeneration] = useState<Generation | null>(null);
  const [pollingId, setPollingId] = useState<string | null>(null);
  const pollStartRef = useRef<number | null>(null);

  const {
    register,
    control,
    handleSubmit,
    setValue,
    getValues,
    watch,
    formState: { errors, isSubmitting },
  } = useForm<DslFormValues>({
    resolver: zodResolver(formSchema),
    defaultValues: {
      baseMockId: "",
      dslScriptId: "",
      dslScript: DSL_EXAMPLE,
      params: "{}",
    },
  });

  const selectedMockId = watch("baseMockId");

  const availableMocks = useMemo(
    () =>
      mocks.filter(
        (mock): mock is Mock & { id: string } =>
          Boolean(mock.id) &&
          (!mockQuery ||
            mock.name?.toLowerCase().includes(mockQuery.toLowerCase()) ||
            JSON.stringify(mock.requestMatch ?? {})
              .toLowerCase()
              .includes(mockQuery.toLowerCase()))
      ),
    [mocks, mockQuery]
  );

  const selectedMock = useMemo(
    () => availableMocks.find((mock) => mock.id === selectedMockId) ?? null,
    [availableMocks, selectedMockId]
  );

  useEffect(() => {
    const context = getUserContext();
    if (!isAuthenticated() || !context?.userId) {
      router.replace("/login?next=/dsl-runner");
      return;
    }

    const loadMocks = async () => {
      setLoadingMocks(true);
      try {
        const items = await listMocks({ sort: "updatedAtDesc", limit: 100 });
        setMocks(items ?? []);
      } catch (error) {
        setServerError(
          isApiError(error) ? error.message : "Не удалось загрузить моки."
        );
      } finally {
        setLoadingMocks(false);
      }
    };

    void loadMocks();
  }, [router]);

  useEffect(() => {
    if (!getValues("dslScriptId")) {
      setValue("dslScriptId", generateUuid(), { shouldValidate: true });
    }
  }, [getValues, setValue]);

  useEffect(() => {
    if (!selectedMockId && availableMocks.length > 0) {
      setValue("baseMockId", availableMocks[0].id, {
        shouldValidate: true,
      });
    }
  }, [availableMocks, selectedMockId, setValue]);

  useEffect(() => {
    if (!pollingId) return undefined;
    pollStartRef.current = Date.now();
    let cancelled = false;

    const poll = async () => {
      if (cancelled) return;
      try {
        const updated = await getGeneration(pollingId);
        setGeneration(updated);
        if (updated.status === "DONE" || updated.status === "FAILED") {
          setPollingId(null);
        }
      } catch (error) {
        setServerError(
          isApiError(error)
            ? error.message
            : "Не удалось получить статус генерации."
        );
        setPollingId(null);
      }

      if (
        pollStartRef.current &&
        Date.now() - pollStartRef.current > 60_000
      ) {
        setPollingId(null);
      }
    };

    const timer = window.setInterval(poll, 1500);
    void poll();

    return () => {
      cancelled = true;
      window.clearInterval(timer);
    };
  }, [pollingId]);

  const runGeneration = (mode: CreateGenerationRequest["mode"]) => {
    void handleSubmit(async (values) => {
      setServerError(null);
      setPreview(null);
      setGeneration(null);
      setPollingId(null);

      const context = getUserContext();
      if (!context?.userId) {
        router.replace("/login?next=/dsl-runner");
        return;
      }

      try {
        await registerDslScriptResource(values.dslScriptId, context.userId);

        const params = parseParams(values.params);
        const payload: CreateGenerationRequest = {
          baseMockId: values.baseMockId,
          dslScriptId: values.dslScriptId,
          dslScript: values.dslScript,
          mode,
          ...(params ? { params } : {}),
        };

        const { data, response } = await createGeneration(payload);
        if (response?.status === 200) {
          const previewData = data as CreateGenerationPreviewResponse;
          setPreview(previewData.planned ?? []);
          return;
        }
        if (response?.status === 202) {
          const accepted = data as CreateGenerationAcceptedResponse;
          if (!accepted.generationId) {
            setServerError("Сервис не вернул generationId.");
            return;
          }
          setGeneration({
            generationId: accepted.generationId,
            status: accepted.status as Generation["status"],
            baseMockId: values.baseMockId,
            dslScriptId: values.dslScriptId,
          });
          setPollingId(accepted.generationId);
        }
      } catch (error) {
        setServerError(
          isApiError(error)
            ? error.message
            : "Не удалось запустить генерацию."
        );
      }
    })();
  };

  return (
    <div className="mx-auto w-full max-w-6xl space-y-6">
      <div>
        <h1 className="text-3xl font-semibold">DSL Runner</h1>
        <p className="mt-1 text-sm text-[var(--text-muted)]">
          Создавайте новые моки из базового шаблона с помощью DSL-сценариев.
        </p>
      </div>

      {serverError ? (
        <Alert>
          <AlertTitle>Ошибка</AlertTitle>
          <AlertDescription>{serverError}</AlertDescription>
        </Alert>
      ) : null}

      <div className="grid gap-6 lg:grid-cols-[minmax(0,1.15fr)_minmax(0,0.85fr)]">
        <div className="space-y-6">
          <Card>
            <CardHeader>
              <CardTitle>Настройки запуска</CardTitle>
            </CardHeader>
            <CardContent className="space-y-4">
              <div className="grid gap-2">
                <label className="text-sm font-medium">Базовый мок</label>
                <Input
                  placeholder="Поиск по мокам"
                  value={mockQuery}
                  onChange={(event) => setMockQuery(event.target.value)}
                />
                <select
                  className="flex h-11 w-full rounded-2xl border border-white/40 bg-white/55 px-4 text-sm text-[var(--text-primary)] shadow-sm backdrop-blur-md outline-none transition focus-visible:border-sky-400 focus-visible:ring-2 focus-visible:ring-[var(--ring)] dark:border-white/10 dark:bg-white/5"
                  {...register("baseMockId")}
                >
                  <option value="" disabled>
                    {loadingMocks ? "Загрузка..." : "Выберите мок"}
                  </option>
                  {availableMocks.map((mock) => (
                    <option key={mock.id} value={mock.id}>
                      {mock.name ?? mock.id}
                    </option>
                  ))}
                </select>
                {errors.baseMockId?.message ? (
                  <p className="text-xs text-rose-400">
                    {errors.baseMockId.message}
                  </p>
                ) : null}
                {selectedMock ? (
                  <p className="text-xs text-[var(--text-muted)]">
                    {describeRequestMatch(selectedMock.requestMatch) ??
                      "Нет данных о requestMatch"}
                  </p>
                ) : null}
              </div>

              <div className="grid gap-2">
                <label className="text-sm font-medium">DSL Script ID</label>
                <div className="flex flex-col gap-2 sm:flex-row">
                  <Input {...register("dslScriptId")} />
                  <Button
                    type="button"
                    variant="ghost"
                    className="sm:w-[180px]"
                    onClick={() =>
                      setValue("dslScriptId", generateUuid(), {
                        shouldValidate: true,
                      })
                    }
                  >
                    Новый ID
                  </Button>
                </div>
                {errors.dslScriptId?.message ? (
                  <p className="text-xs text-rose-400">
                    {errors.dslScriptId.message}
                  </p>
                ) : null}
              </div>
            </CardContent>
          </Card>

          <Card>
            <CardHeader className="flex flex-row items-center justify-between gap-3">
              <CardTitle>DSL скрипт</CardTitle>
              <Button
                type="button"
                variant="ghost"
                size="sm"
                onClick={() =>
                  setValue("dslScript", DSL_EXAMPLE, { shouldValidate: true })
                }
              >
                Вставить пример
              </Button>
            </CardHeader>
            <CardContent>
              <Controller
                name="dslScript"
                control={control}
                render={({ field }) => (
                  <CodeEditor
                    label="DSL Script"
                    value={field.value}
                    onChange={field.onChange}
                    error={errors.dslScript?.message}
                    hint="Поддерживается текстовый DSL (version 1) и legacy JSON DSL."
                    height="360px"
                    autoQuoteKeys={false}
                  />
                )}
              />
            </CardContent>
          </Card>

          <Card>
            <CardHeader>
              <CardTitle>Параметры запуска</CardTitle>
            </CardHeader>
            <CardContent>
              <Controller
                name="params"
                control={control}
                render={({ field }) => (
                  <CodeEditor
                    label="params (JSON)"
                    value={field.value ?? ""}
                    onChange={field.onChange}
                    error={errors.params?.message}
                    hint="Опциональные параметры для DSL (object)."
                    height="200px"
                    autoQuoteKeys
                  />
                )}
              />
            </CardContent>
          </Card>

          <div className="flex flex-wrap gap-3">
            <Button
              type="button"
              disabled={isSubmitting}
              onClick={() => runGeneration("preview")}
            >
              {isSubmitting ? "Запуск..." : "Предпросмотр"}
            </Button>
            <Button
              type="button"
              variant="outline"
              disabled={isSubmitting}
              onClick={() => runGeneration("apply")}
            >
              Применить
            </Button>
          </div>
        </div>

        <div className="space-y-6">
          <Card>
            <CardHeader>
              <CardTitle>Подсказки DSL</CardTitle>
            </CardHeader>
            <CardContent className="space-y-3 text-sm text-[var(--text-muted)]">
              <div>
                <div className="text-xs uppercase tracking-wide text-[var(--text-primary)]">
                  Ops
                </div>
                <div className="mt-2 flex flex-wrap gap-2">
                  {DSL_OPERATIONS.map((op) => (
                    <span
                      key={op}
                      className="rounded-full border border-white/15 px-3 py-1 text-xs text-[var(--text-primary)]"
                    >
                      {op}
                    </span>
                  ))}
                </div>
              </div>
              <div>
                <div className="text-xs uppercase tracking-wide text-[var(--text-primary)]">
                  $fn
                </div>
                <div className="mt-2 flex flex-wrap gap-2">
                  {DSL_FUNCTIONS.map((fn) => (
                    <span
                      key={fn}
                      className="rounded-full border border-white/15 px-3 py-1 text-xs text-[var(--text-primary)]"
                    >
                      {fn}
                    </span>
                  ))}
                </div>
              </div>
              <p className="text-xs text-[var(--text-muted)]">
                Путь к данным: <span className="text-[var(--text-primary)]">input.base.request_match.path</span>
              </p>
            </CardContent>
          </Card>

          <Card>
            <CardHeader>
              <CardTitle>Результат</CardTitle>
            </CardHeader>
            <CardContent className="space-y-4">
              {preview ? (
                <div className="space-y-3">
                  <div className="text-sm text-[var(--text-muted)]">
                    Предпросмотр: {preview.length} элементов
                  </div>
                  <div className="space-y-3">
                    {preview.map((item, index) => (
                      <div
                        key={`${item.name ?? "item"}-${index}`}
                        className="rounded-2xl border border-white/10 bg-white/5 px-4 py-3"
                      >
                        <div className="flex items-center justify-between">
                          <div className="text-sm font-medium">
                            {item.name ?? `Mock ${index + 1}`}
                          </div>
                          {item.enabled != null ? (
                            <span className="text-xs text-[var(--text-muted)]">
                              {item.enabled ? "Активен" : "Выключен"}
                            </span>
                          ) : null}
                        </div>
                        <div className="mt-1 text-xs text-[var(--text-muted)]">
                          {describeRequestMatch(item.requestMatch) ??
                            "requestMatch не задан"}
                        </div>
                      </div>
                    ))}
                  </div>
                </div>
              ) : null}

              {generation ? (
                <div className="space-y-2 text-sm">
                  <div className="text-[var(--text-muted)]">
                    Статус:{" "}
                    <span className="text-[var(--text-primary)]">
                      {generation.status ?? "—"}
                    </span>
                  </div>
                  {generation.generationId ? (
                    <div className="text-xs text-[var(--text-muted)]">
                      Generation ID: {generation.generationId}
                    </div>
                  ) : null}
                  {generation.resultMockIds?.length ? (
                    <div className="space-y-1">
                      <div className="text-xs text-[var(--text-muted)]">
                        Созданные моки:
                      </div>
                      <div className="flex flex-wrap gap-2 text-xs">
                        {generation.resultMockIds.map((id) => (
                          <button
                            key={id}
                            type="button"
                            className="cursor-pointer rounded-full border border-white/15 px-3 py-1 text-[var(--text-primary)] hover:border-white/30"
                            onClick={() => router.push(`/mocks/${id}`)}
                          >
                            {id.slice(0, 8)}…
                          </button>
                        ))}
                      </div>
                    </div>
                  ) : null}
                  {generation.error ? (
                    <div className="text-xs text-rose-400">
                      {generation.error}
                    </div>
                  ) : null}
                </div>
              ) : null}

              {!preview && !generation ? (
                <p className="text-sm text-[var(--text-muted)]">
                  Здесь появятся результаты запуска.
                </p>
              ) : null}
            </CardContent>
          </Card>
        </div>
      </div>
    </div>
  );
}
