import * as React from "react";
import { cn } from "@/lib/utils";

type ButtonVariant = "primary" | "ghost" | "outline";

type ButtonSize = "default" | "sm" | "icon";

export interface ButtonProps
  extends React.ButtonHTMLAttributes<HTMLButtonElement> {
  variant?: ButtonVariant;
  size?: ButtonSize;
}

const baseStyles =
  "glass-sheen relative inline-flex cursor-pointer items-center justify-center gap-2 overflow-hidden rounded-[999px] text-sm font-semibold transition focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-offset-2 focus-visible:ring-offset-transparent disabled:pointer-events-none disabled:opacity-60 before:pointer-events-none before:absolute before:inset-[1px] before:rounded-[999px] before:border before:border-white/60 before:opacity-80 dark:before:border-white/18";

const variantStyles: Record<ButtonVariant, string> = {
  primary:
    "border border-[color:rgba(14,165,233,0.45)] bg-white/35 text-black shadow-[0_14px_30px_-24px_rgba(14,165,233,0.45)] backdrop-blur-xl hover:bg-[color:rgba(14,165,233,0.14)] focus-visible:ring-[var(--ring)] dark:border-[color:rgba(56,189,248,0.4)] dark:bg-white/6 dark:text-white dark:hover:bg-[color:rgba(56,189,248,0.12)]",
  ghost:
    "border border-white/55 bg-white/55 text-[var(--text-primary)] shadow-[0_10px_24px_-24px_rgba(15,23,42,0.3)] backdrop-blur-xl hover:bg-white/70 dark:border-white/15 dark:bg-white/10 dark:hover:bg-white/15",
  outline:
    "border border-white/50 bg-white/40 text-[var(--text-primary)] shadow-[0_10px_24px_-24px_rgba(15,23,42,0.26)] backdrop-blur-xl hover:bg-white/55 dark:border-white/12 dark:bg-white/8 dark:hover:bg-white/14",
};

const sizeStyles: Record<ButtonSize, string> = {
  default: "h-12 px-7",
  sm: "h-10 px-4 text-xs",
  icon: "h-10 w-10",
};

export const Button = React.forwardRef<HTMLButtonElement, ButtonProps>(
  ({ className, variant = "primary", size = "default", ...props }, ref) => {
    return (
      <button
        ref={ref}
        className={cn(baseStyles, variantStyles[variant], sizeStyles[size], className)}
        {...props}
      />
    );
  }
);

Button.displayName = "Button";
