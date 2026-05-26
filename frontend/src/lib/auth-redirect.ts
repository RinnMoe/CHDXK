const fallbackPath = "/"

export function getSafeRedirectPath(value: string | null | undefined) {
  if (!value || !value.startsWith("/")) return fallbackPath
  if (value.startsWith("//")) return fallbackPath

  try {
    const url = new URL(value, window.location.origin)
    if (url.origin !== window.location.origin) return fallbackPath
    if (url.pathname === "/login") return fallbackPath
    return `${url.pathname}${url.search}${url.hash}`
  } catch {
    return fallbackPath
  }
}

export function buildLoginRedirectPath(path: string) {
  const redirect = getSafeRedirectPath(path)
  const params = new URLSearchParams({ redirect })
  return `/login?${params}`
}
