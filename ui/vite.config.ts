import path from "path"
import react from "@vitejs/plugin-react"
import { defineConfig } from "vite"
 
export default defineConfig({
  plugins: [react()],
  resolve: {
    alias: {
      "@": path.resolve(__dirname, "./src"),
    },
  },
  preview: {
    port: 3001
  },
  server: {
    port: 3001,
    proxy: {
      "/api/v1": {
        target: "http://localhost:9808",
        changeOrigin: true,
        ws: true,
        rewriteWsOrigin: true
      },
      "/apps": {
        target: "http://localhost:9808",
        changeOrigin: true
      }
    }
  },
})