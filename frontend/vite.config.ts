import path from "path"
import tailwindcss from "@tailwindcss/vite"
import react from "@vitejs/plugin-react"
import { defineConfig } from "vite"

export default defineConfig({
  plugins: [react(), tailwindcss()],
  resolve: {
    alias: {
      "@": path.resolve(__dirname, "./src"),
    },
  },
  build: {
    rolldownOptions: {
      output: {
        assetFileNames: (assetInfo) => {
          if (/\.(woff2?|ttf|eot|otf)$/.test(assetInfo.name ?? ""))
            return "assets/fonts/[name].[ext]"
          return "assets/[name]-[hash][extname]"
        },
      },
    },
  },
  server: {
    proxy: {
      "/api": {
        target: "http://localhost:8080",
        changeOrigin: true,
      },
    },
  },
})
