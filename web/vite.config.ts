import { defineConfig } from "vite";
import react from "@vitejs/plugin-react";

export default defineConfig({
  // Относительные пути к ассетам: одна сборка работает и в корне, и под префиксом (/domashka/).
  base: "./",
  plugins: [react()],
  server: {
    port: 5180,
    strictPort: true,
    proxy: { "/api": "http://127.0.0.1:8080" },
  },
});
