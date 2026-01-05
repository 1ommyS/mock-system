import * as React from "react";
import { cn } from "@/lib/utils";

export function Alert({
  className,
  ...props
}: React.HTMLAttributes<HTMLDivElement>) {
  return (
    <div
      role="alert"
      className={cn(
        "rounded-2xl border border-rose-200/70 bg-rose-50/60 px-4 py-3 text-sm text-rose-900 shadow-sm backdrop-blur-md dark:border-rose-500/30 dark:bg-rose-500/10 dark:text-rose-100",
        className
      )}
      {...props}
    />
  );
}

export function AlertTitle({
  className,
  ...props
}: React.HTMLAttributes<HTMLHeadingElement>) {
  return (
    <h2 className={cn("text-sm font-semibold", className)} {...props} />
  );
}

export function AlertDescription({
  className,
  ...props
}: React.HTMLAttributes<HTMLParagraphElement>) {
  return (
    <div className={cn("mt-1 text-sm opacity-90", className)} {...props} />
  );
}
