"use client";

import { json } from "@codemirror/lang-json";
import { oneDark } from "@codemirror/theme-one-dark";
import CodeMirror from "@uiw/react-codemirror";
import JSON5 from "json5";
import { useEffect, useRef, useState } from "react";

import { Button } from "@/components/ui/button";
import { cn } from "@/lib/utils";

type CodeEditorProps = {
  label: string;
  value: string;
  onChange: (value: string) => void;
  error?: string;
  hint?: string;
  className?: string;
  autoQuoteKeys?: boolean;
};

export function CodeEditor({
  label,
  value,
  onChange,
  error,
  hint,
  className,
  autoQuoteKeys = true,
}: CodeEditorProps) {
  const [formatError, setFormatError] = useState<string | null>(null);
  const normalizeTimer = useRef<number | null>(null);

  useEffect(() => {
    return () => {
      if (normalizeTimer.current) {
        window.clearTimeout(normalizeTimer.current);
      }
    };
  }, []);

  const normalizeJson = (raw: string) => {
    try {
      const parsed = JSON5.parse(raw);
      return JSON.stringify(parsed, null, 2);
    } catch {
      return null;
    }
  };

  const handleFormat = () => {
    const normalized = normalizeJson(value);
    if (normalized) {
      onChange(normalized);
      setFormatError(null);
      return;
    }
    setFormatError("Не удалось отформатировать JSON.");
  };

  const handleChange = (nextValue: string) => {
    onChange(nextValue);
    setFormatError(null);

    if (!autoQuoteKeys) return;

    if (normalizeTimer.current) {
      window.clearTimeout(normalizeTimer.current);
    }
    normalizeTimer.current = window.setTimeout(() => {
      const normalized = normalizeJson(nextValue);
      if (normalized && normalized !== nextValue) {
        onChange(normalized);
      }
    }, 600);
  };

  return (
    <div className={cn("space-y-2", className)}>
      <div className="flex items-center justify-between">
        <label className="text-sm font-medium">{label}</label>
        <Button variant="ghost" size="sm" type="button" onClick={handleFormat}>
          Форматировать
        </Button>
      </div>
      <div className="rounded-2xl border border-white/40 bg-white/45 shadow-sm backdrop-blur-md dark:border-white/10 dark:bg-white/5">
        <CodeMirror
          value={value}
          height="220px"
          theme={oneDark}
          extensions={[json()]}
          onChange={handleChange}
          basicSetup={{
            highlightActiveLine: false,
            highlightActiveLineGutter: false,
            foldGutter: false,
            lineNumbers: true,
          }}
        />
      </div>
      {hint ? <p className="text-xs text-[var(--text-muted)]">{hint}</p> : null}
      {error ? <p className="text-xs text-rose-400">{error}</p> : null}
      {formatError ? (
        <p className="text-xs text-rose-400">{formatError}</p>
      ) : null}
    </div>
  );
}
