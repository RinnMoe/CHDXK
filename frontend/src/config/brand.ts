export const brand = {
  name: "SJTU选课社区",
  feedbackEmail: "course@sjtu.plus",
} as const

export function formatPageTitle(pageTitle?: string) {
  return pageTitle ? `${pageTitle} - ${brand.name}` : brand.name
}
