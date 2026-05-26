import { Navigate, Outlet } from "react-router-dom"
import { useAuth } from "@/contexts/auth-context"
import { useLoginRedirectPath } from "@/hooks/use-login-redirect"

export function RequireAuth() {
  const { user, isLoading } = useAuth()
  const loginRedirectPath = useLoginRedirectPath()

  if (isLoading) {
    return null
  }

  if (!user) {
    return <Navigate to={loginRedirectPath} replace />
  }

  return <Outlet />
}
