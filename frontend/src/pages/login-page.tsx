import { useEffect } from "react"
import { useNavigate, useSearchParams } from "react-router-dom"
import { AboutAgreementExcerpt } from "@/components/about/about-content"
import { LoginForm } from "@/components/auth/login-form"
import { PageShell } from "@/components/layout/page-shell"
import { PageTitle } from "@/components/common/page-title"
import { useAuth } from "@/contexts/auth-context"
import { getSafeRedirectPath } from "@/lib/auth-redirect"

export function LoginPage() {
  const { user, isLoading } = useAuth()
  const navigate = useNavigate()
  const [searchParams] = useSearchParams()
  const redirectPath = getSafeRedirectPath(searchParams.get("redirect"))

  useEffect(() => {
    if (!isLoading && user) {
      navigate(redirectPath, { replace: true })
    }
  }, [user, isLoading, navigate, redirectPath])

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
