import { useEffect } from "react"
import { useNavigate } from "react-router-dom"
import { AboutAgreementExcerpt } from "@/components/about/about-content"
import { LoginForm } from "@/components/auth/login-form"
import { PageShell } from "@/components/layout/page-shell"
import { PageTitle } from "@/components/common/page-title"
import { useAuth } from "@/contexts/auth-context"

export function LoginPage() {
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
      <PageTitle>登录</PageTitle>
      <PageShell showAnnouncements={false}>
        <div className="flex justify-center">
          <LoginForm />
        </div>
        <AboutAgreementExcerpt />
      </PageShell>
    </>
  )
}
