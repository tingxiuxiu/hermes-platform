import { defineConfig } from "vite";
import path from "path";
import react from "@vitejs/plugin-react";
import tailwindcss from "@tailwindcss/vite";

export default defineConfig({
  plugins: [react(), tailwindcss()],
  resolve: {
    alias: {
      "@": path.resolve(__dirname, "./src"),
    },
  },
  // 代理后端8080端口
  server: {
    proxy: {
      "/api": {
        target: "http://localhost:8080",
        changeOrigin: true,
        secure: false,
        rewrite: (path) => path.replace(/^\/api/, ""),
      },
    },
  },
  build: {
    rollupOptions: {
      output: {
        manualChunks(id) {
          if (id.includes("node_modules")) {
            if (id.includes("react-router") || id.includes("@remix-run")) {
              return "router";
            }
            if (id.includes("@tanstack/react-query")) {
              return "query";
            }
            if (id.includes("radix-ui")) {
              return "ui";
            }
            if (id.includes("lucide-react")) {
              return "icons";
            }
            if (
              id.includes("i18next") ||
              id.includes("react-i18next") ||
              id.includes("i18next-http-backend") ||
              id.includes("i18next-browser-languagedetector")
            ) {
              return "i18n";
            }
            if (
              id.includes("zustand") ||
              id.includes("clsx") ||
              id.includes("tailwind-merge") ||
              id.includes("class-variance-authority") ||
              id.includes("zod") ||
              id.includes("react-hook-form") ||
              id.includes("@hookform/resolvers")
            ) {
              return "utils";
            }
            if (
              id.includes("react") ||
              id.includes("react-dom") ||
              id.includes("scheduler")
            ) {
              return "react-vendor";
            }
          }
        },
      },
    },
  },
});
