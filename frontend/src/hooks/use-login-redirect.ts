import { useLocation } from "@tanstack/react-router"
import { buildLoginRedirectPath } from "@/lib/auth-redirect"

export function useLoginRedirectPath() {
  const location = useLocation()
  return buildLoginRedirectPath(
    `${location.pathname}${location.searchStr}${location.hash}`
  )
}
