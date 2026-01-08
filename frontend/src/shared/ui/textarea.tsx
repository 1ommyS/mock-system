import * as React from "react";
import { cn } from "@/shared/lib/utils";

export interface TextareaProps
  extends React.TextareaHTMLAttributes<HTMLTextAreaElement> {}

export const Textarea = React.forwardRef<HTMLTextAreaElement, TextareaProps>(
  ({ className, ...props }, ref) => {
    return (
      <textarea
        ref={ref}
        className={cn(
          "min-h-[140px] w-full rounded-2xl border border-white/40 bg-white/45 px-4 py-3 text-sm text-[var(--text-primary)] shadow-sm backdrop-blur-md outline-none transition focus-visible:border-sky-400 focus-visible:ring-2 focus-visible:ring-[var(--ring)] dark:border-white/10 dark:bg-white/5",
          className
        )}
        {...props}
      />
    );
  }
);

Textarea.displayName = "Textarea";
