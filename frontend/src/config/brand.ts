export const brand = {
  name: "JCourse",
  feedbackEmail: "jcourse@sjtu.edu.cn",
} as const

export function formatPageTitle(pageTitle?: string) {
  return pageTitle ? `${pageTitle} - ${brand.name}` : brand.name
}
