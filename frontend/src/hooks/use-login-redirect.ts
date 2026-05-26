import { useLocation } from "react-router-dom"
import { buildLoginRedirectPath } from "@/lib/auth-redirect"

export function useLoginRedirectPath() {
  const location = useLocation()
  return buildLoginRedirectPath(
    `${location.pathname}${location.search}${location.hash}`
  )
}
