import path from "path";
import { defineConfig } from "vite";
import react from "@vitejs/plugin-react";
import { lingui } from "@lingui/vite-plugin";
import { sentryVitePlugin } from "@sentry/vite-plugin";

// https://vitejs.dev/config/
export default defineConfig({
  plugins: [
    react({
      babel: {
        plugins: ["@lingui/babel-plugin-lingui-macro"],
        presets: ["jotai-babel/preset"],
      },
    }),
    lingui(),
    sentryVitePlugin({
      org: "konstruktor",
      project: "fatura-react",
      telemetry: false,
      release: {
        name: process.env.GITHUB_SHA || "development",
      },
      sourcemaps: {
        assets: "./dist/**",
        ignore: ["node_modules"],
      },
    }),
  ],
  optimizeDeps: {
    include: ["pdfjs-dist"],
  },
  css: {
    preprocessorOptions: {
      scss: {
        api: "modern-compiler",
      },
    },
  },
  resolve: {
    alias: [
      { find: "src", replacement: "/src" },
      // Wails v2 generated bindings
      { find: "wailsjs", replacement: path.resolve(__dirname, "./wailsjs") },
    ],
  },
  define: {
    global: 'globalThis',
  },
  // Prevent vite from obscuring errors
  clearScreen: false,
  // Wails dev server expects a fixed port. HMR must connect over ws://localhost
  // explicitly — when the WebView loads from wails://wails/, Vite's client would
  // otherwise compute the WebSocket URL as ws://wails:/ (invalid) from import.meta.url.
  server: {
    strictPort: true,
    port: 5173,
    host: "127.0.0.1",
    hmr: {
      protocol: "ws",
      host: "localhost",
      port: 5173,
    },
  },
  envPrefix: ["VITE_"],
  build: {
    target: "es2020",
    minify: "esbuild",
    sourcemap: true,
  },
});
