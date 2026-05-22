import { useEffect } from "react"
import { useNavigate } from "react-router-dom"
import { PasswordResetForm } from "@/components/auth/password-reset-form"
import { PageShell } from "@/components/layout/page-shell"
import { useAuth } from "@/contexts/auth-context"

export function PasswordResetPage() {
  const { user, isLoading } = useAuth()
  const navigate = useNavigate()

  useEffect(() => {
    if (!isLoading && user) {
      navigate("/", { replace: true })
    }
  }, [user, isLoading, navigate])

  if (isLoading) return null

  return (
    <>
      <title>重置密码 - JCourse</title>
      <PageShell>
      <div className="flex justify-center">
        <PasswordResetForm />
      </div>
      </PageShell>
    </>
  )
}
