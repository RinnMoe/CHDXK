import { useEffect } from "react"
import { useNavigate } from "react-router-dom"
import { PasswordResetForm } from "@/components/auth/password-reset-form"
import { PageShell } from "@/components/layout/page-shell"
import { PageTitle } from "@/components/common/page-title"
import { useAuth } from "@/contexts/auth-context"

export function PasswordResetPage() {
  const { user, isLoading } = useAuth()
  const navigate = useNavigate()

  useEffect(() => {
    if (!isLoading && user) {
      navigate("/", { replace: true })
    }
  }, [user, isLoading, navigate])

  if (isLoading || user) return null

  return (
    <>
      <PageTitle>重置密码</PageTitle>
      <PageShell showAnnouncements={false}>
        <div className="flex justify-center">
          <PasswordResetForm />
        </div>
      </PageShell>
    </>
  )
}
