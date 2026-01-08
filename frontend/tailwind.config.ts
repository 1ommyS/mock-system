import type { Config } from "tailwindcss";

const config: Config = {
  darkMode: "class",
  content: [
    "./src/app/**/*.{ts,tsx}",
    "./src/shared/**/*.{ts,tsx}",
    "./src/features/**/*.{ts,tsx}",
    "./src/entities/**/*.{ts,tsx}",
  ],
  theme: {
    extend: {
      colors: {
        base: {
          50: "#f8fafc",
          100: "#f1f5f9",
          200: "#e2e8f0",
          300: "#cbd5f5",
          500: "#64748b",
          700: "#334155",
          800: "#1f2937",
          900: "#0f172a",
        },
        brand: {
          400: "#2dd4bf",
          500: "#14b8a6",
          600: "#0d9488",
          700: "#0f766e",
        },
      },
      boxShadow: {
        soft: "0 18px 40px -24px rgba(15, 23, 42, 0.35)",
        glow: "0 0 0 1px rgba(20, 184, 166, 0.35), 0 10px 25px -15px rgba(20, 184, 166, 0.6)",
      },
    },
  },
  plugins: [],
};

export default config;
