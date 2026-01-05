import path from "path";
import { createRequire } from "module";
import { fileURLToPath } from "url";
import type { StorybookConfig } from "@storybook/react-webpack5";

const config: StorybookConfig = {
  stories: ["../src/**/*.stories.@(js|jsx|ts|tsx|mdx)"],
  addons: [
    "@storybook/addon-links",
    "@storybook/addon-essentials",
    "@storybook/addon-interactions",
  ],
  framework: {
    name: "@storybook/react-webpack5",
    options: {},
  },
  staticDirs: ["../public"],
  typescript: {
    reactDocgen: false,
  },
  webpackFinal: async (baseConfig) => {
    const dirname = path.dirname(fileURLToPath(import.meta.url));
    const require = createRequire(import.meta.url);
    const babelLoader = require.resolve("babel-loader");
    const presetEnv = require.resolve("@babel/preset-env");
    const presetReact = require.resolve("@babel/preset-react");
    const presetTypeScript = require.resolve("@babel/preset-typescript");
    baseConfig.resolve = baseConfig.resolve || {};
    baseConfig.resolve.alias = {
      ...(baseConfig.resolve.alias || {}),
      "@": path.resolve(dirname, "../src"),
    };
    baseConfig.resolve.extensions = [
      ...(baseConfig.resolve.extensions || []),
      ".ts",
      ".tsx",
    ];
    baseConfig.module = baseConfig.module || { rules: [] };
    baseConfig.module.rules = [
      ...(baseConfig.module.rules || []),
      {
        test: /\.(ts|tsx)$/,
        exclude: /node_modules/,
        use: [
          {
            loader: babelLoader,
            options: {
              presets: [
                [presetEnv, { targets: "defaults" }],
                [presetReact, { runtime: "automatic" }],
                presetTypeScript,
              ],
            },
          },
        ],
      },
    ];
    return baseConfig;
  },
};

export default config;
