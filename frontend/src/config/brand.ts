import { copy } from "@/lib/copy"

export const brand = {
  name: copy("copy.user.config.brand.743e9916a4b2"),
  feedbackEmail: copy("copy.brand.feedback_email").trim() || undefined,
} as const

export function formatPageTitle(pageTitle?: string) {
  return pageTitle ? `${pageTitle} - ${brand.name}` : brand.name
}
