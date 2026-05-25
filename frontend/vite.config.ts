import path from "path"
import tailwindcss from "@tailwindcss/vite"
import react from "@vitejs/plugin-react"
import { defineConfig } from "vite"
import { VitePWA } from "vite-plugin-pwa"

const appName = "SJTU选课社区"
const appDescription = "课程检索、教师检索与课程点评社区"

const reactPackages = new Set([
  "@tanstack/react-query",
  "react",
  "react-dom",
  "react-router",
  "react-router-dom",
])
const contentPackages = new Set([
  "react-markdown",
  "rehype-sanitize",
  "remark-gfm",
])
const chartPackages = new Set(["recharts"])
const uiPackages = new Set([
  "@remixicon/react",
  "class-variance-authority",
  "clsx",
  "radix-ui",
  "tailwind-merge",
])

function getNodePackageName(id: string) {
  const packagePath = id.split(/[/\\]node_modules[/\\]/).pop()
  if (!packagePath || packagePath === id) return null

  const [scopeOrName, packageName] = packagePath.split(/[/\\]/)
  if (!scopeOrName) return null
  if (scopeOrName.startsWith("@")) return `${scopeOrName}/${packageName}`
  return scopeOrName
}

function isNodePackage(id: string, packages: Set<string>) {
  const packageName = getNodePackageName(id)
  return packageName ? packages.has(packageName) : false
}

export default defineConfig({
  plugins: [
    react(),
    tailwindcss(),
    VitePWA({
      registerType: "autoUpdate",
      injectRegister: null,
      includeAssets: ["pwa-icon.svg", "pwa-192x192.png", "pwa-512x512.png"],
      manifest: {
        name: appName,
        short_name: "选课社区",
        description: appDescription,
        start_url: "/",
        scope: "/",
        display: "standalone",
        theme_color: "#0f766e",
        background_color: "#f8fafc",
        lang: "zh-CN",
        icons: [
          {
            src: "/pwa-192x192.png",
            sizes: "192x192",
            type: "image/png",
          },
          {
            src: "/pwa-512x512.png",
            sizes: "512x512",
            type: "image/png",
          },
          {
            src: "/pwa-512x512.png",
            sizes: "512x512",
            type: "image/png",
            purpose: "maskable",
          },
        ],
      },
      workbox: {
        cleanupOutdatedCaches: true,
        globPatterns: ["**/*.{js,css,html,svg,png,ico,woff2}"],
        navigateFallback: "/index.html",
        runtimeCaching: [
          {
            urlPattern: ({ url }) => url.pathname.startsWith("/api/"),
            handler: "NetworkOnly",
            options: {
              cacheName: "api-network-only",
            },
          },
        ],
      },
    }),
  ],
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
        chunkFileNames: "assets/[name]-[hash].js",
        entryFileNames: "assets/[name]-[hash].js",
        codeSplitting: {
          groups: [
            {
              name: "vendor-react",
              test: (id) => isNodePackage(id, reactPackages),
              priority: 30,
            },
            {
              name: "vendor-markdown",
              test: (id) => isNodePackage(id, contentPackages),
              priority: 20,
            },
            {
              name: "vendor-charts",
              test: (id) => isNodePackage(id, chartPackages),
              priority: 20,
            },
            {
              name: "vendor-ui",
              test: (id) => isNodePackage(id, uiPackages),
              priority: 10,
            },
            {
              name: "vendor",
              test: /node_modules[\\/]/,
              entriesAware: true,
              minSize: 20 * 1024,
            },
          ],
        },
      },
    },
  },
  server: {
    proxy: {
      "/api/": {
        target: "http://localhost:8080",
        changeOrigin: true,
      },
    },
  },
})
