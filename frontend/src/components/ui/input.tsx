import * as React from "react";
import { cn } from "@/lib/utils";

export interface InputProps
  extends React.InputHTMLAttributes<HTMLInputElement> {}

export const Input = React.forwardRef<HTMLInputElement, InputProps>(
  ({ className, type, ...props }, ref) => {
    return (
      <input
        ref={ref}
        type={type}
        className={cn(
          "flex h-11 w-full rounded-2xl border border-white/40 bg-white/55 px-4 text-sm text-[var(--text-primary)] placeholder:text-[var(--text-muted)] shadow-sm backdrop-blur-md outline-none transition focus-visible:border-sky-400 focus-visible:ring-2 focus-visible:ring-[var(--ring)] dark:border-white/10 dark:bg-white/5",
          className
        )}
        {...props}
      />
    );
  }
);

Input.displayName = "Input";
