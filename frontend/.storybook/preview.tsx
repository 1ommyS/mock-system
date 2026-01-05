import "@fontsource/space-grotesk/400.css";
import "@fontsource/space-grotesk/500.css";
import "@fontsource/space-grotesk/600.css";
import "@fontsource/jetbrains-mono/400.css";
import "../src/app/globals.css";
import type { Preview } from "@storybook/react";

const preview: Preview = {
  parameters: {
    actions: { argTypesRegex: "^on[A-Z].*" },
    controls: {
      matchers: {
        color: /(background|color)$/i,
        date: /Date$/,
      },
    },
    backgrounds: {
      default: "app",
      values: [
        { name: "app", value: "#0b0f19" },
        { name: "panel", value: "#111827" },
      ],
    },
  },
  decorators: [
    (Story) => {
      if (typeof document !== "undefined") {
        document.documentElement.classList.add("dark");
        document.documentElement.style.setProperty(
          "--font-space",
          "\"Space Grotesk\""
        );
        document.documentElement.style.setProperty(
          "--font-mono",
          "\"JetBrains Mono\""
        );
      }

      return (
        <div className="dark min-h-screen bg-[var(--page-bg)] p-8 text-[var(--text-primary)]">
          <Story />
        </div>
      );
    },
  ],
};

export default preview;
