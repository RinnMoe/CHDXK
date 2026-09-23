export const brand = {
  name: import.meta.env.VITE_SITE_NAME?.trim() || "频道课程榜",
  feedbackEmail: import.meta.env.VITE_SUPPORT_EMAIL?.trim() || undefined,
} as const

export function formatPageTitle(pageTitle?: string) {
  return pageTitle ? `${pageTitle} - ${brand.name}` : brand.name
}
