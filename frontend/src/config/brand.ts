export const brand = {
  name: "JCourse",
} as const

export function formatPageTitle(pageTitle?: string) {
  return pageTitle ? `${pageTitle} - ${brand.name}` : brand.name
}
