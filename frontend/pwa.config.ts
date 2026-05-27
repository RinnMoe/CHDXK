import type { VitePWA } from "vite-plugin-pwa"

const appName = "SJTU选课社区"
const appDescription = "课程检索、教师检索与课程点评社区"

export const pwaOptions: Parameters<typeof VitePWA>[0] = {
  registerType: "autoUpdate",
  injectRegister: null,
  includeAssets: ["pwa-icon.svg", "pwa-192x192.png", "pwa-512x512.png"],
  manifest: {
    id: "/",
    name: appName,
    short_name: "选课社区",
    description: appDescription,
    start_url: "/",
    scope: "/",
    display: "standalone",
    theme_color: "#0f766e",
    background_color: "#f8fafc",
    lang: "zh-CN",
    categories: ["education", "productivity"],
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
    shortcuts: [
      {
        name: "课程",
        short_name: "课程",
        description: "检索课程与查看课程详情",
        url: "/course",
        icons: [
          {
            src: "/pwa-192x192.png",
            sizes: "192x192",
            type: "image/png",
          },
        ],
      },
      {
        name: "点评",
        short_name: "点评",
        description: "浏览课程点评",
        url: "/review",
        icons: [
          {
            src: "/pwa-192x192.png",
            sizes: "192x192",
            type: "image/png",
          },
        ],
      },
      {
        name: "我的点评",
        short_name: "我的点评",
        description: "查看和管理我发布的点评",
        url: "/review/mine",
        icons: [
          {
            src: "/pwa-192x192.png",
            sizes: "192x192",
            type: "image/png",
          },
        ],
      },
    ],
  },
  workbox: {
    cleanupOutdatedCaches: true,
    globPatterns: ["**/*.{js,css,html,svg,png,ico,woff2}"],
    navigateFallback: "/index.html",
    navigateFallbackDenylist: [/^\/api\//],
  },
}
