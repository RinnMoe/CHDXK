import type { VitePWA } from "vite-plugin-pwa"

export const pwaOptions: Parameters<typeof VitePWA>[0] = {
  registerType: "autoUpdate",
  injectRegister: null,
  includeAssets: ["pwa-icon.svg", "pwa-192x192.png", "pwa-512x512.png"],
  manifest: false,
  workbox: {
    cleanupOutdatedCaches: true,
    globPatterns: ["**/*.{js,css,html,svg,png,ico,woff2}"],
    navigateFallback: "/index.html",
    navigateFallbackDenylist: [/^\/api\//],
  },
}
