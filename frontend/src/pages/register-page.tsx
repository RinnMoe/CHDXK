import { useEffect } from "react"
import { useNavigate } from "react-router-dom"
import { RegisterForm } from "@/components/auth/register-form"
import { PageShell } from "@/components/layout/page-shell"
import { PageTitle } from "@/components/common/page-title"
import { useAuth } from "@/contexts/auth-context"

export function RegisterPage() {
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
      <PageTitle>注册</PageTitle>
      <PageShell showAnnouncements={false}>
        <div className="flex justify-center">
          <RegisterForm />
        </div>
      </PageShell>
    </>
  )
}
